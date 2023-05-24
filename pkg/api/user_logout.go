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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

// UserLogoutHandler godoc
//
//	@Summary      Logout
//	@Description  logs-out user
//	@ID 		  user-logout
//	@Tags         user
//	@Accept       json
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/user/logout [post]
func (h *Handler) UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user logout: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user logout: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "user logout: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// logout user
	err = h.dfsAPI.LogoutUser(sessionId)
	if err != nil {
		if err == u.ErrUserNotLoggedIn || err == u.ErrInvalidUserName {
			h.logger.Errorf("user logout: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "user logout: " + err.Error()})
			return
		}
		h.logger.Errorf("user logout: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user logout: " + err.Error()})
		return
	}

	cookie.ClearSession(w)
	jsonhttp.OK(w, &response{Message: "user logged out successfully"})
}
