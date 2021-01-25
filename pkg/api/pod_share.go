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
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"net/http"
	"resenje.org/jsonhttp"
)

type PodSharingReference struct {
	Reference string `json:"pod_sharing_reference"`
}

func (h *Handler) PodShareHandler(w http.ResponseWriter, r *http.Request) {
	pod := r.FormValue("pod")
	if pod == "" {
		h.logger.Errorf("pod share: \"pod\" argument missing")
		jsonhttp.BadRequest(w, "pod share: \"pod\" argument missing")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		h.logger.Errorf("pod share: \"password\" argument missing")
		jsonhttp.BadRequest(w, "pod share: \"password\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod share: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod share: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod stat: \"cookie-id\" parameter missing in cookie")
		return
	}

	// fetch pod stat
	sharingRef, err := h.dfsAPI.PodShare(pod, password, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod share: %v", err)
			jsonhttp.BadRequest(w, "pod share: "+ err.Error())
			return
		}
		h.logger.Errorf("pod share: %v", err)
		jsonhttp.InternalServerError(w, "pod share: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &PodSharingReference{
		Reference: sharingRef,
	})
}

func (h *Handler) PodReceiveInfoHandler(w http.ResponseWriter, r *http.Request) {
	sharingRefString := r.FormValue("ref")
	if sharingRefString == "" {
		h.logger.Errorf("pod receive info: \"ref\" argument missing")
		jsonhttp.BadRequest(w, "pod receive info: \"ref\" argument missing")
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

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("pod receive info: invalid reference: ", err)
		jsonhttp.BadRequest(w, "pod receive info: invalid reference:"+err.Error())
		return
	}

	shareInfo, err := h.dfsAPI.PodReceiveInfo(sessionId, sharingRef)
	if err != nil {
		h.logger.Errorf("pod receive info: %v", err)
		jsonhttp.InternalServerError(w, "pod receive info: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, shareInfo)
}

func (h *Handler) PodReceiveHandler(w http.ResponseWriter, r *http.Request) {
	sharingRefString := r.FormValue("ref")
	if sharingRefString == "" {
		h.logger.Errorf("pod receive: \"ref\" argument missing")
		jsonhttp.BadRequest(w, "pod receive: \"ref\" argument missing")
		return
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

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("pod receive: invalid reference: ", err)
		jsonhttp.BadRequest(w, "pod receive: invalid reference:"+err.Error())
		return
	}

	err := h.dfsAPI.PodReceive(sessionId, sharingRef)
	if err != nil {
		h.logger.Errorf("pod receive: %v", err)
		jsonhttp.InternalServerError(w, "pod receive: "+err.Error())
		return
	}

	jsonhttp.OK(w, "pod added")
}