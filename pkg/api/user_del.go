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
	"net/http"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func (h *Handler) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
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
	err = h.dfsAPI.DeleteUser(password, sessionId, w)
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
	jsonhttp.OK(w, "user deleted successfully")
}
