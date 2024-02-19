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
)

// DirPresentResponse is used to represent if a directory is present
type DirPresentResponse struct {
	Present bool   `json:"present"`
	Error   string `json:"error,omitempty"`
}

// DirectoryPresentHandler godoc
//
//	@Summary      Is directory present
//	@Description  DirectoryPresentHandler is the api handler which says if a directory is present or not
//	@ID		      directory-present-handler
//	@Tags         dir
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      dirPath query string true "dir path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  DirPresentResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/present [get]
func (h *Handler) DirectoryPresentHandler(w http.ResponseWriter, r *http.Request) {
	driveName, isGroup := "", false
	keys, ok := r.URL.Query()["groupName"]
	if ok || (len(keys) == 1 && len(keys[0]) > 0) {
		driveName = keys[0]
		isGroup = true
	} else {
		keys, ok = r.URL.Query()["podName"]
		if !ok || len(keys[0]) < 1 {
			h.logger.Errorf("dir present: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "dir present: \"podName\" argument missing"})
			return
		}
		driveName = keys[0]
		if driveName == "" {
			h.logger.Errorf("dir present: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "dir present: \"podName\" argument missing"})
			return
		}
	}

	keys, ok = r.URL.Query()["dirPath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("dir present: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir present: \"dirPath\" argument missing"})
		return
	}
	dirToCheck := keys[0]

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

	// check if user is present
	present, err := h.dfsAPI.IsDirPresent(driveName, dirToCheck, sessionId, isGroup)
	if err != nil {
		jsonhttp.OK(w, &DirPresentResponse{
			Present: present,
			Error:   err.Error(),
		})
	} else {
		jsonhttp.OK(w, &DirPresentResponse{
			Present: present,
		})
	}

}
