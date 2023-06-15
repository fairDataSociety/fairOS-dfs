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

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"resenje.org/jsonhttp"
)

// KVTableRequest is the request to create a key value table
type KVTableRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	IndexType string `json:"indexType,omitempty"`
}

// KVCreateHandler godoc
//
//	@Summary      Create a key value table
//	@Description  KVCreateHandler is the api handler to create a key value table
//	@ID		   	  kv-create-handler
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      kv_table_request body KVTableRequest true "kv table request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/new [post]
func (h *Handler) KVCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv create: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv create: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVTableRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv create: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv create: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv create: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv create: \"podName\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv create: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv create: \"tableName\" argument missing"})
		return
	}

	// by default the index type in string
	idxType := kvReq.IndexType
	if idxType == "" {
		idxType = "string"
	}

	var indexType collection.IndexType
	switch idxType {
	case "string":
		indexType = collection.StringIndex
	case "number":
		indexType = collection.NumberIndex
	case "bytes":
	default:
		h.logger.Errorf("kv create: invalid \"indexType\" ")
		jsonhttp.BadRequest(w, &response{Message: "kv create: invalid \"indexType\""})
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

	err = h.dfsAPI.KVCreate(sessionId, podName, name, indexType)
	if err != nil {
		h.logger.Errorf("kv create: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv create: " + err.Error()})
		return
	}
	jsonhttp.Created(w, &response{Message: "kv store created"})
}
