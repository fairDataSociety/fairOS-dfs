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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type UserExportResponse struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (h *Handler) ExportUserHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user export: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user export: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "save contact: \"cookie-id\" parameter missing in cookie")
		return
	}

	name, address, err := h.dfsAPI.ExportUser(sessionId)
	if err != nil {
		h.logger.Errorf("user export: %v", err)
		jsonhttp.InternalServerError(w, "user export: "+err.Error())
		return
	}

	jsonhttp.OK(w, &UserExportResponse{
		Name:    name,
		Address: address,
	})
}
