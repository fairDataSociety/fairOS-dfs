package api

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/gorilla/mux"
	"resenje.org/jsonhttp"
)

// PublicPodGetFileHandler godoc
//
//	@Summary      download file from a shared pod
//	@Description  PodReceiveInfoHandler is the api handler to download file from a shared pod
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      filePath query string true "file location in the pod"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public-file [get]
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
//	@Summary      download file from a shared pod
//	@Description  PublicPodFilePathHandler is the api handler to download file from a shared pod
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public/{ref}/{file} [get]
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
	filePath := params["file"]
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
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	reader, size, err := h.dfsAPI.PublicPodFileDownload(shareInfo, filePath)
	if err != nil {
		h.logger.Errorf("public pod file download: %v", err)
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
		return
	}

	defer reader.Close()
	w.Header().Set("Content-Length", strconv.Itoa(int(size)))
	w.Header().Set("Content-Type", contentType)
	if strings.HasPrefix(filePath, "static/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		h.logger.Errorf("download: %v", err)
		w.Header().Set("Content-Type", " application/json")
		jsonhttp.InternalServerError(w, "public pod file download: "+err.Error())
	}
}

// PublicPodGetDirHandler godoc
//
//	@Summary      List directory content
//	@Description  PublicPodGetDirHandler is the api handler to list content of a directory from a public pod
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      dirPath query string true "dir location in the pod"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public-dir [get]
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

// PublicPodKVEntryGetHandler godoc
//
//	@Summary      get key from public pod
//	@Description  PublicPodKVEntryGetHandler is the api handler to get key from key value store from a public pod
//	@Tags         public
//	@Accept       json
//	@Produce      json
//	@Param	      sharingRef query string true "pod sharing reference"
//	@Param	      tableName query string true "table name"
//	@Param	      key query string true "key to look up"
//	@Success      200  {object}  pod.ShareInfo
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /public-kv [get]
func (h *Handler) PublicPodKVEntryGetHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sharingRef"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod kv get: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod kv get: \"sharingRef\" argument missing")
		return
	}

	sharingRefString := keys[0]
	if sharingRefString == "" {
		h.logger.Errorf("public pod kv get: \"sharingRef\" argument missing")
		jsonhttp.BadRequest(w, "public pod kv get: \"sharingRef\" argument missing")
		return
	}

	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		h.logger.Errorf("public pod kv get: invalid reference: ", err)
		jsonhttp.BadRequest(w, "public pod kv get: invalid reference:"+err.Error())
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "public pod kv get: \"tableName\" argument missing"})
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("public pod kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "public pod kv get: \"tableName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("public pod kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "public pod kv get: \"key\" argument missing"})
		return
	}
	key := keys[0]
	if key == "" {
		h.logger.Errorf("public pod kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "public pod kv get: \"key\" argument missing"})
		return
	}

	shareInfo, err := h.dfsAPI.PublicPodReceiveInfo(ref)
	if err != nil {
		h.logger.Errorf("public pod kv get: %v", err)
		jsonhttp.InternalServerError(w, "public pod kv get: "+err.Error())
		return
	}

	columns, data, err := h.dfsAPI.PublicPodKVEntryGet(shareInfo, name, key)
	if err != nil {
		h.logger.Errorf("public pod kv get: %v", err)
		if err == collection.ErrEntryNotFound {
			jsonhttp.NotFound(w, &response{Message: "public pod kv get: " + err.Error()})
			return
		}
		jsonhttp.InternalServerError(w, &response{Message: "public pod kv get: " + err.Error()})
		return
	}

	var resp KVResponse
	if columns != nil {
		resp.Keys = columns
	} else {
		resp.Keys = []string{key}
	}
	resp.Values = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &resp)
}
