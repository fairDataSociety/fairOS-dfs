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
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"net/http"
	"resenje.org/jsonhttp"
	"strconv"
)

type DocResponse struct {
	Docs []Document `json:"documents"`
}

type Document struct {
	Doc []byte `json:"doc"`
}

func (h *Handler) DocGetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc get: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"name\" argument missing")
		return
	}

	expr := r.FormValue("expr")
	if name == "" {
		h.logger.Errorf("doc get: \"expr\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"expr\" argument missing")
		return
	}

	noOfRecords := r.FormValue("limit")
	if noOfRecords == "" {
		noOfRecords = "10" // default is 10
	}
	limit, err := strconv.Atoi(noOfRecords)
	if err != nil {
		h.logger.Errorf("doc get: \"limit\" invalid argument")
		jsonhttp.BadRequest(w, "doc get: \"limit\" invalid argument")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc get: \"cookie-id\" parameter missing in cookie")
		return
	}

	data, err := h.dfsAPI.DocGet(sessionId, name, expr, limit)
	if err != nil {
		h.logger.Errorf("doc get: %v", err)
		jsonhttp.InternalServerError(w, "doc get: "+err.Error())
		return
	}

	var docs DocResponse
	for _, d := range data {
		doc := Document{
			Doc: d,
		}
		docs.Docs = append(docs.Docs, doc)
	}

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &docs)
}
