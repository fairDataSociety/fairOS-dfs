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

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// FileDeleteRequest is used in the file delete request
type FileDeleteRequest struct {
	PodName   string `json:"podName,omitempty"`
	GroupName string `json:"groupName,omitempty"`
	FilePath  string `json:"filePath,omitempty"`
}

// FileDeleteHandler godoc
//
//	@Summary      Delete a file
//	@Description  FileReceiveHandler is the api handler to delete a file from a given pod
//	@ID		      file-delete-handler
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      file_delete_request body FileDeleteRequest true "pod name and file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      404  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/delete [delete]
func (h *Handler) FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file delete: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file delete: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq FileDeleteRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("file delete: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file delete: could not decode arguments"})
		return
	}
	driveName, isGroup := fsReq.GroupName, true
	if driveName == "" {
		driveName = fsReq.PodName
		isGroup = false
		if driveName == "" {
			h.logger.Errorf("file delete: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "file delete: \"podName\" argument missing"})
			return
		}
	}

	podFileWithPath := fsReq.FilePath
	if podFileWithPath == "" {
		h.logger.Errorf("file delete: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file delete: \"filePath\" argument missing"})
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
	// delete file
	err = h.dfsAPI.DeleteFile(driveName, podFileWithPath, sessionId, isGroup)
	if err != nil {
		if errors.Is(err, dfs.ErrPodNotOpen) {
			h.logger.Errorf("file delete: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "file delete: " + err.Error()})
			return
		}
		if errors.Is(err, pod.ErrInvalidFile) {
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
