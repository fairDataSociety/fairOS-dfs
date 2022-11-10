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

// KVDeleteHandler godoc
//
//	@Summary      Delete a key value table
//	@Description  KVDeleteHandler is the api handler to delete a key value table
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      kv_table_request body KVTableRequest true "kv table request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/delete [delete]
func (h *Handler) KVDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv delete: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv delete: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVTableRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv delete: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv delete: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv delete: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv delete: \"pod_name\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv delete: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv delete: \"table_name\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv delete: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv delete: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv delete: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.KVDelete(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("kv delete: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv delete: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "kv store deleted"})
}
