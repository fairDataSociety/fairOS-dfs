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
	"github.com/fairdatasociety/fairOS-dfs/pkg/fnm/eth"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

var (
	jsonContentType = "application/json"
)

type UserSignupResponse struct {
	Address   string `json:"address"`
	Mnemonic  string `json:"mnemonic,omitempty"`
	NameHash  string `json:"nameHash,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Message   string `json:"message,omitempty"`
}

// UserSignupHandler is the api handler to creata new user
// it takes two mandatory arguments and one optional argument
// - user_name: the name of the user to create
// - password: the password of the user
// * mnemonic: a 12 word mnemonic to use to create the hd wallet of the user
func (h *Handler) UserSignupHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("user signup: invalid request body type")
		jsonhttp.BadRequest(w, "user signup: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var userReq common.UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		h.logger.Errorf("user signup: could not decode arguments")
		jsonhttp.BadRequest(w, "user signup: could not decode arguments")
		return
	}

	user := userReq.UserName
	password := userReq.Password
	mnemonic := userReq.Mnemonic
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
	address, createdMnemonic, nameHash, publicKey, ui, err := h.dfsAPI.CreateUser(user, password, mnemonic, "")
	if err != nil {
		if err == u.ErrUserAlreadyPresent {
			h.logger.Errorf("user signup: %v", err)
			jsonhttp.BadRequest(w, "user signup: "+err.Error())
			return
		}
		if err == eth.ErrInsufficientBalance {
			h.logger.Errorf("user signup: %v", err)
			jsonhttp.PaymentRequired(w, &UserSignupResponse{
				Address:  address,
				Mnemonic: createdMnemonic,
				Message:  eth.ErrInsufficientBalance.Error(),
			})
			return
		}
		h.logger.Errorf("user signup: %v", err)
		jsonhttp.InternalServerError(w, "user signup: "+err.Error())
		return
	}
	err = cookie.SetSession(ui.GetSessionId(), w, h.cookieDomain)
	if err != nil {
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
		Address:   address,
		Mnemonic:  mnemonic,
		NameHash:  "0x" + nameHash,
		PublicKey: publicKey,
		Message:   "user signed-up successfully",
	})
}
