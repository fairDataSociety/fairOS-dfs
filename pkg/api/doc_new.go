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
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DocRequest
type DocRequest struct {
	PodName     string `json:"podName,omitempty"`
	TableName   string `json:"tableName,omitempty"`
	SimpleIndex string `json:"si,omitempty"`
	Mutable     bool   `json:"mutable,omitempty"`
}

// SimpleDocRequest
type SimpleDocRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
}

// DocCreateHandler godoc
//
//	@Summary      Create in doc table
//	@Description  DocCreateHandler is the api handler to create a new document database
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_request body DocRequest true "doc table info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/new [post]
func (h *Handler) DocCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc create: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc create: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc create: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc create: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc create: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc create: \"podName\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc create: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc  create: \"tableName\" argument missing"})
		return
	}

	// by default, add the index type "id" as stringIndex
	indexes := make(map[string]collection.IndexType)
	si := docReq.SimpleIndex
	if si != "" {
		idxs := strings.Split(si, ",")
		for _, idx := range idxs {
			nt := strings.Split(idx, "=")
			if len(nt) != 2 {
				h.logger.Errorf("doc create: \"si\" invalid argument ")
				jsonhttp.BadRequest(w, &response{Message: "doc  create: \"si\" invalid argument"})
				return
			}
			switch nt[1] {
			case "string":
				indexes[nt[0]] = collection.StringIndex
			case "number":
				indexes[nt[0]] = collection.NumberIndex
			case "map":
				indexes[nt[0]] = collection.MapIndex
			case "list":
				indexes[nt[0]] = collection.ListIndex
			case "bytes":
			default:
				h.logger.Errorf("doc create: invalid \"indexType\" ")
				jsonhttp.BadRequest(w, &response{Message: "doc create: invalid \"indexType\""})
				return
			}
		}
	}

	mutable := docReq.Mutable

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc create: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc create: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc create: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.DocCreate(sessionId, podName, name, indexes, mutable)
	if err != nil {
		h.logger.Errorf("doc create: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc create: " + err.Error()})
		return
	}

	jsonhttp.Created(w, &response{Message: "document db created"})
}
