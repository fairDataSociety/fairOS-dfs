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
)

func (h *Handler) FileStatHandler(w http.ResponseWriter, r *http.Request) {
	podFileWithPath := r.FormValue("pod_path_file")
	if podFileWithPath == "" {
		h.logger.Errorf("file stat: \"pod_path_file\" argument missing")
		jsonhttp.BadRequest(w, "file stat: \"pod_path_file\" argument missing")
		return
	}

	podPath := r.FormValue("pod_path")
	if podPath == "" {
		h.logger.Errorf("file stat: \"pod_path\" argument missing")
		jsonhttp.BadRequest(w, "file pod_path: \"path\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file stat: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file stat: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "file stat: \"cookie-id\" parameter missing in cookie")
		return
	}

	// get file stat
	stat, err := h.dfsAPI.FileStat(podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("file stat: %v", err)
			jsonhttp.BadRequest(w, "file stat: "+err.Error())
			return
		}
		h.logger.Errorf("file stat: %v", err)
		jsonhttp.InternalServerError(w, "file stat: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, stat)
}
