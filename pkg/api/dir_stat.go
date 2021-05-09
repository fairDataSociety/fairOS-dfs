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
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func (h *Handler) DirectoryStatHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["dir_path"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("dir present: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, "dir present: \"dir_path\" argument missing")
		return
	}
	dir := keys[0]

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("dir stat: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("dir stat: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "dir stat: \"cookie-id\" parameter missing in cookie")
		return
	}

	// stat directory
	ds, err := h.dfsAPI.DirectoryStat(dir, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen || err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrPodNotOpened {
			h.logger.Errorf("dir stat: %v", err)
			jsonhttp.BadRequest(w, "dir stat: "+err.Error())
			return
		}
		h.logger.Errorf("dir stat: %v", err)
		jsonhttp.InternalServerError(w, "dir stat: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, ds)
}
