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
	"encoding/json"
	"net/http"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
)

// ReceiveFileResponse represents the response for receiving a file
type ReceiveFileResponse struct {
	FileName string `json:"fileName"`
}

// FileSharingReference represents a file reference
type FileSharingReference struct {
	Reference string `json:"fileSharingReference"`
}

// FileShareRequest is the request to share a file
type FileShareRequest struct {
	PodName     string `json:"podName,omitempty"`
	FilePath    string `json:"filePath,omitempty"`
	Destination string `json:"destUser,omitempty"`
}

// FileShareHandler godoc
//
//	@Summary      Share a file
//	@Description  FileShareHandler is the api handler to share a file from a given pod
//	@ID		      file-share-handler
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      file_share_request body FileShareRequest true "file share request params"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  FileSharingReference
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/share [post]
func (h *Handler) FileShareHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file share: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file share: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq FileShareRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("file share: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file share: could not decode arguments"})
		return
	}

	podName := fsReq.PodName
	if podName == "" {
		h.logger.Errorf("file share: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"podName\" argument missing"})
		return
	}

	podFileWithPath := fsReq.FilePath
	if podFileWithPath == "" {
		h.logger.Errorf("file share: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"filePath\" argument missing"})
		return
	}
	destinationRef := fsReq.Destination
	if destinationRef == "" {
		h.logger.Errorf("file share: \"destUser\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"destUser\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file share: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file share: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"cookie-id\" parameter missing in cookie"})
		return
	}

	sharingRef, err := h.dfsAPI.ShareFile(podName, podFileWithPath, destinationRef, sessionId)
	if err != nil {
		h.logger.Errorf("file share: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file share: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &FileSharingReference{
		Reference: sharingRef,
	})
}

// FileReceiveHandler godoc
//
//	@Summary      Receive a file
//	@Description  FileReceiveHandler is the api handler to receive a file in a given pod
//	@ID		      file-receive-handler
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      sharingRef query string true "sharing reference"
//	@Param	      dirPath query string true "file location"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  FileSharingReference
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/receive [get]
func (h *Handler) FileReceiveHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("file receive: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"sharingRef\" argument missing"})
		return
	}
	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("file receive: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"sharingRef\" argument missing"})
		return
	}

	keys1, ok1 := r.URL.Query()["dirPath"]
	if !ok1 || len(keys1[0]) < 1 || keys1[0] == "" {
		h.logger.Errorf("file receive: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"dirPath\" argument missing"})
		return
	}
	dir := keys1[0]

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file receive: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file receive: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"cookie-id\" parameter missing in cookie"})
		return
	}

	filePath, err := h.dfsAPI.ReceiveFile(podName, sessionId, sharingRefString, dir)
	if err != nil {
		h.logger.Errorf("file receive: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file receive: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &ReceiveFileResponse{
		FileName: filePath,
	})
}

// FileReceiveInfoHandler godoc
//
//	@Summary      Receive a file info
//	@Description  FileReceiveInfoHandler is the api handler to receive a file info
//	@ID		      file-receive-info-handler
//	@Tags         file
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "sharing reference"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  user.ReceiveFileInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/receiveinfo [get]
func (h *Handler) FileReceiveInfoHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive info: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"sharingRef\" argument missing"})
		return
	}
	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("file receive info: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"sharingRef\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file receive info: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file receive info: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"cookie-id\" parameter missing in cookie"})
		return
	}

	receiveInfo, err := h.dfsAPI.ReceiveInfo(sessionId, sharingRefString)
	if err != nil {
		h.logger.Errorf("file receive info: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file receive info: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, receiveInfo)
}
