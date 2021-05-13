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

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// PodSyncHandler is the api handler to sync a pod's contents from the Swarm network
// it takes no arguments
func (h *Handler) PodSyncHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod sync: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod sync: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod sync: \"cookie-id\" parameter missing in cookie")
		return
	}

	// fetch pods and list them
	err = h.dfsAPI.SyncPod(sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName ||
			err == p.ErrTooLongPodName ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("pod sync: %v", err)
			jsonhttp.BadRequest(w, "pod sync: "+err.Error())
			return
		}
		h.logger.Errorf("pod sync: %v", err)
		jsonhttp.InternalServerError(w, "pod sync: "+err.Error())
		return
	}
	jsonhttp.OK(w, "pod synced successfully")
}
