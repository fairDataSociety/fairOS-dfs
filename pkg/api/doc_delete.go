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

// DocDeleteHandler godoc
//
//	@Summary      Delete a doc table
//	@Description  DocDeleteHandler is the api handler to delete the given document database
//	@ID		      doc-delete
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_request body SimpleDocRequest true "doc table info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/delete [delete]
func (h *Handler) DocDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc delete: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc delete: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq SimpleDocRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc delete: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc delete: could not decode arguments"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc delete: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc  delete: \"tableName\" argument missing"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc delete: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc delete: \"podName\" argument missing"})
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
	err = h.dfsAPI.DocDelete(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("doc delete: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc delete: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "document store deleted"})
}
