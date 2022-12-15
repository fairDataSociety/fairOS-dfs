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
	"fmt"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"resenje.org/jsonhttp"
)

// PodSharingReference
type PodSharingReference struct {
	Reference string `json:"podSharingReference"`
}

// PodShareHandler godoc
//
//	@Summary      Share pod
//	@Description  PodShareHandler is the api handler to share a pod to the public
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body common.PodShareRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  PodSharingReference
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/share [post]
func (h *Handler) PodShareHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod share: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod share: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq common.PodShareRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod share: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod share: could not decode arguments"})
		return
	}

	pod := podReq.PodName
	if pod == "" {
		h.logger.Errorf("pod share: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod share: \"podName\" argument missing"})
		return
	}

	sharedPodName := podReq.SharedPodName
	if sharedPodName == "" {
		sharedPodName = pod
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod share: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod share: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "pod stat: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// fetch pod stat
	sharingRef, err := h.dfsAPI.PodShare(pod, sharedPodName, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod share: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "pod share: " + err.Error()})
			return
		}
		h.logger.Errorf("pod share: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod share: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &PodSharingReference{
		Reference: sharingRef,
	})
}

// PodReceiveInfoHandler godoc
//
//	@Summary      Receive shared pod info
//	@Description  PodReceiveInfoHandler is the api handler to receive shared pod info from shared reference
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/receiveinfo [get]
func (h *Handler) PodReceiveInfoHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("pod receive info: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "pod receive info: \"sharingRef\" argument missing")
		return
	}

	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("pod receive info: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "pod receive info: \"sharingRef\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod receive info: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod receive info: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod receive info: \"cookie-id\" parameter missing in cookie")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("pod receive info: invalid reference: ", err)
		jsonhttp.BadRequest(w, "pod receive info: invalid reference:"+err.Error())
		return
	}

	shareInfo, err := h.dfsAPI.PodReceiveInfo(sessionId, ref)
	if err != nil {
		h.logger.Errorf("pod receive info: %v", err)
		jsonhttp.InternalServerError(w, "pod receive info: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, shareInfo)
}

// PodReceiveHandler godoc
//
//	@Summary      Receive shared pod
//	@Description  PodReceiveHandler is the api handler to receive shared pod from shared reference
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      sharedPodName query string false "pod name to be saved as"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/receive [get]
func (h *Handler) PodReceiveHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("pod receive: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "pod receive: \"sharingRef\" argument missing")
		return
	}

	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("pod receive: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "pod receive: \"sharingRef\" argument missing")
		return
	}

	sharedPodName := ""
	keys, ok = r.URL.Query()["sharedPodName"]
	if ok && len(keys[0]) == 1 {
		sharedPodName = keys[0]
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod receive: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod receive: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod receive: \"cookie-id\" parameter missing in cookie")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("pod receive: invalid reference: ", err)
		jsonhttp.BadRequest(w, "pod receive: invalid reference:"+err.Error())
		return
	}

	pi, err := h.dfsAPI.PodReceive(sessionId, sharedPodName, ref)
	if err != nil {
		h.logger.Errorf("pod receive: %v", err)
		jsonhttp.InternalServerError(w, "pod receive: "+err.Error())
		return
	}

	addedStr := fmt.Sprintf("public pod %q, added as shared pod", pi.GetPodName())
	jsonhttp.OK(w, &response{Message: addedStr})
}
