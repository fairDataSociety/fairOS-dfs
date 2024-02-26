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

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// DirectoryRenameHandler godoc
//
//	@Summary      Rename directory
//	@Description  DirectoryRenameHandler is the api handler to rename a directory.
//	@ID		      directory-rename-handler
//	@Tags         dir
//	@Accept       json
//	@Produce      json
//	@Param	      dir_request body common.RenameRequest true "old name and new path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/rename [post]
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
	driveName, isGroup := renameReq.GroupName, true
	if driveName == "" {
		driveName = renameReq.PodName
		isGroup = false
		if driveName == "" {
			h.logger.Errorf("rename-dir: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"podName\" argument missing"})
			return
		}
	}

	oldPath := renameReq.OldPath
	if oldPath == "" {
		h.logger.Errorf("rename-dir: \"oldPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"oldPath\" argument missing"})
		return
	}

	newPath := renameReq.NewPath
	if newPath == "" {
		h.logger.Errorf("rename-dir: \"newPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "rename-dir: \"newPath\" argument missing"})
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
	err = h.dfsAPI.RenameDir(driveName, oldPath, newPath, sessionId, isGroup)
	if err != nil {
		if errors.Is(err, dfs.ErrPodNotOpen) || errors.Is(err, dfs.ErrUserNotLoggedIn) ||
			errors.Is(err, p.ErrInvalidDirectory) ||
			errors.Is(err, p.ErrTooLongDirectoryName) ||
			errors.Is(err, p.ErrPodNotOpened) {
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
