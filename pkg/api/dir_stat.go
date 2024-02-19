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
	"errors"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// DirectoryStatHandler godoc
//
//	@Summary      Directory stat
//	@Description  DirectoryStatHandler is the api handler which gives the information about a directory
//	@ID		      directory-stat-handler
//	@Tags         dir
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      dirPath query string true "dir path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  dir.Stats
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/stat [get]
func (h *Handler) DirectoryStatHandler(w http.ResponseWriter, r *http.Request) {
	driveName, isGroup := "", false
	keys, ok := r.URL.Query()["groupName"]
	if ok || (len(keys) == 1 && len(keys[0]) > 0) {
		driveName = keys[0]
		isGroup = true
	} else {
		keys, ok = r.URL.Query()["podName"]
		if !ok || len(keys[0]) < 1 {
			h.logger.Errorf("dir: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "dir: \"podName\" argument missing"})
			return
		}
		driveName = keys[0]
		if driveName == "" {
			h.logger.Errorf("dir: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "dir: \"podName\" argument missing"})
			return
		}
	}
	keys, ok = r.URL.Query()["dirPath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("dir present: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir present: \"dirPath\" argument missing"})
		return
	}
	dir := keys[0]

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

	// stat directory
	ds, err := h.dfsAPI.DirectoryStat(driveName, dir, sessionId, isGroup)
	if err != nil {
		if errors.Is(err, dfs.ErrPodNotOpen) || errors.Is(err, dfs.ErrUserNotLoggedIn) ||
			errors.Is(err, p.ErrPodNotOpened) {
			h.logger.Errorf("dir stat: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "dir stat: " + err.Error()})
			return
		}
		h.logger.Errorf("dir stat: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "dir stat: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, ds)
}
