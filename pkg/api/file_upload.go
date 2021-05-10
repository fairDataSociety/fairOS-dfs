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

	"github.com/dustin/go-humanize"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"resenje.org/jsonhttp"
)

type UploadFileResponse struct {
	Responses []Response
}

type Response struct {
	FileName string `json:"file_name"`
	Message  string `json:"message,omitempty"`
}

const (
	defaultMaxMemory  = 32 << 20 // 32 MB
	CompressionHeader = "fairOS-dfs-Compression"
)

func (h *Handler) FileUploadHandler(w http.ResponseWriter, r *http.Request) {
	podPath := r.FormValue("dir_path")
	if podPath == "" {
		h.logger.Errorf("file upload: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, "file upload: \"dir_path\" argument missing")
		return
	}

	blockSize := r.FormValue("block_size")
	if blockSize == "" {
		h.logger.Errorf("file upload: \"block_size\" argument missing")
		jsonhttp.BadRequest(w, "file upload: \"block_size\" argument missing")
		return
	}

	compression := r.Header.Get(CompressionHeader)
	if compression != "" {
		if compression != "snappy" && compression != "gzip" {
			h.logger.Errorf("file upload: invalid value for \"compression\" header")
			jsonhttp.BadRequest(w, "file upload: invalid value for \"compression\" header")
			return
		}
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file upload: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file upload: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "file upload: \"cookie-id\" parameter missing in cookie")
		return
	}

	//  get the files parameter from the multi part
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("file upload: %v", err)
		jsonhttp.BadRequest(w, "file upload: "+err.Error())
		return
	}

	bs, err := humanize.ParseBytes(blockSize)
	if err != nil {
		h.logger.Errorf("file upload: %v", err)
		jsonhttp.BadRequest(w, "file upload: "+err.Error())
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		h.logger.Errorf("file upload: parameter \"files\" missing")
		jsonhttp.BadRequest(w, "file upload: parameter \"files\" missing")
		return
	}

	// upload files one by one
	var responses []Response
	for _, file := range files {
		fd, err := file.Open()
		defer func() {
			err := fd.Close()
			if err != nil {
				h.logger.Errorf("file upload: error closing file: %v", err)
			}
		}()
		if err != nil {
			h.logger.Errorf("file upload: %v", err)
			responses = append(responses, Response{FileName: file.Filename, Message: err.Error()})
			continue
		}

		//upload file to bee
		err = h.dfsAPI.UploadFile(file.Filename, sessionId, file.Size, fd, podPath, compression, uint32(bs))
		if err != nil {
			if err == dfs.ErrPodNotOpen {
				h.logger.Errorf("file upload: %v", err)
				jsonhttp.BadRequest(w, "file upload: "+err.Error())
				return
			}
			h.logger.Errorf("file upload: %v", err)
			responses = append(responses, Response{FileName: file.Filename, Message: err.Error()})
			continue
		}
		responses = append(responses, Response{FileName: file.Filename, Message: "uploaded succesfully"})
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &UploadFileResponse{
		Responses: responses,
	})
}
