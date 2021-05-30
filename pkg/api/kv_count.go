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
	"fmt"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// KVCountHandler is the api handler to count the number of rows in a key value table
// it has ony one argument
// - table_name: the name of the key value table
func (h *Handler) KVCountHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv count: invalid request body type")
		jsonhttp.BadRequest(w, "kv count: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv count: could not decode arguments")
		jsonhttp.BadRequest(w, "kv count: could not decode arguments")
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv count: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv count: \"pod_name\" argument missing")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv count: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv count: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv count: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv count: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv count: \"cookie-id\" parameter missing in cookie")
		return
	}

	count, err := h.dfsAPI.KVCount(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("kv count: %v", err)
		jsonhttp.InternalServerError(w, "kv count: "+err.Error())
		return
	}

	cntStr := fmt.Sprintf("kv store %s has %d records", name, count)
	jsonhttp.OK(w, cntStr)
}
