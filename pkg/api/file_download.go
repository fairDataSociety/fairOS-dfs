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

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"resenje.org/jsonhttp"
)

// FileDownloadHandlerPost godoc
//
//	@Summary      Download a file
//	@Description  FileDownloadHandlerPost is the api handler to download a file from a given pod
//	@ID		      file-download-handler-post
//	@Tags         file
//	@Accept       mpfd
//	@Produce      */*
//	@Param	      podName formData string true "pod name"
//	@Param	      filePath formData string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {array}  byte
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/download [post]
func (h *Handler) FileDownloadHandlerPost(w http.ResponseWriter, r *http.Request) {
	driveName, isGroup := r.FormValue("groupName"), true
	if driveName == "" {
		isGroup = false
		driveName = r.FormValue("podName")
		if driveName == "" {
			h.logger.Errorf("download: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "download: \"podName\" argument missing"})
			return
		}
	}

	podFileWithPath := r.FormValue("filePath")
	if podFileWithPath == "" {
		h.logger.Errorf("download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"filePath\" argument missing"})
		return
	}

	h.handleDownload(w, r, driveName, podFileWithPath, isGroup)
}

// FileDownloadHandlerGet godoc
//
//	@Summary      Download a file
//	@Description  FileDownloadHandlerGet is the api handler to download a file from a given pod
//	@ID		      file-download-handler
//	@Tags         file
//	@Accept       json
//	@Produce      */*
//	@Param	      podName query string true "pod name"
//	@Param	      filePath query string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {array}  byte
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/download [get]
func (h *Handler) FileDownloadHandlerGet(w http.ResponseWriter, r *http.Request) {
	driveName, isGroup := "", false
	keys, ok := r.URL.Query()["groupName"]
	if ok || (len(keys) == 1 && len(keys[0]) > 0) {
		driveName = keys[0]
		isGroup = true
	} else {
		keys, ok := r.URL.Query()["podName"]
		if !ok || len(keys[0]) < 1 {
			h.logger.Errorf("download \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "download: \"podName\" argument missing"})
			return
		}
		driveName = keys[0]
		if driveName == "" {
			h.logger.Errorf("download: \"podName\" argument missing")
			jsonhttp.BadRequest(w, &response{Message: "download: \"podName\" argument missing"})
			return
		}
	}

	keys, ok = r.URL.Query()["filePath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"filePath\" argument missing"})
		return
	}
	podFileWithPath := keys[0]
	if podFileWithPath == "" {
		h.logger.Errorf("download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "download: \"filePath\" argument missing"})
		return
	}

	h.handleDownload(w, r, driveName, podFileWithPath, isGroup)
}

func (h *Handler) handleDownload(w http.ResponseWriter, r *http.Request, podName, podFileWithPath string, isGroup bool) {
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

	// download file from bee
	reader, size, err := h.dfsAPI.DownloadFile(podName, podFileWithPath, sessionId, false)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("download: %v", err)
			jsonhttp.BadRequest(w, "download: "+err.Error())
			return
		}
		if err == file.ErrFileNotFound {
			h.logger.Errorf("download: %v", err)
			jsonhttp.NotFound(w, "download: "+err.Error())
			return
		}
		h.logger.Errorf("download: %v", err)
		jsonhttp.InternalServerError(w, "download: "+err.Error())
		return
	}
	// skipcq: GO-S2307
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
