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

// LoginStatus is the json response sent to check if user is logged-in
type LoginStatus struct {
	LoggedIn bool `json:"loggedin"`
}

// IsUserLoggedInHandler godoc
//
//	@Summary      Is user logged-in
//	@Description  Check if the given user is logged-in
//	@Tags         user
//	@Accept       json
//	@Produce      json
//	@Param	      userName query string true "username"
//	@Success      200  {object}  LoginStatus
//	@Failure      400  {object}  response
//	@Router       /v1/user/isloggedin [get]
func (h *Handler) IsUserLoggedInHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["userName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("user isloggedin: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user isloggedin: \"userName\" argument missing"})
		return
	}

	user := keys[0]
	if user == "" {
		h.logger.Errorf("user isloggedin: \"userName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user isloggedin: \"userName\" argument missing"})
		return
	}

	yes := h.dfsAPI.IsUserLoggedIn(user)

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &LoginStatus{LoggedIn: yes})
}
