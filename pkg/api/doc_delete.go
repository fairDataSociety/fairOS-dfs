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

// DocDeleteHandler is the api handler to delete the given document database
// it takes only one argument
// table_name: the document database to delete
func (h *Handler) DocDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc delete: invalid request body type")
		jsonhttp.BadRequest(w, "doc delete: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq common.DocRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc delete: could not decode arguments")
		jsonhttp.BadRequest(w, "doc delete: could not decode arguments")
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc delete: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "doc  delete: \"table_name\" argument missing")
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc delete: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc delete: \"pod_name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc delete: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc delete: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc delete: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocDelete(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("doc delete: %v", err)
		jsonhttp.InternalServerError(w, "doc delete: "+err.Error())
		return
	}
	jsonhttp.OK(w, "document store deleted")
}
