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
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// ListFileResponse is used to list directories and files
type ListFileResponse struct {
	Directories []dir.Entry  `json:"dirs,omitempty"`
	Files       []file.Entry `json:"files,omitempty"`
}

// DirectoryLsHandler is the api handler for listing the contents of a directory.
// it takes only one argument
// - dir_path: the path of the directory to list it contents
func (h *Handler) DirectoryLsHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "ls: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]

	keys, ok = r.URL.Query()["dir_path"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("ls: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "ls: \"dir_path\" argument missing"})
		return
	}
	directory := keys[0]

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("ls: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("ls: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "ls: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// list directory
	dEntries, fEntries, err := h.dfsAPI.ListDir(podName, directory, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("ls: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "ls: " + err.Error()})
			return
		}
		if err == dir.ErrDirectoryNotPresent {
			h.logger.Errorf("ls: %v", err)
			jsonhttp.NotFound(w, &response{Message: "ls: " + err.Error()})
			return
		}
		h.logger.Errorf("ls: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "ls: " + err.Error()})
		return
	}

	if dEntries == nil {
		dEntries = make([]dir.Entry, 0)
	}
	if fEntries == nil {
		fEntries = make([]file.Entry, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &ListFileResponse{
		Directories: dEntries,
		Files:       fEntries,
	})
}
