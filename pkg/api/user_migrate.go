package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// UserMigrateHandler is the api handler to migrate local user credential to secondary location in swarm
// it takes only one argument
// - password: the password of the user
func (h *Handler) UserMigrateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user migrate: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "user migrate: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserSignupRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user migrate: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "user migrate: could not decode arguments"})
		return
	}

	password := userReq.Password
	if password == "" {
		h.logger.Errorf("user migrate: \"password\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user migrate: \"password\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user migrate: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user migrate: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "user migrate: \"cookie-id\" parameter missing in cookie"})
		return
	}

	jsonhttp.BadRequest(w, &response{Message: "user migrate: deprecated"})
}
