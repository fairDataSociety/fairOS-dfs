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

type LoginStatus struct {
	LoggedIn bool `json:"loggedin"`
}

// IsUserLoggedInHandler is the api handler to check if a user is logged in or not
// it takes one argument
// -user_name: the user name to check if it logged in or not
func (h *Handler) IsUserLoggedInHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["user_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("user isloggedin: \"user_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user isloggedin: \"user_name\" argument missing"})
		return
	}

	user := keys[0]
	if user == "" {
		h.logger.Errorf("user isloggedin: \"user\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "user isloggedin: \"user\" argument missing"})
		return
	}

	yes := h.dfsAPI.IsUserLoggedIn(user)

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &LoginStatus{LoggedIn: yes})
}
