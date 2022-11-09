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
	"resenje.org/jsonhttp"
)

// UserLoginHandler godoc
//
//	@Deprecated
func (h *Handler) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user login: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "user login: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserSignupRequest
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
	jsonhttp.BadRequest(w, &response{Message: "user login: deprecated"})
}
