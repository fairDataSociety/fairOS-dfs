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
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"resenje.org/jsonhttp"
)

func (h *Handler) DocIndexJsonHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc indexjson: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc indexjson: \"name\" argument missing")
		return
	}

	podFile := r.FormValue("file")
	if podFile == "" {
		h.logger.Errorf("doc indexjson: \"file\" argument missing")
		jsonhttp.BadRequest(w, "doc indexjson: \"file\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc indexjson: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc indexjson: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc indexjson: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocIndexJson(sessionId, name, podFile)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrFileNotPresent {
			h.logger.Errorf("doc indexjson: %v", err)
			jsonhttp.BadRequest(w, "doc indexjson: "+err.Error())
			return
		}
		h.logger.Errorf("doc indexjson: %v", err)
		jsonhttp.InternalServerError(w, "doc indexjson: "+err.Error())
		return
	}
	jsonhttp.OK(w, "indexing started")
}
