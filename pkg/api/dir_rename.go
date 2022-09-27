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

// DirectoryRenameHandler is the api handler to create a new directory.
// it takes two arguments
// - old-path: the directory path to rename along with its absolute path
// - new-path: the new directory path along with its absolute path
func (h *Handler) DirectoryRenameHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("rename-dir: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var renameReq common.RenameRequest
	err := decoder.Decode(&renameReq)
	if err != nil {
		h.logger.Errorf("rename-dir: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: could not decode arguments"})
		return
	}

	podName := renameReq.PodName
	if podName == "" {
		h.logger.Errorf("rename-dir: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"pod_name\" argument missing"})
		return
	}

	oldPath := renameReq.OldPath
	if oldPath == "" {
		h.logger.Errorf("rename-dir: \"old_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"old_path\" argument missing"})
		return
	}

	newPath := renameReq.NewPath
	if newPath == "" {
		h.logger.Errorf("rename-dir: \"new_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"new_path\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("rename-dir: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("rename-dir: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// make directory
	err = h.dfsAPI.RenameDir(podName, oldPath, newPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidDirectory ||
			err == p.ErrTooLongDirectoryName ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("rename-dir: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "rename-dir: " + err.Error()})
			return
		}
		h.logger.Errorf("rename-dir: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "rename-dir: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "directory renamed successfully"})
}
