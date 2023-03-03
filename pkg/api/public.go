package api

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"resenje.org/jsonhttp"
)

// PublicPodGetFileHandler godoc
//
//	@Summary      Receive shared pod info
//	@Description  PodReceiveInfoHandler is the api handler to receive shared pod info from shared reference
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      filePath query string true "file location in the pod"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public [get]
func (h *Handler) PublicPodGetFileHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}

	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("public pod file download: invalid reference: ", err)
		jsonhttp.BadRequest(w, "public pod file download: invalid reference:"+err.Error())
		return
	}
	keys, ok = r.URL.Query()["filePath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod file download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"filePath\" argument missing")
		return
	}
	filePath := keys[0]
	if filePath == "" {
		h.logger.Errorf("public pod file download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"filePath\" argument missing")
		return
	}

	shareInfo, err := h.dfsAPI.PublicPodReceiveInfo(ref)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	reader, size, err := h.dfsAPI.PublicPodFileDownload(shareInfo, filePath)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	defer reader.Close()

	sizeString := strconv.FormatUint(size, 10)
	w.Header().Set("Content-Length", sizeString)

	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
	}
}

// PublicPodFilePathHandler godoc
//
//	@Summary      Receive shared pod info
//	@Description  PodReceiveInfoHandler is the api handler to receive shared pod info from shared reference
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      filePath query string true "file location in the pod"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public [get]
func (h *Handler) PublicPodFilePathHandler(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	sharingRefString, ok := params["ref"]
	if !ok {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}
	if sharingRefString == "" {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("public pod file download: invalid reference: ", err)
		jsonhttp.BadRequest(w, "public pod file download: invalid reference:"+err.Error())
		return
	}
	filePath, ok := params["file"]
	if !ok {
		h.logger.Errorf("public pod file download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"filePath\" argument missing")
		return
	}
	if filePath == "" {
		filePath = "/index.html"
	} else {
		filePath = "/" + filePath
	}

	shareInfo, err := h.dfsAPI.PublicPodReceiveInfo(ref)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	reader, size, err := h.dfsAPI.PublicPodFileDownload(shareInfo, filePath)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	defer reader.Close()

	sizeString := strconv.FormatUint(size, 10)
	w.Header().Set("Content-Length", sizeString)

	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
	}
}

// PublicPodGetDirHandler godoc
//
//	@Summary      Receive shared pod info
//	@Description  PodReceiveInfoHandler is the api handler to receive shared pod info from shared reference
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      dirPath query string true "dir location in the pod"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public [get]
func (h *Handler) PublicPodGetDirHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}

	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("public pod file download: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"sharingRef\" argument missing")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("public pod file download: invalid reference: ", err)
		jsonhttp.BadRequest(w, "public pod file download: invalid reference:"+err.Error())
		return
	}
	keys, ok = r.URL.Query()["dirPath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod file download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"filePath\" argument missing")
		return
	}
	dirPath := keys[0]
	if dirPath == "" {
		h.logger.Errorf("public pod file download: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, "public pod file download: \"filePath\" argument missing")
		return
	}

	shareInfo, err := h.dfsAPI.PublicPodReceiveInfo(ref)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	dEntries, fEntries, err := h.dfsAPI.PublicPodDisLs(shareInfo, dirPath)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	if dEntries == nil {
		dEntries = make([]dir.Entry, 0)
	}
	if fEntries == nil {
		fEntries = make([]file.Entry, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &ListFileResponse{
		Directories: dEntries,
		Files:       fEntries,
	})
}
