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
	"resenje.org/jsonhttp"
)

// KVCountHandler is the api handler to count the number of rows in a key value table
// it has ony one argument
// - table_name: the name of the key value table

// KVCountHandler godoc
//
//	@Summary      Count rows in a key value table
//	@Description  KVCountHandler is the api handler to count the number of rows in a key value table
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      kv_table_request body KVTableRequest true "kv table request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  collection.TableKeyCount
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/count [post]
func (h *Handler) KVCountHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv count: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv count: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVTableRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv count: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv count: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv count: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv count: \"podName\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv count: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv count: \"tableName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv count: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv count: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv count: \"cookie-id\" parameter missing in cookie"})
		return
	}

	count, err := h.dfsAPI.KVCount(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("kv count: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv count: " + err.Error()})
		return
	}

	jsonhttp.OK(w, count)
}
