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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"resenje.org/jsonhttp"
)

// FileDownloadHandlerPost godoc
//
//	@Summary      Download a file
//	@Description  FileDownloadHandlerPost is the api handler to download a file from a given pod
//	@Tags         file
//	@Accept       mpfd
//	@Produce      */*
//	@Param	      pod_name formData string true "pod name"
//	@Param	      file_path formData string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {array}  byte
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/download [post]
func (h *Handler) FileDownloadHandlerPost(w http.ResponseWriter, r *http.Request) {
	podName := r.FormValue("pod_name")
	if podName == "" {
		h.logger.Errorf("download: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"pod_name\" argument missing"})
		return
	}

	podFileWithPath := r.FormValue("file_path")
	if podFileWithPath == "" {
		h.logger.Errorf("download: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"file_path\" argument missing"})
		return
	}

	h.handleDownload(w, r, podName, podFileWithPath)

}

// FileDownloadHandlerGet godoc
//
//	@Summary      Download a file
//	@Description  FileDownloadHandlerGet is the api handler to download a file from a given pod
//	@Tags         file
//	@Accept       json
//	@Produce      */*
//	@Param	      pod_name query string true "pod name"
//	@Param	      file_path query string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {array}  byte
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/download [get]
func (h *Handler) FileDownloadHandlerGet(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("download \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("download: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"pod_name\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["file_path"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("download: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"file_path\" argument missing"})
		return
	}
	podFileWithPath := keys[0]
	if podFileWithPath == "" {
		h.logger.Errorf("download: \"file_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"file_path\" argument missing"})
		return
	}

	h.handleDownload(w, r, podName, podFileWithPath)
}

func (h *Handler) handleDownload(w http.ResponseWriter, r *http.Request, podName, podFileWithPath string) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("download: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("download: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "download: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// download file from bee
	reader, size, err := h.dfsAPI.DownloadFile(podName, podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("download: %v", err)
			jsonhttp.BadRequest(w, "download: "+err.Error())
			return
		}
		if err == file.ErrFileNotPresent || err == file.ErrFileNotFound {
			h.logger.Errorf("download: %v", err)
			jsonhttp.NotFound(w, "download: "+err.Error())
			return
		}
		h.logger.Errorf("download: %v", err)
		jsonhttp.InternalServerError(w, "download: "+err.Error())
		return
	}
	defer reader.Close()
	sizeString := strconv.FormatUint(size, 10)
	w.Header().Set("Content-Length", sizeString)

	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "download: "+err.Error())
	}
}
