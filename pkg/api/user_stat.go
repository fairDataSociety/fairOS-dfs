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
	"resenje.org/jsonhttp"
)

// UserStatHandler godoc
//
//	@Summary      User stat
//	@Description  show user stats
//	@Tags         v1
//	@Accept       json
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  user.Stat
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/user/stat [get]
func (h *Handler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user stat: invalid cookie: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("user stat: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "user stat: \"cookie-id\" parameter missing in cookie"})
		return
	}

	userStat, err := h.dfsAPI.GetUserStat(sessionId)
	if err != nil {
		h.logger.Errorf("user stat: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user stat: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, userStat)
}
