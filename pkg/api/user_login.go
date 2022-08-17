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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

// UserLoginV2Handler is the api handler to login a user
// it takes two arguments
// - user_name: the name of the user to login
// - password: the password of the user
func (h *Handler) UserLoginV2Handler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user login: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "user login: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user login: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "user login: could not decode arguments"})
		return
	}

	user := userReq.UserName
	password := userReq.Password
	if user == "" {
		h.logger.Errorf("user login: \"user_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user login: \"user_name\" argument missing"})
		return
	}
	if password == "" {
		h.logger.Errorf("user login: \"password\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user login: \"password\" argument missing"})
		return
	}

	// login user
	ui, nameHash, publicKey, err := h.dfsAPI.LoginUserV2(user, password, "")
	if err != nil {
		if err == u.ErrUserAlreadyLoggedIn ||
			err == u.ErrInvalidUserName ||
			err == u.ErrInvalidPassword {
			h.logger.Errorf("user login: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "user login: " + err.Error()})
			return
		}
		h.logger.Errorf("user login: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user login: " + err.Error()})
		return
	}
	err = cookie.SetSession(ui.GetSessionId(), w, h.cookieDomain)
	if err != nil {
		h.logger.Errorf("user login: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user login: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &UserSignupResponse{
		Address:   ui.GetAccount().GetUserAccountInfo().GetAddress().Hex(),
		NameHash:  "0x" + nameHash,
		PublicKey: publicKey,
		Message:   "user logged-in successfully",
	})
}
