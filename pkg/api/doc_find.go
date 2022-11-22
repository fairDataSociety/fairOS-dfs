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

// DocFindResponse is used for listing rows from a document database
type DocFindResponse struct {
	Docs [][]byte `json:"docs"`
}

// DocFindHandler godoc
//
//	@Summary      Get rows from a given doc datastore
//	@Description  DocFindHandler is the api handler to select rows from a given document datastore
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      expr query string true "expression to search for"
//	@Param	      limit query string false "number od documents"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  DocFindResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/find [get]
func (h *Handler) DocFindHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc find: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc find: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc find: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"tableName\" argument missing"})
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("doc find: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"tableName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["expr"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("pod stat: \"expr\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod stat: \"expr\" argument missing"})
		return
	}
	expr := keys[0]
	if expr == "" {
		h.logger.Errorf("doc find: \"expr\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"expr\" argument missing"})
		return
	}

	limit := ""
	keys, ok = r.URL.Query()["limit"]
	if ok && len(keys[0]) >= 1 {
		limit = keys[0]
	}

	var limitInt int
	if limit == "" {
		limitInt = 10
	} else {
		lmt, err := strconv.Atoi(limit)
		if err != nil {
			h.logger.Errorf("doc find: invalid value for argument \"limit\"")
			jsonhttp.BadRequest(w, &response{Message: "doc find: invalid value for argument \"limit\""})
			return
		}
		limitInt = lmt
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc find: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc find: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc find: \"cookie-id\" parameter missing in cookie"})
		return
	}

	data, err := h.dfsAPI.DocFind(sessionId, podName, name, expr, limitInt)
	if err != nil {
		h.logger.Errorf("doc find: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc find: " + err.Error()})
		return
	}

	var docs DocFindResponse
	docs.Docs = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &docs)
}
