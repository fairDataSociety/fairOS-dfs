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

	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

type UserSignupResponse struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

func (h *Handler) UserSignupHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	password := r.FormValue("password")
	mnemonic := r.FormValue("mnemonic") // this is optional
	if user == "" {
		h.logger.Errorf("user signup: \"user\" argument missing")
		jsonhttp.BadRequest(w, "user signup: \"user\" argument missing")
		return
	}
	if password == "" {
		h.logger.Errorf("user signup: \"password\" argument missing")
		jsonhttp.BadRequest(w, "user signup: \"password\" argument missing")
		return
	}

	// create user
	address, createdMnemonic, err := h.dfsAPI.CreateUser(user, password, mnemonic, w, "")
	if err != nil {
		if err == u.ErrUserAlreadyPresent {
			h.logger.Errorf("user signup: %v", err)
			jsonhttp.BadRequest(w, "user signup: "+err.Error())
			return
		}
		h.logger.Errorf("user signup: %v", err)
		jsonhttp.InternalServerError(w, "user signup: "+err.Error())
		return
	}

	if mnemonic == "" {
		mnemonic = createdMnemonic
	} else {
		mnemonic = ""
	}

	// send the response
	w.Header().Set("Content-Type", " application/json")
	jsonhttp.Created(w, &UserSignupResponse{
		Address:  address,
		Mnemonic: mnemonic,
	})
}
