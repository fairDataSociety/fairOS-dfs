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

// FileDeleteHandler is the api handler to delete a file from a given pod
//
//	it takes only one argument
//
// file_path: the absolute path of the file in the pod
func (h *Handler) FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file delete: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file delete: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq common.FileSystemRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("file delete: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file delete: could not decode arguments"})
		return
	}

	podName := fsReq.PodName
	if podName == "" {
		h.logger.Errorf("file delete: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file delete: \"pod_name\" argument missing"})
		return
	}

	podFileWithPath := fsReq.FilePath
	if podFileWithPath == "" {
		h.logger.Errorf("file delete: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file delete: \"file_path\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file delete: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file delete: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file delete: \"cookie-id\" parameter missing in cookie"})
		return
	}
	// delete file
	err = h.dfsAPI.DeleteFile(podName, podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("file delete: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "file delete: " + err.Error()})
			return
		}
		if err == pod.ErrInvalidFile {
			h.logger.Errorf("file delete: %v", err)
			jsonhttp.NotFound(w, &response{Message: "file delete: " + err.Error()})
			return
		}
		h.logger.Errorf("file delete: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file delete: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "file deleted successfully"})
}
