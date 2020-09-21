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

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

type ListFileResponse struct {
	Entries []dir.DirOrFileEntry `json:"entries"`
}

type DirOrFileEntry struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Size             string `json:"size,omitempty"`
	CreationTime     string `json:"creation_time"`
	ModificationTime string `json:"modification_time"`
	AccessTime       string `json:"access_time"`
}

func (h *Handler) DirectoryLsHandler(w http.ResponseWriter, r *http.Request) {
	directory := r.FormValue("dir")
	if directory == "" {
		h.logger.Errorf("ls: \"dir\" argument missing")
		jsonhttp.BadRequest(w, "ls: \"dir\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("ls: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("ls: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "ls: \"cookie-id\" parameter missing in cookie")
		return
	}

	// list directory
	entries, err := h.dfsAPI.ListDir(directory, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("ls: %v", err)
			jsonhttp.BadRequest(w, "ls: "+err.Error())
			return
		}
		h.logger.Errorf("ls: %v", err)
		jsonhttp.InternalServerError(w, "ls: "+err.Error())
		return
	}

	if entries == nil {
		entries = make([]dir.DirOrFileEntry, 0)
	}
	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &ListFileResponse{
		Entries: entries,
	})
}
