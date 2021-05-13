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

type PodCreateResponse struct {
	Reference string `json:"reference"`
}

// PodCreateHandler is the api handler to create a new pod
// it takes two arguments
// - pod_name: the name of the pod to create
// - password: the password of the user
func (h *Handler) PodCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod new: invalid request body type")
		jsonhttp.BadRequest(w, "pod new: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq common.PodRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod new: could not decode arguments")
		jsonhttp.BadRequest(w, "pod new: could not decode arguments")
		return
	}

	pod := podReq.PodName
	password := podReq.Password
	if password == "" {
		h.logger.Errorf("pod new: \"password\" argument missing")
		jsonhttp.BadRequest(w, "pod new: \"password\" argument missing")
		return
	}
	if pod == "" {
		h.logger.Errorf("pod new: \"pod\" argument missing")
		jsonhttp.BadRequest(w, "pod new: \"pod\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod new: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod new: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod new: \"cookie-id\" parameter missing in cookie")
		return
	}

	// create pod
	_, err = h.dfsAPI.CreatePod(pod, password, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName ||
			err == p.ErrTooLongPodName ||
			err == p.ErrPodAlreadyExists ||
			err == p.ErrMaxPodsReached {
			h.logger.Errorf("pod new: %v", err)
			jsonhttp.BadRequest(w, "pod new: "+err.Error())
			return
		}
		h.logger.Errorf("pod new: %v", err)
		jsonhttp.InternalServerError(w, "pod new: "+err.Error())
		return
	}

	jsonhttp.Created(w, "pod created successfully")
}
