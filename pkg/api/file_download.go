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
	"io"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
)

// FileDownloadHandler is the api handler to download a file from a given pod
//  it takes only one argument
// file_path: the absolute path of the file in the pod
func (h *Handler) FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	podFileWithPath := r.FormValue("file_path")
	if podFileWithPath == "" {
		h.logger.Errorf("download: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, "download: \"file_path\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("download: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("download: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "download: \"cookie-id\" parameter missing in cookie")
		return
	}

	// download file from bee
	reader, size, err := h.dfsAPI.DownloadFile(podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("download: %v", err)
			jsonhttp.BadRequest(w, "download: "+err.Error())
			return
		}
		h.logger.Errorf("download: %v", err)
		jsonhttp.InternalServerError(w, "download: "+err.Error())
		return
	}

	sizeString := strconv.FormatUint(size, 10)
	w.Header().Set("Content-Length", sizeString)

	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "stat dir: "+err.Error())
	}
	_ = reader.Close()
}
