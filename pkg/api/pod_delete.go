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

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
)

func (h *Handler) PodDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod delete: invalid request body type")
		jsonhttp.BadRequest(w, "pod delete: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq common.PodRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod delete: could not decode arguments")
		jsonhttp.BadRequest(w, "pod delete: could not decode arguments")
		return
	}

	podName := podReq.PodName
	if podName == "" {
		h.logger.Errorf("pod delete: \"pod\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod delete: \"pod\" parameter missing in cookie")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("delete pod: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("delete pod: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "delete pod: \"cookie-id\" parameter missing in cookie")
		return
	}

	// delete pod
	err = h.dfsAPI.DeletePod(podName, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn {
			h.logger.Errorf("delete pod: %v", err)
			jsonhttp.BadRequest(w, "delete pod: "+err.Error())
			return
		}
		h.logger.Errorf("delete pod: %v", err)
		jsonhttp.InternalServerError(w, "delete pod: "+err.Error())
		return
	}
	jsonhttp.OK(w, "pod deleted successfully")
}
