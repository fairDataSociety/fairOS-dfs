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

// FileStatHandler godoc
//
//	@Summary      Info of a file
//	@Description  FileStatHandler is the api handler to get the information of a file
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      pod_name query string true "pod name"
//	@Param	      pod_name query string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  file.Stats
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/stat [get]
func (h *Handler) FileStatHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file stat: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file stat: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("file stat: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file stat: \"pod_name\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["file_path"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file stat: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file stat: \"file_path\" argument missing"})
		return
	}
	podFileWithPath := keys[0]
	if podFileWithPath == "" {
		h.logger.Errorf("file stat: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file stat: \"pod_path_file\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file stat: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file stat: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file stat: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// get file stat
	stat, err := h.dfsAPI.FileStat(podName, podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("file stat: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "file stat: " + err.Error()})
			return
		}
		h.logger.Errorf("file stat: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file stat: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, stat)
}
