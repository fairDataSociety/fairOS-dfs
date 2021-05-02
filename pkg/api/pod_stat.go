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

type PodStatResponse struct {
	PodName          string `json:"name"`
	PodAddress          string `json:"address"`
}

func (h *Handler) PodStatHandler(w http.ResponseWriter, r *http.Request) {
	pod := r.FormValue("pod")
	if pod == "" {
		h.logger.Errorf("pod stat: \"pod\" argument missing")
		jsonhttp.BadRequest(w, "pod stat: \"pod\" argument missing")
		return
	}
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod stat: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod stat: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod stat: \"cookie-id\" parameter missing in cookie")
		return
	}

	// fetch pod stat
	stat, err := h.dfsAPI.PodStat(pod, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod stat: %v", err)
			jsonhttp.BadRequest(w, "pod stat: "+err.Error())
			return
		}
		h.logger.Errorf("pod stat: %v", err)
		jsonhttp.InternalServerError(w, "pod stat: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &PodStatResponse{
		PodName:          stat.PodName,
		PodAddress:       stat.PodAddress,
	})
}
