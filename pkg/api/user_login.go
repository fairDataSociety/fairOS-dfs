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
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

// UserLoginHandler is the api handler to login a user
// it takes two arguments
// - user_name: the name of the user to login
// - password: the password of the user
func (h *Handler) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user login: invalid request body type")
		jsonhttp.BadRequest(w, "user login: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user login: could not decode arguments")
		jsonhttp.BadRequest(w, "user login: could not decode arguments")
		return
	}

	user := userReq.UserName
	password := userReq.Password
	if user == "" {
		h.logger.Errorf("user login: \"user_name\" argument missing")
		jsonhttp.BadRequest(w, "user login: \"user_name\" argument missing")
		return
	}
	if password == "" {
		h.logger.Errorf("user login: \"password\" argument missing")
		jsonhttp.BadRequest(w, "user login: \"password\" argument missing")
		return
	}

	// login user
	err = h.dfsAPI.LoginUser(user, password, w, "")
	if err != nil {
		if err == u.ErrUserAlreadyLoggedIn ||
			err == u.ErrInvalidUserName ||
			err == u.ErrInvalidPassword {
			h.logger.Errorf("user login: %v", err)
			jsonhttp.BadRequest(w, "user login: "+err.Error())
			return
		}
		h.logger.Errorf("user login: %v", err)
		jsonhttp.InternalServerError(w, "user login: "+err.Error())
		return
	}
	jsonhttp.OK(w, "user logged-in successfully")
}
