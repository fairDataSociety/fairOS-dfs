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

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// KVCreateHandler is the api handler to create a key value table
// it takes two arguments
// - table_name: the name of the kv table
// - index_type: the name of the index (ex: string, number)
func (h *Handler) KVCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv create: invalid request body type")
		jsonhttp.BadRequest(w, "kv create: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv create: could not decode arguments")
		jsonhttp.BadRequest(w, "kv create: could not decode arguments")
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv create: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv create: \"pod_name\" argument missing")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv create: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "kv create: \"table_name\" argument missing")
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
		jsonhttp.BadRequest(w, "kv create: invalid \"indexType\"")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv create: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv create: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv create: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.KVCreate(sessionId, podName, name, indexType)
	if err != nil {
		h.logger.Errorf("kv create: %v", err)
		jsonhttp.InternalServerError(w, "kv create: "+err.Error())
		return
	}
	jsonhttp.Created(w, "kv store created")
}
