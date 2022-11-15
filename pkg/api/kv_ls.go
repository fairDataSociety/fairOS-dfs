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
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type Collections struct {
	Tables []Collection
}
type Collection struct {
	Name           string   `json:"table_name"`
	IndexedColumns []string `json:"indexes"`
	CollectionType string   `json:"type"`
}

// KVListHandler godoc
//
//	@Summary      List all key value tables
//	@Description  KVListHandler is the api handler to list all the key value tables in a pod
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      pod_name query string true "pod name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  Collections
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/ls [get]
func (h *Handler) KVListHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv ls: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv ls: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv ls: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv ls: \"cookie-id\" parameter missing in cookie"})
		return
	}

	collections, err := h.dfsAPI.KVList(sessionId, podName)
	if err != nil {
		h.logger.Errorf("kv ls: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv ls: " + err.Error()})
		return
	}

	var col Collections
	for k, v := range collections {
		m := Collection{
			Name:           k,
			IndexedColumns: v,
			CollectionType: "KV Store",
		}
		col.Tables = append(col.Tables, m)
	}

	jsonhttp.OK(w, col)
}
