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

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"resenje.org/jsonhttp"

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

// DirectoryLsHandler godoc
//
//	@Summary      List directory
//	@Description  DirectoryLsHandler is the api handler for listing the contents of a directory.
//	@ID		      directory-ls-handler
//	@Tags         dir
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      dirPath query string true "dir path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  ListFileResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/ls [get]
func (h *Handler) DirectoryLsHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("ls: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "ls: \"podName\" argument missing"})
		return
	}
	podName := keys[0]

	keys, ok = r.URL.Query()["dirPath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("ls: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "ls: \"dirPath\" argument missing"})
		return
	}
	directory := keys[0]

	// get sessionId from request
	sessionId, err := auth.GetSessionIdFromRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
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
