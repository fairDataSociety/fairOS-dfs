/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

// UserDeleteHandler is the api handler to delete a user
// it takes only one argument
// - password: the password of the user
func (h *Handler) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user signup: invalid request body type")
		jsonhttp.BadRequest(w, "user signup: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user signup: could not decode arguments")
		jsonhttp.BadRequest(w, "user signup: could not decode arguments")
		return
	}

	password := userReq.Password
	if password == "" {
		h.logger.Errorf("user delete: \"password\" argument missing")
		jsonhttp.BadRequest(w, "user delete: \"password\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user delete: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user delete: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "user delete: \"cookie-id\" parameter missing in cookie")
		return
	}

	// delete user
	err = h.dfsAPI.DeleteUser(password, sessionId)
	if err != nil {
		if err == u.ErrInvalidUserName ||
			err == u.ErrInvalidPassword ||
			err == u.ErrUserNotLoggedIn {
			h.logger.Errorf("user delete: %v", err)
			jsonhttp.BadRequest(w, "user delete: "+err.Error())
			return
		}
		h.logger.Errorf("user delete: %v", err)
		jsonhttp.InternalServerError(w, "user delete: "+err.Error())
		return
	}

	// clear cookie
	cookie.ClearSession(w)

	jsonhttp.OK(w, "user deleted successfully")
}
