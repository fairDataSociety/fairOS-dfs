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
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"net/http"
	"resenje.org/jsonhttp"
	"strings"
)

func (h *Handler) DocCreateHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc create: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc  create: \"name\" argument missing")
		return
	}

	// by default, add the index type "id" as stringIndex
	simpleIndexes := make(map[string]collection.IndexType)
	si := r.FormValue("si")
	if si != "" {
		indexes := strings.Split(si, ",")
		for _, idx := range indexes {
			nt := strings.Split(idx, "=")
			if len(nt) != 2 {
				h.logger.Errorf("doc create: \"si\" invalid argument ")
				jsonhttp.BadRequest(w, "doc  create: \"si\" invalid argument")
				return
			}
			switch nt[1] {
			case "string":
				simpleIndexes[nt[0]] = collection.StringIndex
			case "number":
				simpleIndexes[nt[0]] = collection.NumberIndex
			case "bytes":
			default:
				h.logger.Errorf("doc create: invalid \"indexType\" ")
				jsonhttp.BadRequest(w, "doc create: invalid \"indexType\"")
				return
			}
		}
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc create: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc create: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc create: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocCreate(sessionId, name, simpleIndexes)
	if err != nil {
		h.logger.Errorf("doc create: %v", err)
		jsonhttp.InternalServerError(w, "doc create: "+err.Error())
		return
	}
	jsonhttp.OK(w, "document store created")
}
