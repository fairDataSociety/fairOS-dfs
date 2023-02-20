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
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"resenje.org/jsonhttp"
)

// UploadFileResponse
type UploadFileResponse struct {
	Responses []UploadResponse
}

// UploadResponse
type UploadResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message,omitempty"`
}

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	//CompressionHeader
	CompressionHeader = "fairOS-dfs-Compression"
)

// FileUploadHandler godoc
//
//	@Summary      Upload a file
//	@Description  FileUploadHandler is the api handler to upload a file from a local file system to the dfs
//	@Tags         file
//	@Accept       mpfd
//	@Produce      json
//	@Param	      podName formData string true "pod name"
//	@Param	      dirPath formData string true "location"
//	@Param	      blockSize formData string true "block size to break the file" example(4Kb, 1Mb)
//	@Param	      files formData file true "file to upload"
//	@Param	      fairOS-dfs-Compression header string false "cookie parameter" example(snappy, gzip)
//	@Param	      Cookie header string true "cookie parameter"
//	@Param	      overwrite formData string false "overwrite the file if already exists" example(true, false)
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/upload [Post]
func (h *Handler) FileUploadHandler(w http.ResponseWriter, r *http.Request) {
	podName := r.FormValue("podName")
	if podName == "" {
		h.logger.Errorf("file upload: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file upload: \"podName\" argument missing"})
		return
	}

	podPath := r.FormValue("dirPath")
	if podPath == "" {
		h.logger.Errorf("file upload: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file upload: \"dirPath\" argument missing"})
		return
	}

	blockSize := r.FormValue("blockSize")
	if blockSize == "" {
		h.logger.Errorf("file upload: \"blockSize\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file upload: \"blockSize\" argument missing"})
		return
	}

	compression := r.Header.Get(CompressionHeader)
	if compression != "" {
		if compression != "snappy" && compression != "gzip" {
			h.logger.Errorf("file upload: invalid value for \"compression\" header")
			jsonhttp.BadRequest(w, &response{Message: "file upload: invalid value for \"compression\" header"})
			return
		}
	}
	var err error
	overwrite := true
	overwriteString := r.FormValue("overwrite")
	if overwriteString != "" {
		overwrite, err = strconv.ParseBool(overwriteString)
		if err != nil {
			h.logger.Errorf("file upload: \"overwrite\" argument is wrong")
			jsonhttp.BadRequest(w, &response{Message: "file upload: \"overwrite\" argument is wrong"})
			return
		}
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file upload: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file upload: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file upload: \"cookie-id\" parameter missing in cookie"})
		return
	}

	//  get the files parameter from the multipart
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("file upload: %v", err)
		jsonhttp.BadRequest(w, &response{Message: "file upload: " + err.Error()})
		return
	}

	bs, err := humanize.ParseBytes(blockSize)
	if err != nil {
		h.logger.Errorf("file upload: %v", err)
		jsonhttp.BadRequest(w, &response{Message: "file upload: " + err.Error()})
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		h.logger.Errorf("file upload: parameter \"files\" missing")
		jsonhttp.BadRequest(w, &response{Message: "file upload: parameter \"files\" missing"})
		return
	}

	// upload files one by one
	var responses []UploadResponse
	for _, file := range files {
		fd, err := file.Open()
		if err != nil {
			h.logger.Errorf("file upload: %v", err)
			responses = append(responses, UploadResponse{FileName: file.Filename, Message: err.Error()})
			continue
		}
		err = h.handleFileUpload(podName, file.Filename, sessionId, file.Size, fd, podPath, compression, uint32(bs), overwrite)
		if err != nil {
			if err == dfs.ErrPodNotOpen {
				h.logger.Errorf("file upload: %v", err)
				jsonhttp.BadRequest(w, &response{Message: "file upload: " + err.Error()})
				return
			}
			h.logger.Errorf("file upload: %v", err)
			responses = append(responses, UploadResponse{FileName: file.Filename, Message: err.Error()})
			continue
		}
		responses = append(responses, UploadResponse{FileName: file.Filename, Message: "uploaded successfully"})
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &UploadFileResponse{
		Responses: responses,
	})
}

func (h *Handler) handleFileUpload(podName, podFileName, sessionId string, fileSize int64, f multipart.File, podPath, compression string, blockSize uint32, overwrite bool) error {
	defer f.Close()
	return h.dfsAPI.UploadFile(podName, podFileName, sessionId, fileSize, f, podPath, compression, blockSize, overwrite)
}
