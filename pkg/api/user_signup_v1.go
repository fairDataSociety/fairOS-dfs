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

// UserSignupHandler is the api handler to create new user
// it takes two mandatory arguments and one optional argument
// - user_name: the name of the user to create
// - password: the password of the user
// * mnemonic: a 12 word mnemonic to use to create the hd wallet of the user
func (h *Handler) UserSignupHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user signup: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "user signup: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user signup: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "user signup: could not decode arguments"})
		return
	}

	user := userReq.UserName
	password := userReq.Password
	if user == "" {
		h.logger.Errorf("user signup: \"user\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user signup: \"user\" argument missing"})
		return
	}
	if password == "" {
		h.logger.Errorf("user signup: \"password\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user signup: \"password\" argument missing"})
		return
	}
	jsonhttp.BadRequest(w, &response{Message: "user signup: deprecated"})
}
