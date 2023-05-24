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
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"

	"resenje.org/jsonhttp"
)

// FileUpdateHandler godoc
//
//	@Summary      Update a file
//	@Description  FileUpdateHandler is the api handler to update a file from a given offset
//	@ID		      file-update-handler
//	@Tags         file
//	@Accept       mpfd
//	@Produce      json
//	@Param	      podName formData string true "pod name"
//	@Param	      filePath formData string true "location"
//	@Param	      file formData file true "file content to update"
//	@Param	      offset formData string true "file offset"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/update [Post]
func (h *Handler) FileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file update: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file update: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file update: \"cookie-id\" parameter missing in cookie"})
		return
	}

	podName := r.FormValue("podName")
	if podName == "" {
		h.logger.Errorf("file update: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file update: \"podName\" argument missing"})
		return
	}

	fileNameWithPath := r.FormValue("filePath")
	if fileNameWithPath == "" {
		h.logger.Errorf("file update: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file update: \"filePath\" argument missing"})
		return
	}
	offset, err := strconv.ParseUint(r.FormValue("offset"), 10, 64)
	if err != nil {
		h.logger.Errorf("file update: \"offset\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file update: \"offset\" argument missing or wrong"})
		return
	}

	//  get the files parameter from the multipart
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("file update: %v", err)
		jsonhttp.BadRequest(w, &response{Message: "file update: " + err.Error()})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Errorf("file update: parameter \"file\" missing")
		jsonhttp.BadRequest(w, &response{Message: "file update: parameter \"file\" missing"})
		return
	}
	defer file.Close()

	_, err = h.dfsAPI.WriteAtFile(podName, fileNameWithPath, sessionId, file, offset, false)
	if err != nil {
		h.logger.Errorf("file update: writeAt failed: %s", err.Error())
		jsonhttp.BadRequest(w, &response{Message: "file update: writeAt failed: " + err.Error()})
		return
	}
	res := UploadResponse{
		FileName: fileNameWithPath,
		Message:  "updated successfully",
	}
	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &UploadFileResponse{
		Responses: []UploadResponse{res},
	})
}
