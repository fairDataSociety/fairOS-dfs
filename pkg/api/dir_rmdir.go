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

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// DirectoryRmdirHandler godoc
//
//	@Summary      Remove directory
//	@Description  DirectoryRmdirHandler is the api handler to remove a directory.
//	@Tags         dir
//	@Accept       json
//	@Produce      json
//	@Param	      dir_request body DirRequest true "pod name and dir path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/rmdir [delete]
func (h *Handler) DirectoryRmdirHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("rmdir: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "rmdir: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq DirRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("rmdir: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "rmdir: could not decode arguments"})
		return
	}

	podName := fsReq.PodName
	if podName == "" {
		h.logger.Errorf("rmdir: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rmdir: \"pod_name\" argument missing"})
		return
	}

	dir := fsReq.DirectoryPath
	if dir == "" {
		h.logger.Errorf("rmdir: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rmdir: \"dir_path\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Error("rmdir: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("rmdir: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "rmdir: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// remove directory
	err = h.dfsAPI.RmDir(podName, dir, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("rmdir: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "rmdir: " + err.Error()})
			return
		}
		h.logger.Errorf("rmdir: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "rmdir: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "directory removed successfully"})
}
