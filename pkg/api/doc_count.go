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

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"resenje.org/jsonhttp"
)

// DocCountRequest is used for doc count
type DocCountRequest struct {
	PodName     string `json:"podName,omitempty"`
	TableName   string `json:"tableName,omitempty"`
	SimpleIndex string `json:"si,omitempty"`
	Mutable     bool   `json:"mutable,omitempty"`
	Expression  string `json:"expr,omitempty"`
}

// DocCountHandler godoc
//
//	@Summary      Count number of document in a table
//	@Description  DocCountHandler is the api handler to count the number of documents in a given document database
//	@ID		      doc-count
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_request body DocCountRequest true "doc table info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  collection.TableKeyCount
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/count [post]
func (h *Handler) DocCountHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc count: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc count: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocCountRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc count: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc count: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc count: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc count: \"podName\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc count: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc count: \"tableName\" argument missing"})
		return
	}

	expr := docReq.Expression

	// get sessionId from request
	sessionId, err := auth.GetSessionIdFromRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
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
