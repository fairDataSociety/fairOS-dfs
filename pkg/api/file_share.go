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

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"resenje.org/jsonhttp"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// ReceiveFileResponse represents the response for receiving a file
type ReceiveFileResponse struct {
	FileName string `json:"file_name"`
}

// FileSharingReference represents a file reference
type FileSharingReference struct {
	Reference string `json:"file_sharing_reference"`
}

// FileShareHandler is the api handler to share a file from a given pod
// it takes two arguments
// file_path: the absolute path of the file in the pod
// dest_user: the address of the destination user (this is not used now)
func (h *Handler) FileShareHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file share: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file share: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var fsReq common.FileSystemRequest
	err := decoder.Decode(&fsReq)
	if err != nil {
		h.logger.Errorf("file share: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file share: could not decode arguments"})
		return
	}

	podName := fsReq.PodName
	if podName == "" {
		h.logger.Errorf("file share: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"pod_name\" argument missing"})
		return
	}

	podFileWithPath := fsReq.FilePath
	if podFileWithPath == "" {
		h.logger.Errorf("file share: \"pod_path_file\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"pod_path_file\" argument missing"})
		return
	}
	destinationRef := fsReq.Destination
	if destinationRef == "" {
		h.logger.Errorf("file share: \"to\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file share: \"to\" argument missing"})
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

// FileReceiveHandler is the api handler to receive a file in a given pod
// it takes two arguments
// pod_name: the name of the pod
// sharing_ref: the sharing reference of a file
func (h *Handler) FileReceiveHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("file receive: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"pod_name\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["sharing_ref"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive: \"sharing_ref\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"sharing_ref\" argument missing"})
		return
	}
	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("file receive: \"ref\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"ref\" argument missing"})
		return
	}

	keys1, ok1 := r.URL.Query()["dir_path"]
	if !ok1 || len(keys1[0]) < 1 || keys1[0] == "" {
		h.logger.Errorf("file receive: \"dir_path\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive: \"dir_path\" argument missing"})
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

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("file receive: invalid reference: ", err)
		jsonhttp.BadRequest(w, &response{Message: "file receive: invalid reference:" + err.Error()})
		return
	}

	filePath, err := h.dfsAPI.ReceiveFile(podName, sessionId, sharingRef, dir)
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

// FileReceiveInfoHandler is the api handler to receive a file info
// it takes two arguments
// pod_name: the name of the pod
// sharing_ref: the sharing reference of a file
func (h *Handler) FileReceiveInfoHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive info: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"pod_name\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("file receive info: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"pod_name\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["sharing_ref"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("file receive info: \"sharing_ref\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"sharing_ref\" argument missing"})
		return
	}
	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("file receive info: \"ref\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file receive info: \"ref\" argument missing"})
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

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("file receive info: invalid reference: ", err)
		jsonhttp.BadRequest(w, &response{Message: "file receive info: invalid reference:" + err.Error()})
		return
	}

	receiveInfo, err := h.dfsAPI.ReceiveInfo(podName, sharingRef, sessionId)
	if err != nil {
		h.logger.Errorf("file receive info: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "file receive info: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, receiveInfo)
}
