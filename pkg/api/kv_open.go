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

// KVOpenHandler godoc
//
//	@Summary      Open a key value table
//	@Description  KVOpenHandler is the api handler to open a key value table
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      kv_table_request body KVTableRequest true "kv table request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/open [post]
func (h *Handler) KVOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv open: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv open: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVTableRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv open: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv open: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv open: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv open: \"podName\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv open: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv open: \"tableName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv open: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.KVOpen(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("kv open: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv open: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "kv store opened"})
}
