package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DirModeRequest is used for changing dir mode
type DirModeRequest struct {
	PodName string `json:"podName,omitempty"`
	DirPath string `json:"dirPath,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

// DirectoryModeHandler godoc
//
//	@Summary      change mode of a directory
//	@Description  DirectoryModeHandler is the api handler to change mode of a directory
//	@ID		      directory-mode-handler
//	@Tags         dir
//	@Accept       json
//	@Produce      json
//	@Param	      dir_request body DirModeRequest true "dir mode request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/dir/chmod [post]
func (h *Handler) DirectoryModeHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("dir chmod: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var chmodReq DirModeRequest
	err := decoder.Decode(&chmodReq)
	if err != nil {
		h.logger.Errorf("dir chmod: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: could not decode arguments"})
		return
	}

	podName := chmodReq.PodName
	if podName == "" {
		h.logger.Errorf("dir chmod: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: \"podName\" argument missing"})
		return
	}

	dirPath := chmodReq.DirPath
	if dirPath == "" {
		h.logger.Errorf("dir chmod: \"dirPath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: \"dirPath\" argument missing"})
		return
	}

	modeStr := chmodReq.Mode
	if modeStr == "" {
		h.logger.Errorf("dir chmod: \"mode\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: \"mode\" argument missing"})
		return
	}

	mode, err := strconv.ParseUint(modeStr, 10, 32)
	if err != nil {
		h.logger.Errorf("dir chmod: invalid mode: %v", err)
		jsonhttp.BadRequest(w, &response{Message: fmt.Sprintf("dir chmod: invalid mode: %v", err)})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("dir chmod: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("dir chmod: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "dir chmod: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.ChmodDir(podName, dirPath, sessionId, uint32(mode))
	if err != nil {
		h.logger.Errorf("dir chmod: %v", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "directory renamed successfully"})
}
