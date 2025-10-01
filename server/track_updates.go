package server

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	trackstar "github.com/autonomouskoi/trackstar/pb"
)

func (s *Server) addTrackUpdate(w http.ResponseWriter, r *http.Request) {
	userID, err := s.auth.parse(r.Header.Get(headerToken))
	if err != nil {
		defaultHTTPError(w, http.StatusForbidden)
		s.logger.Warn("bad token for track update",
			"remote", r.RemoteAddr,
			"error", err.Error(),
		)
		return
	}
	if pathUserID := r.PathValue("userID"); pathUserID != userID {
		http.Error(w, "token mismatch", http.StatusForbidden)
		s.logger.Warn("token mismatch",
			"path_user_id", pathUserID,
			"token_user_id", userID,
		)
		return
	}

	startedStr := r.PathValue("started")
	if startedStr == "" {
		http.Error(w, "required: started", http.StatusBadRequest)
		return
	}
	started, err := strconv.ParseInt(startedStr, 10, 64)
	if err != nil {
		http.Error(w, "parsing started: "+err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get(headerContentType) != contentTypeProto {
		defaultHTTPError(w, http.StatusNotAcceptable)
		return
	}
	contentLen := 0
	if contentLenStr := r.Header.Get(headerContentLength); contentLenStr == "" {
		defaultHTTPError(w, http.StatusLengthRequired)
		return
	} else if contentLen, err = strconv.Atoi(contentLenStr); err != nil || contentLen < 1 {
		http.Error(w, "bad content-length", http.StatusBadRequest)
		return
	}
	if contentLen > 4096 {
		defaultHTTPError(w, http.StatusRequestEntityTooLarge)
		return
	}

	tuBytes := make([]byte, contentLen)
	n, err := r.Body.Read(tuBytes)
	if err != nil && !errors.Is(err, io.EOF) {
		defaultHTTPError(w, http.StatusInternalServerError)
		s.logger.Error("reading track update body",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"started", started,
			"error", err.Error(),
		)
		return
	}
	if n != contentLen {
		http.Error(w, "content-length does not body length", http.StatusBadRequest)
		return
	}

	var tu trackstar.TrackUpdate
	if err := proto.Unmarshal(tuBytes, &tu); err != nil {
		http.Error(w, "bad track update", http.StatusBadRequest)
		s.logger.Warn("parsing track update",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"started", started,
			"error", err.Error(),
		)
		return
	}

	if err := s.store.AddTrackUpdate(r.Context(), userID, started, &tu); err != nil {
		defaultHTTPError(w, http.StatusInsufficientStorage)
		s.logger.Error("adding track update",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"started", started,
			"error", err.Error(),
		)
		return
	}
	s.subs.Send(&TrackUpdate{
		UserID:  userID,
		Session: started,
		Update:  &tu,
	})
	s.logger.Debug("submitting track update",
		"remote", r.RemoteAddr,
		"user_id", userID,
		"started", started,
		"idx", tu.Index,
	)

	w.WriteHeader(http.StatusOK)
}

func (srv *Server) sessionsList(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	sessions, err := srv.store.SessionsList(r.Context(), userID)
	if err != nil {
		defaultHTTPError(w, http.StatusInternalServerError)
		srv.logger.Error("listing sessions",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"error", err.Error(),
		)
		return
	}
	srv.sendJSON(w, map[string][]int64{"sessions": sessions})
}

func (srv *Server) sessionGet(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	startedStr := r.PathValue("started")
	started, err := strconv.ParseInt(startedStr, 10, 64)
	if err != nil {
		defaultHTTPError(w, http.StatusBadRequest)
		return
	}
	updates, err := srv.store.SessionGet(r.Context(), userID, started)
	if err != nil {
		defaultHTTPError(w, http.StatusInternalServerError)
		srv.logger.Error("listing sessions",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"started", startedStr,
			"error", err.Error(),
		)
		return
	}
	switch r.FormValue("download") {
	case "csv":
		srv.sendCSV(w, userID, started, updates)
		return
	}
	updatesJSON := struct {
		Updates []json.RawMessage `json:"updates"`
	}{}
	for _, update := range updates {
		b, err := protojson.Marshal(update)
		if err != nil {
			defaultHTTPError(w, http.StatusInternalServerError)
			srv.logger.Error("marshalling update", "error", err.Error())
			return
		}
		updatesJSON.Updates = append(updatesJSON.Updates, b)
	}
	srv.sendJSON(w, updatesJSON)
}

func (srv *Server) sendCSV(w http.ResponseWriter, userID string, started int64, tracks []*trackstar.TrackUpdate) {
	w.Header().Set(headerContentType, "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.csv"`, userID, time.UnixMilli(started).Format(time.DateOnly)))
	csvW := csv.NewWriter(w)
	csvW.Write([]string{
		"index", "when", "deck ID", "artist", "title",
	})
	for _, tu := range tracks {
		csvW.Write([]string{
			strconv.Itoa(int(tu.Index)),
			time.Unix(tu.GetWhen(), 0).Format(time.RFC3339),
			tu.GetDeckId(),
			tu.GetTrack().Artist,
			tu.GetTrack().Title,
		})
	}
	csvW.Flush()
}
