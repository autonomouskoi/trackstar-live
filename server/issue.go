package server

import "net/http"

func (srv *Server) handleIssue(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get(headerToken)
	if auth == "" {
		http.Error(w, "required header: "+headerToken, http.StatusBadRequest)
		return
	}
	userID := r.PostFormValue("user_id")
	if userID == "" {
		http.Error(w, "required param: user_id", http.StatusBadRequest)
		return
	}
	t, err := srv.auth.mintToken(userID, auth)
	if err != nil {
		defaultHTTPError(w, http.StatusForbidden)
		srv.logger.Warn("minting token",
			"error", err.Error(),
			"remote", r.RemoteAddr,
			"user_id", userID,
		)
		return
	}
	srv.logger.Info("issued token",
		"user_id", userID,
		"remote", r.RemoteAddr,
		"issued_at", t.IssuedAt,
		"expires", t.ExpiresAt,
	)
	srv.sendProto(w, t)
}
