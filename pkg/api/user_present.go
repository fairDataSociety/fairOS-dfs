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
)

type PresentResponse struct {
	Present bool `json:"present"`
}

// UserPresentHandler godoc
//
//	@Tags         user
//	@Deprecated
//	@Router       /v1/user/present [get]
func (h *Handler) UserPresentHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["userName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("user present: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user present: \"userName\" argument missing"})
		return
	}

	user := keys[0]
	if user == "" {
		h.logger.Errorf("user present: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user present: \"userName\" argument missing"})
		return
	}
	jsonhttp.BadRequest(w, &response{Message: "user present: deprecated"})
}

// UserPresentV2Handler godoc
//
//	@Summary      Check if user is present
//	@Description  checks if the new user is present in the new ENS based authentication
//	@Tags         user
//	@Produce      json
//	@Param	      userName query string true "username"
//	@Success      200  {object}  PresentResponse
//	@Failure      400  {object}  response
//	@Router       /v2/user/present [get]
func (h *Handler) UserPresentV2Handler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["userName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("user present: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user present: \"userName\" argument missing"})
		return
	}

	user := keys[0]
	if user == "" {
		h.logger.Errorf("user present: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user present: \"userName\" argument missing"})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	// check if user is present
	if h.dfsAPI.IsUserNameAvailableV2(user) {
		jsonhttp.OK(w, &PresentResponse{
			Present: true,
		})
	} else {
		jsonhttp.OK(w, &PresentResponse{
			Present: false,
		})
	}
}
