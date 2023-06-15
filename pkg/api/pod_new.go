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

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// PodCreateHandler godoc
//
//	@Summary      Create pod
//	@Description  PodCreateHandler is the api handler to create a new pod
//	@ID           pod-create-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodNameRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/new [post]
func (h *Handler) PodCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod new: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod new: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq PodNameRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod new: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod new: could not decode arguments"})
		return
	}

	pod := podReq.PodName
	if pod == "" {
		h.logger.Errorf("pod new: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod new: \"podName\" argument missing"})
		return
	}

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

	// create pod
	_, err = h.dfsAPI.CreatePod(pod, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName ||
			err == p.ErrTooLongPodName ||
			err == p.ErrPodAlreadyExists ||
			err == p.ErrMaxPodsReached {
			h.logger.Errorf("pod new: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "pod new: " + err.Error()})
			return
		}
		h.logger.Errorf("pod new: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod new: " + err.Error()})
		return
	}

	jsonhttp.Created(w, &response{Message: "pod created successfully"})
}
