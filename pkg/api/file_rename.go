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

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// FileRenameHandler godoc
//
//	@Summary      Info of a file
//	@Description  FileRenameHandler is the api handler to get the information of a file
//	@ID		      file-rename-handler
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      rename_request body common.RenameRequest true "old name & new name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      404  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/rename [post]
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

	driveName, isGroup := renameReq.GroupName, true
	if driveName == "" {
		driveName = renameReq.PodName
		isGroup = false
		if driveName == "" {
			h.logger.Errorf("file rename: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "file rename: \"podName\" argument missing"})
			return
		}
	}

	podFileWithPath := renameReq.OldPath
	if podFileWithPath == "" {
		h.logger.Errorf("file rename: \"oldPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"oldPath\" argument missing"})
		return
	}

	newPodFileWithPath := renameReq.NewPath
	if newPodFileWithPath == "" {
		h.logger.Errorf("file rename: \"newPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file rename: \"newPath\" argument missing"})
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
	// rename file
	err = h.dfsAPI.RenameFile(driveName, podFileWithPath, newPodFileWithPath, sessionId, isGroup)
	if err != nil {
		if errors.Is(err, dfs.ErrPodNotOpen) {
			h.logger.Errorf("file rename: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "file rename: " + err.Error()})
			return
		}
		if errors.Is(err, pod.ErrInvalidFile) {
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
