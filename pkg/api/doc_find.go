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
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type DocFindResponse struct {
	Docs [][]byte `json:"docs"`
}

func (h *Handler) DocFindHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc find: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"name\" argument missing")
		return
	}

	expr := r.FormValue("expr")
	if expr == "" {
		h.logger.Errorf("doc find: \"expr\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"expr\" argument missing")
		return
	}

	limit := r.FormValue("limit")
	var limitInt int
	if limit == "" {
		limitInt = 10
	} else {
		lmt, err := strconv.Atoi(limit)
		if err != nil {
			h.logger.Errorf("doc find: invalid value for argument \"limit\"")
			jsonhttp.BadRequest(w, "doc find: invalid value for argument \"limit\"")
			return
		}
		limitInt = lmt
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc find: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc find: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc find: \"cookie-id\" parameter missing in cookie")
		return
	}

	data, err := h.dfsAPI.DocFind(sessionId, name, expr, limitInt)
	if err != nil {
		h.logger.Errorf("doc find: %v", err)
		jsonhttp.InternalServerError(w, "doc find: "+err.Error())
		return
	}

	var docs DocFindResponse
	docs.Docs = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &docs)
}
