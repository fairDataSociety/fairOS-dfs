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

func (h *Handler) CollectionCreateHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("collection create: \"name\" argument missing")
		jsonhttp.BadRequest(w, "collection create: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("collection create: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("collection create: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "collection create: \"cookie-id\" parameter missing in cookie")
		return
	}

	index := "key"
	err = h.dfsAPI.CreateCollection(sessionId, name, index)
	if err != nil {
		h.logger.Errorf("collection create: %v", err)
		jsonhttp.InternalServerError(w, "collection create: "+err.Error())
		return
	}
	jsonhttp.OK(w, "kv store created")
}
