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

// UserDeleteHandler godoc
//
//	@Tags         user
//	@Deprecated
//	@Router       /v1/user/delete [post]
func (h *Handler) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user signup: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "user signup: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserSignupRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user signup: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "user signup: could not decode arguments"})
		return
	}

	password := userReq.Password
	if password == "" {
		h.logger.Errorf("user delete: \"password\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user delete: \"password\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user delete: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user delete: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "user delete: \"cookie-id\" parameter missing in cookie"})
		return
	}
	jsonhttp.BadRequest(w, &response{Message: "user delete: deprecated"})
}

// UserDeleteRequest is the json request to delete user
type UserDeleteRequest struct {
	Password string `json:"password,omitempty"`
}

// UserDeleteV2Handler godoc
//
//	@Summary      Delete user for ENS based authentication
//	@Description  deletes user info from swarm
//
// @ID  		user-delete-v2
//
//	@Tags         user
//	@Produce      json
//	@Param	      UserDeleteRequest body UserDeleteRequest true "user delete request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v2/user/delete [delete]
func (h *Handler) UserDeleteV2Handler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user delete: invalid request body type")
		jsonhttp.BadRequest(w, "user delete: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq UserDeleteRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user delete: could not decode arguments")
		jsonhttp.BadRequest(w, "user delete: could not decode arguments")
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
	err = h.dfsAPI.DeleteUserV2(password, sessionId)
	if err != nil {
		if err == u.ErrInvalidUserName ||
			err == u.ErrInvalidPassword ||
			err == u.ErrUserNotLoggedIn {
			h.logger.Errorf("user delete: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "user delete: " + err.Error()})
			return
		}
		h.logger.Errorf("user delete: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user delete: " + err.Error()})
		return
	}

	// clear cookie
	cookie.ClearSession(w)

	jsonhttp.OK(w, &response{Message: "user deleted successfully"})
}
