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

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// PodSyncHandler godoc
//
//	@Summary      Sync pod
//	@Description  PodSyncHandler is the api handler to sync a pod's content
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodNameRequest true "pod name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/sync [post]
func (h *Handler) PodSyncHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod sync: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod sync: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq PodNameRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod sync: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod sync: could not decode arguments"})
		return
	}
	podName := podReq.PodName
	if podName == "" {
		h.logger.Errorf("pod sync: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod sync: \"podName\" argument missing"})
		return
	}
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod sync: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod sync: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "pod sync: \"cookie-id\" parameter missing in cookie"})
		return
	}
	// fetch pods and list them
	err = h.dfsAPI.SyncPod(podName, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName ||
			err == p.ErrTooLongPodName ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("pod sync: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "pod sync: " + err.Error()})
			return
		}
		h.logger.Errorf("pod sync: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod sync: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "pod synced successfully"})
}
