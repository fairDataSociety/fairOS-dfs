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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"resenje.org/jsonhttp"
)

// DocIndexRequest is used to index a json file from a pod directly
type DocIndexRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	FileName  string `json:"fileName,omitempty"`
}

// DocIndexJsonHandler godoc
//
//	@Summary      Index a json file that is present in a pod, in to the given document database
//	@Description  DocIndexJsonHandler is the api handler to index a json file that is present in a pod, in to the given document database
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      index_request body DocIndexRequest true "index request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/indexjson [post]
func (h *Handler) DocIndexJsonHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc indexjson: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocIndexRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc indexjson: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc indexjson: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: \"podName\" argument missing"})
		return
	}

	tableName := docReq.TableName
	if tableName == "" {
		h.logger.Errorf("doc indexjson: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: \"tableAme\" argument missing"})
		return
	}

	podFile := docReq.FileName
	if podFile == "" {
		h.logger.Errorf("doc indexjson: \"fileName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: \"fileName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc indexjson: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc indexjson: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc indexjson: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.DocIndexJson(sessionId, podName, tableName, podFile)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrFileNotPresent {
			h.logger.Errorf("doc indexjson: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "doc indexjson: " + err.Error()})
			return
		}
		h.logger.Errorf("doc indexjson: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc indexjson: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "indexing started"})
}
