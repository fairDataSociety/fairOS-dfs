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

func (h *Handler) DocOpenHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc open: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc open: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc open: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocOpen(sessionId, name)
	if err != nil {
		h.logger.Errorf("doc open: %v", err)
		jsonhttp.InternalServerError(w, "doc open: "+err.Error())
		return
	}
	jsonhttp.OK(w, "document store opened")
}
