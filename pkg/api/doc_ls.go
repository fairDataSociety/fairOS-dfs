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

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DocumentDBs represent a list of documentDB
type DocumentDBs struct {
	Tables []documentDB
}

type documentDB struct {
	Name           string              `json:"table_name"`
	IndexedColumns []collection.SIndex `json:"indexes"`
	CollectionType string              `json:"type"`
}

// DocListHandler is the api handler which lists all the document database in a pod
// it takes no arguments
func (h *Handler) DocListHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc ls: \"pod_name\" argument missing")
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc ls: \"pod_name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc ls: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc ls: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc ls: \"cookie-id\" parameter missing in cookie")
		return
	}

	collections, err := h.dfsAPI.DocList(sessionId, podName)
	if err != nil {
		h.logger.Errorf("doc ls: %v", err)
		jsonhttp.InternalServerError(w, "doc ls: "+err.Error())
		return
	}

	var col DocumentDBs
	for name, dbSchema := range collections {
		var indexes []collection.SIndex
		indexes = append(indexes, dbSchema.SimpleIndexes...)
		indexes = append(indexes, dbSchema.MapIndexes...)
		indexes = append(indexes, dbSchema.ListIndexes...)
		m := documentDB{
			Name:           name,
			IndexedColumns: indexes,
			CollectionType: "Document Store",
		}
		col.Tables = append(col.Tables, m)
	}

	jsonhttp.OK(w, col)
}
