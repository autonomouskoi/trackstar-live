package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	trackstar "github.com/autonomouskoi/trackstar/pb"
)

type TrackUpdate struct {
	UserID  string                 `json:"user_id"`
	Session int64                  `json:"started"`
	Update  *trackstar.TrackUpdate `json:"update"`
}

type Subs struct {
	lock sync.RWMutex
	subs map[string][]chan *TrackUpdate
}

func NewSubs() *Subs {
	return &Subs{
		subs: map[string][]chan *TrackUpdate{},
	}
}

func (s *Subs) Add(userID string, sub chan *TrackUpdate) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.subs[userID] = append(s.subs[userID], sub)
}

func (s *Subs) Delete(userID string, sub chan *TrackUpdate) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var newSubs []chan *TrackUpdate
	// this should be something smarter
	for _, havSub := range s.subs[userID] {
		if havSub == sub {
			close(havSub)
		} else {
			newSubs = append(newSubs, havSub)
		}
	}
	s.subs[userID] = newSubs
}

func (s *Subs) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for userID, subs := range s.subs {
		for _, sub := range subs {
			close(sub)
		}
		delete(s.subs, userID)
	}
}

func (s *Subs) Send(tu *TrackUpdate) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, sub := range s.subs[tu.UserID] {
		select {
		case sub <- tu:
		default: // drop it
		}
	}
}

func (srv *Server) sub(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		defaultHTTPError(w, http.StatusInternalServerError)
		srv.logger.Error("accepting web socket",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"error", err.Error(),
		)
		return
	}
	defer c.CloseNow()

	in := make(chan *TrackUpdate)
	srv.subs.Add(userID, in)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		srv.subs.Delete(userID, in)
	}()

	for tu := range in {
		srv.logger.Debug("sending track update to client",
			"remote", r.RemoteAddr,
			"user_id", userID,
			"started", tu.Session,
			"idx", tu.Update.Index,
		)
		if err := wsjson.Write(r.Context(), c, tu); err != nil {
			cancel()
		}
	}

	c.Close(websocket.StatusNormalClosure, "")
}
