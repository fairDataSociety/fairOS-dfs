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

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// PodListResponse is the json response to list all the pods
type PodListResponse struct {
	Pods       []string `json:"pods"`
	SharedPods []string `json:"sharedPods"`
}

// PodListHandler godoc
//
//	@Summary      List pods
//	@Description  PodListHandler is the api handler to list all pods
//	@ID           pod-list-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  PodListResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/ls [get]
func (h *Handler) PodListHandler(w http.ResponseWriter, r *http.Request) {
	// get sessionId from request
	sessionId, err := auth.GetSessionIdFromRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}

	// fetch pods and list them
	pods, sharedPods, err := h.dfsAPI.ListPods(sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == pod.ErrPodNotOpened {
			h.logger.Errorf("ls pod: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "ls pod: " + err.Error()})
			return
		}
		h.logger.Errorf("ls pod: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "ls pod: " + err.Error()})
		return
	}
	if pods == nil {
		pods = make([]string, 0)
	}
	if sharedPods == nil {
		sharedPods = make([]string, 0)
	}
	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &PodListResponse{
		Pods:       pods,
		SharedPods: sharedPods,
	})
}
