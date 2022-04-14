package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

// UserMigrateHandler is the api handler to migrate local user credential to secondary location in swarm
// it takes only one argument
// - password: the password of the user
func (h *Handler) UserMigrateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user migrate: invalid request body type")
		jsonhttp.BadRequest(w, "user migrate: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user migrate: could not decode arguments")
		jsonhttp.BadRequest(w, "user migrate: could not decode arguments")
		return
	}

	password := userReq.Password
	if password == "" {
		h.logger.Errorf("user migrate: \"password\" argument missing")
		jsonhttp.BadRequest(w, "user migrate: \"password\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user migrate: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user migrate: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "user migrate: \"cookie-id\" parameter missing in cookie")
		return
	}

	// migrate user
	err = h.dfsAPI.MigrateUser(password, sessionId)
	if err != nil {
		if err == u.ErrInvalidUserName ||
			err == u.ErrInvalidPassword ||
			err == u.ErrUserNotLoggedIn {
			h.logger.Errorf("user migrate: %v", err)
			jsonhttp.BadRequest(w, "user migrate: "+err.Error())
			return
		}
		h.logger.Errorf("user migrate: %v", err)
		jsonhttp.InternalServerError(w, "user migrate: "+err.Error())
		return
	}

	// clear cookie
	cookie.ClearSession(w)

	jsonhttp.OK(w, "user migrated successfully")
}
