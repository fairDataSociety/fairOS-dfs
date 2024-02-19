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
	"errors"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// DirRequest is used to create directory
type DirRequest struct {
	PodName       string `json:"podName,omitempty"`
	GroupName     string `json:"groupName,omitempty"`
	DirectoryPath string `json:"dirPath,omitempty"`
}

// DirectoryMkdirHandler godoc
//
//	@Summary      Create directory
//	@Description  DirectoryMkdirHandler is the api handler to create a new directory.
//	@ID		      directory-mkdir-handler
//	@Tags         dir
//	@Accept       json
//	@Produce      json
//	@Param	      dir_request body DirRequest true "pod name and dir path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/mkdir [post]
func (h *Handler) DirectoryMkdirHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("mkdir: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "mkdir: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq DirRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("mkdir: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "mkdir: could not decode arguments"})
		return
	}
	driveName, isGroup := fsReq.GroupName, true
	if driveName == "" {
		driveName = fsReq.PodName
		isGroup = false
		if driveName == "" {
			h.logger.Errorf("mkdir: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "mkdir: \"podName\" argument missing"})
			return
		}
	}

	dirToCreateWithPath := fsReq.DirectoryPath
	if dirToCreateWithPath == "" {
		h.logger.Errorf("mkdir: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "mkdir: \"dirPath\" argument missing"})
		return
	}

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

	// make directory
	err = h.dfsAPI.Mkdir(driveName, dirToCreateWithPath, sessionId, 0, isGroup)
	if err != nil {
		if errors.Is(err, dfs.ErrPodNotOpen) || errors.Is(err, dfs.ErrUserNotLoggedIn) ||
			errors.Is(err, p.ErrInvalidDirectory) ||
			errors.Is(err, p.ErrTooLongDirectoryName) ||
			errors.Is(err, p.ErrPodNotOpened) {
			h.logger.Errorf("mkdir: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "mkdir: " + err.Error()})
			return
		}
		h.logger.Errorf("mkdir: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "mkdir: " + err.Error()})
		return
	}
	jsonhttp.Created(w, &response{Message: "directory created successfully"})
}
