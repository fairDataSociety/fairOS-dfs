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

type UserPresentResponse struct {
	Present bool `json:"present"`
}

func (h *Handler) UserPresentHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	if user == "" {
		h.logger.Errorf("user present: \"user\" argument missing")
		jsonhttp.BadRequest(w, "user present: \"user\" argument missing")
		return
	}

	w.Header().Set("Content-Type", " application/json")
	// check if user is present
	if h.dfsAPI.IsUserNameAvailable(user) {
		jsonhttp.OK(w, &UserPresentResponse{
			Present: true,
		})
	} else {
		jsonhttp.OK(w, &UserPresentResponse{
			Present: false,
		})
	}
}
