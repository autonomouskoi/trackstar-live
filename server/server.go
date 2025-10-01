package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	headerToken         = "x-extension-jwt"
	headerContentLength = "Content-Length"
	headerContentType   = "Content-Type"

	contentTypeProto = "application/protobuf"
)

type Server struct {
	handler http.Handler
	auth    *jwtAuth
	logger  *slog.Logger
	store   Store
	subs    *Subs
}

func New(cfg *ServerConfig, logger *slog.Logger, store Store) (*Server, error) {
	mux := http.NewServeMux()

	srv := &Server{
		logger: logger,
		auth:   newJWTAuth(cfg),
		store:  store,
		subs:   NewSubs(),
	}

	mux.HandleFunc("POST /_issue", srv.handleIssue)
	mux.HandleFunc("POST /_trackUpdate/{userID}/{started}", srv.addTrackUpdate)
	mux.HandleFunc("GET /_trackUpdate/{userID}", srv.sessionsList)
	mux.HandleFunc("GET /_trackUpdate/{userID}/{started}", srv.sessionGet)
	mux.HandleFunc("GET /_sub/{userID}", srv.sub)

	indexPath := filepath.Join(cfg.HTMLPath, "index.html")
	mux.HandleFunc("GET /u/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexPath)
	})
	mux.Handle("GET /", http.FileServer(http.Dir(cfg.HTMLPath)))

	// mux wrappers here

	srv.handler = mux

	return srv, nil
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.handler.ServeHTTP(w, r)
}

func (srv *Server) sendProto(w http.ResponseWriter, msg protoreflect.ProtoMessage) {
	b, err := proto.Marshal(msg)
	if err != nil {
		name := msg.ProtoReflect().Descriptor().FullName()
		srv.logger.Error("marshalling", "type", name, "error", err.Error())
		defaultHTTPError(w, http.StatusInternalServerError)
		return
	}
	w.Header().Set(headerContentType, contentTypeProto)
	w.Header().Set(headerContentLength, strconv.Itoa(len(b)))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, bytes.NewReader(b))
}

func (srv *Server) sendJSON(w http.ResponseWriter, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		srv.logger.Error("marshalling JSON", "type", fmt.Sprintf("%T", v))
	}
	w.Header().Set(headerContentType, "application/json")
	w.Header().Set(headerContentLength, strconv.Itoa(len(b)))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, bytes.NewReader(b))
}

func defaultHTTPError(w http.ResponseWriter, statusCode int) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}
