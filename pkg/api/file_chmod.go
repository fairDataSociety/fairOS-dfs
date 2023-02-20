package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// FileModeRequest is used to change file permission mode
type FileModeRequest struct {
	PodName  string `json:"podName,omitempty"`
	FilePath string `json:"filePath,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// FileModeHandler godoc
//
//	@Summary      chmod a file
//	@Description  FileModeHandler is the api handler to change mode of a file
//	@Tags         file
//	@Accept       mpfd
//	@Produce      json
//	@Param	      file_request body FileModeRequest true "file mode request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/chmod [Post]
func (h *Handler) FileModeHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("file chmod: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var chmodReq FileModeRequest
	err := decoder.Decode(&chmodReq)
	if err != nil {
		h.logger.Errorf("file chmod: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: could not decode arguments"})
		return
	}

	podName := chmodReq.PodName
	if podName == "" {
		h.logger.Errorf("file chmod: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: \"podName\" argument missing"})
		return
	}

	filePath := chmodReq.FilePath
	if filePath == "" {
		h.logger.Errorf("file chmod: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: \"filePath\" argument missing"})
		return
	}

	modeStr := chmodReq.Mode
	if modeStr == "" {
		h.logger.Errorf("file chmod: \"mode\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: \"mode\" argument missing"})
		return
	}

	mode, err := strconv.ParseUint(modeStr, 10, 32)
	if err != nil {
		h.logger.Errorf("file chmod: invalid mode: %v", err)
		jsonhttp.BadRequest(w, &response{Message: fmt.Sprintf("file chmod: invalid mode: %v", err)})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("file chmod: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("file chmod: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "file chmod: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.ChmodFile(podName, filePath, sessionId, uint32(mode))
	if err != nil {
		h.logger.Errorf("file chmod: %v", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &response{Message: "file mode changed successfully"})
}
