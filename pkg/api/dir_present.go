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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type DirPresentResponse struct {
	Present bool   `json:"present"`
	Error   string `json:"error,omitempty"`
}

// DirectoryPresentHandler is the api handler which says if a directory is present or not
// it takes only one argument
// - dir-path: the directory to check along with its absolute path
func (h *Handler) DirectoryPresentHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("dir present: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "dir present: \"pod_name\" argument missing")
		return
	}
	podName := keys[0]

	keys, ok = r.URL.Query()["dir_path"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("dir present: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, "dir present: \"dir_path\" argument missing")
		return
	}
	dirToCheck := keys[0]

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("dir present: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("dir present: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "dir present: \"cookie-id\" parameter missing in cookie")
		return
	}

	// check if user is present
	present, err := h.dfsAPI.IsDirPresent(podName, dirToCheck, sessionId)
	if err != nil {
		jsonhttp.OK(w, &DirPresentResponse{
			Present: present,
			Error:   err.Error(),
		})
	} else {
		jsonhttp.OK(w, &DirPresentResponse{
			Present: present,
		})
	}

}
