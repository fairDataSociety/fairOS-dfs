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
	"errors"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// PodOpenHandler godoc
//
//	@Summary      Open pod
//	@Description  PodOpenHandler is the api handler to open pod
//	@ID           pod-open-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodNameRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/open [post]
func (h *Handler) PodOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod open: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod open: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq PodNameRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod open: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod open: could not decode arguments"})
		return
	}
	pod := podReq.PodName
	if pod == "" {
		h.logger.Errorf("pod open: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod open: \"podName\" argument missing"})
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

	// open pod
	_, err = h.dfsAPI.OpenPod(pod, sessionId)
	if err != nil {
		if errors.Is(err, dfs.ErrUserNotLoggedIn) ||
			errors.Is(err, p.ErrInvalidPodName) {
			h.logger.Errorf("pod open: %v", err)
			jsonhttp.NotFound(w, &response{Message: "pod open: " + err.Error()})
			return
		}
		h.logger.Errorf("pod open: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod open: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "pod opened successfully"})
}

// PodOpenAsyncHandler godoc
//
//	@Summary      Open pod
//	@Description  PodOpenAsyncHandler is the api handler to open pod asynchronously
//	@ID           pod-open-async-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodNameRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Deprecated
//	@Router       /v1/pod/open-async [post]
func (h *Handler) PodOpenAsyncHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.BadRequest(w, &response{Message: "pod/open-async: deprecated"})
}
