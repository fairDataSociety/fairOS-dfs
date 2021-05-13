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
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

type PodOpenResponse struct {
	Reference string `json:"reference"`
}

// PodOpenHandler is the api handler to close a open pod
// it takes two arguments
// - pod_name: the name of the pod to open
// - password: the password of the user
func (h *Handler) PodOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod open: invalid request body type")
		jsonhttp.BadRequest(w, "pod open: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq common.PodRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod open: could not decode arguments")
		jsonhttp.BadRequest(w, "pod open: could not decode arguments")
		return
	}
	pod := podReq.PodName

	// password will be empty in case of opening a shared pod
	// so allow even if it is not set
	password := podReq.Password

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod open: \"cookie-id\" parameter missing in cookie")
		return
	}

	// open pod
	_, err = h.dfsAPI.OpenPod(pod, password, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod open: %v", err)
			jsonhttp.BadRequest(w, "pod open: "+err.Error())
			return
		}
		h.logger.Errorf("pod open: %v", err)
		jsonhttp.InternalServerError(w, "pod open: "+err.Error())
		return
	}

	jsonhttp.OK(w, "pod opened successfully")
}
