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
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DocCountHandler is the api handler to count the number of documents in
// a given document database
// it takes two arguments
// - table_name: the name of the table to count the rows
// - expr: the expression for selecting certain rows
func (h *Handler) DocCountHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc count: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc count: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq common.DocRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc count: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc count: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc count: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc count: \"pod_name\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc count: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc count: \"table_name\" argument missing"})
		return
	}

	expr := docReq.Expression

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc count: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc count: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc count: \"cookie-id\" parameter missing in cookie"})
		return
	}

	count, err := h.dfsAPI.DocCount(sessionId, podName, name, expr)
	if err != nil {
		h.logger.Errorf("doc count: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc count: " + err.Error()})
		return
	}

	jsonhttp.OK(w, count)
}
