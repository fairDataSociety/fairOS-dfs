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
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type ReceiveFileResponse struct {
	FileName  string `json:"file_name"`
}

type FileSharingReference struct {
	Reference string `json:"file_sharing_reference"`
}

func (h *Handler) FileShareHandler(w http.ResponseWriter, r *http.Request) {
	podFileWithPath := r.FormValue("pod_path_file")
	if podFileWithPath == "" {
		h.logger.Errorf("file share: \"pod_path_file\" argument missing")
		jsonhttp.BadRequest(w, "file share: \"pod_path_file\" argument missing")
		return
	}
	destinationRef := r.FormValue("to")
	if destinationRef == "" {
		h.logger.Errorf("file share: \"to\" argument missing")
		jsonhttp.BadRequest(w, "file share: \"to\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file share: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file share: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "file share: \"cookie-id\" parameter missing in cookie")
		return
	}

	sharingRef, err := h.dfsAPI.ShareFile(podFileWithPath, destinationRef, sessionId)
	if err != nil {
		h.logger.Errorf("file share: %v", err)
		jsonhttp.InternalServerError(w, "file share: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &FileSharingReference{
		Reference: sharingRef,
	})
}

func (h *Handler) FileReceiveHandler(w http.ResponseWriter, r *http.Request) {
	sharingRefString := r.FormValue("ref")
	if sharingRefString == "" {
		h.logger.Errorf("file receive: \"ref\" argument missing")
		jsonhttp.BadRequest(w, "file receive: \"ref\" argument missing")
		return
	}

	dir := r.FormValue("dir")
	if dir == "" {
		h.logger.Errorf("file receive: \"dir\" argument missing")
		jsonhttp.BadRequest(w, "file receive: \"dir\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file receive: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file receive: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "file receive: \"cookie-id\" parameter missing in cookie")
		return
	}

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("file receive: invalid reference: ", err)
		jsonhttp.BadRequest(w, "file receive: invalid reference:"+err.Error())
		return
	}

	filePath, err := h.dfsAPI.ReceiveFile(sessionId, sharingRef, dir)
	if err != nil {
		h.logger.Errorf("file receive: %v", err)
		jsonhttp.InternalServerError(w, "file receive: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &ReceiveFileResponse{
		FileName:  filePath,
	})
}

func (h *Handler) FileReceiveInfoHandler(w http.ResponseWriter, r *http.Request) {
	sharingRefString := r.FormValue("ref")
	if sharingRefString == "" {
		h.logger.Errorf("file receive info: \"ref\" argument missing")
		jsonhttp.BadRequest(w, "file receive info: \"ref\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file receive info: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file receive info: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "file receive info: \"cookie-id\" parameter missing in cookie")
		return
	}

	sharingRef, err := utils.ParseSharingReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("file receive info: invalid reference: ", err)
		jsonhttp.BadRequest(w, "file receive info: invalid reference:"+err.Error())
		return
	}

	receiveInfo, err := h.dfsAPI.ReceiveInfo(sessionId, sharingRef)
	if err != nil {
		h.logger.Errorf("file receive info: %v", err)
		jsonhttp.InternalServerError(w, "file receive info: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, receiveInfo)
}
