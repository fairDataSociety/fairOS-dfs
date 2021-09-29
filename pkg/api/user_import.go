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

// ImportUserHandler is the api handler to import an exported user in to a new machine
// it takes four arguments, to mandatory and one of the other two is optional
// - user_name: the name of the user to import
// - password: the password of the user
//  one of the below is optional
// - address: address of the user
// - mnemonic: 12 word mnemonic
func (h *Handler) ImportUserHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user import: invalid request body type")
		jsonhttp.BadRequest(w, "user import: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user import: could not decode arguments")
		jsonhttp.BadRequest(w, "user import: could not decode arguments")
		return
	}

	user := userReq.UserName
	address := userReq.Address
	mnemonic := userReq.Mnemonic
	password := userReq.Password
	if user == "" {
		h.logger.Errorf("user import: \"user\" argument missing")
		jsonhttp.BadRequest(w, "user import: \"user\" argument missing")
		return
	}

	if password == "" {
		h.logger.Errorf("user import: \"password\" argument missing")
		jsonhttp.BadRequest(w, "user import: \"password\" argument missing")
		return
	}

	if address == "" && mnemonic == "" {
		h.logger.Errorf("user import: either \"address\" or \"mnemonic\" is mandatory")
		jsonhttp.BadRequest(w, "user import: either \"address\" or \"mnemonic\" is mandatory")
		return
	}

	if mnemonic != "" && address == "" {
		address, _, err := h.dfsAPI.CreateUser(user, password, mnemonic, w, "")
		if err != nil {
			if err == u.ErrUserAlreadyPresent {
				h.logger.Errorf("user import: %v", err)
				jsonhttp.BadRequest(w, "user import: "+err.Error())
				return
			}
			h.logger.Errorf("user import: %v", err)
			jsonhttp.InternalServerError(w, "user import: "+err.Error())
			return
		}

		// send the response
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.Created(w, &UserSignupResponse{
			Address: address,
		})
	}

	if address != "" {
		err := h.dfsAPI.ImportUserUsingAddress(user, password, address, w, "")
		if err != nil {
			h.logger.Errorf("user import: %v", err)
			jsonhttp.InternalServerError(w, "user import: "+err.Error())
			return
		}

		w.Header().Set("Content-Type", " application/json")
		jsonhttp.Created(w, &UserSignupResponse{
			Address: address,
		})
	}

}
