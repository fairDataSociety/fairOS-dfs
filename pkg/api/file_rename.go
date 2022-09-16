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
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// FileRenameHandler is the api handler to rename a file from a given pod
// it takes two arguments
// - old-path: the file path to rename along with its absolute path
// - new-path: the new file path along with its absolute path
func (h *Handler) FileRenameHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file rename: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file rename: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var renameReq common.RenameRequest
	err := decoder.Decode(&renameReq)
	if err != nil {
		h.logger.Errorf("file rename: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file rename: could not decode arguments"})
		return
	}

	podName := renameReq.PodName
	if podName == "" {
		h.logger.Errorf("file rename: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"pod_name\" argument missing"})
		return
	}

	podFileWithPath := renameReq.OldPath
	if podFileWithPath == "" {
		h.logger.Errorf("file rename: \"old_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"old_path\" argument missing"})
		return
	}

	newPodFileWithPath := renameReq.NewPath
	if newPodFileWithPath == "" {
		h.logger.Errorf("file rename: \"new_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"new_path\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file rename: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file rename: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"cookie-id\" parameter missing in cookie"})
		return
	}
	// delete file
	err = h.dfsAPI.RenameFile(podName, podFileWithPath, newPodFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("file rename: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "file rename: " + err.Error()})
			return
		}
		if err == pod.ErrInvalidFile {
			h.logger.Errorf("file rename: %v", err)
			jsonhttp.NotFound(w, &response{Message: "file rename: " + err.Error()})
			return
		}
		h.logger.Errorf("file rename: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file rename: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "file renamed successfully"})
}