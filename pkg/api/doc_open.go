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

// DocOpenHandler godoc
//
//	@Summary      Open a doc table
//	@Description  DocOpenHandler is the api handler to open a document database
//	@ID		      doc-open
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_request body DocRequest true "doc table info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  DocumentDBs
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/open [post]
func (h *Handler) DocOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc open: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc open: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc open: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc open: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc open: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc open: \"podName\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc open: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc open: \"tableName\" argument missing"})
		return
	}

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

	err = h.dfsAPI.DocOpen(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("doc open: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc open: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "document store opened"})
}
