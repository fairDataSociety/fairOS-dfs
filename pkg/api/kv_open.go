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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// KVOpenHandler is the api handler to open a key value table
// it takes only one argument
// - table_name: the name of the kv table
func (h *Handler) KVOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv open: invalid request body type")
		jsonhttp.BadRequest(w, "kv open: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv open: could not decode arguments")
		jsonhttp.BadRequest(w, "kv open: could not decode arguments")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv open: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv open: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv open: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.KVOpen(sessionId, name)
	if err != nil {
		h.logger.Errorf("kv open: %v", err)
		jsonhttp.InternalServerError(w, "kv open: "+err.Error())
		return
	}
	jsonhttp.OK(w, "kv store opened")
}
