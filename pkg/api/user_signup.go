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
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	u "github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"resenje.org/jsonhttp"
)

var (
	jsonContentType = "application/json"
)

// UserSignupResponse
type UserSignupResponse struct {
	Address   string `json:"address"`
	NameHash  string `json:"nameHash,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	Message   string `json:"message,omitempty"`
	Mnemonic  string `json:"mnemonic,omitempty"`
}

// UserSignupV2Handler godoc
//
//	@Summary      Register New User
//	@Description  registers new user with the new ENS based authentication
//	@Tags         user
//	@Accept       json
//	@Produce      json
//	@Param	      user_request body common.UserSignupRequest true "username"
//	@Success      201  {object}  UserSignupResponse
//	@Failure      400  {object}  response
//	@Failure      402  {object}  UserSignupResponse
//	@Failure      500  {object}  response
//	@Router       /v2/user/signup [post]
func (h *Handler) UserSignupV2Handler(w http.ResponseWriter, r *http.Request) {
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

	user := userReq.UserName
	password := userReq.Password
	mnemonic := userReq.Mnemonic
	if user == "" {
		h.logger.Errorf("user signup: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user signup: \"userName\" argument missing"})
		return
	}
	if password == "" {
		h.logger.Errorf("user signup: \"password\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user signup: \"password\" argument missing"})
		return
	}

	// create user
	address, createdMnemonic, nameHash, publicKey, ui, err := h.dfsAPI.CreateUserV2(user, password, mnemonic, "")
	if err != nil {
		if err == u.ErrUserAlreadyPresent {
			h.logger.Errorf("user signup: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "user signup: " + err.Error()})
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
		jsonhttp.InternalServerError(w, &response{Message: "user signup: " + err.Error()})
		return
	}
	err = cookie.SetSession(ui.GetSessionId(), w, h.cookieDomain)
	if err != nil {
		h.logger.Errorf("user signup: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "user signup: " + err.Error()})
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
