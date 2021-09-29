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

// DocFindHandler is the api handler to select rows from a given document database
// it takes three arguments
// table_name: the document database from which to select the rows
// expr: the expression which helps in selection particular rows
// limit: the threshold of documents to return in the result
func (h *Handler) DocFindHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc find: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"pod_name\" argument missing")
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc find: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"pod_name\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["table_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc find: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"table_name\" argument missing")
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("doc find: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"table_name\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["expr"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("pod stat: \"expr\" argument missing")
		jsonhttp.BadRequest(w, "pod stat: \"expr\" argument missing")
		return
	}
	expr := keys[0]
	if expr == "" {
		h.logger.Errorf("doc find: \"expr\" argument missing")
		jsonhttp.BadRequest(w, "doc find: \"expr\" argument missing")
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

	data, err := h.dfsAPI.DocFind(sessionId, podName, name, expr, limitInt)
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
