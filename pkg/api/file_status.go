package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"resenje.org/jsonhttp"
)

// StatusResponse is the response for file status
type StatusResponse struct {
	Total     int64 `json:"total"`
	Processed int64 `json:"processed"`
	Synced    int64 `json:"synced"`
}

// FileStatusHandler godoc
//
//	@Summary      Sync status of a file
//	@Description  FileStatusHandler is the api handler to check sync status of a file from a given pod
//	@ID		      file-status-handler
//	@Tags         file
//	@Accept       json
//	@Produce      */*
//	@Param	      podName query string true "pod name"
//	@Param	      filePath query string true "file path"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {array}  StatusResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/file/status [get]
func (h *Handler) FileStatusHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("status \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "status: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("status: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "status: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["filePath"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("status: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "status: \"filePath\" argument missing"})
		return
	}
	podFileWithPath := keys[0]
	if podFileWithPath == "" {
		h.logger.Errorf("status: \"filePath\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "status: \"filePath\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("status: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("status: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "status: \"cookie-id\" parameter missing in cookie"})
		return
	}

	// status of file
	t, p, s, err := h.dfsAPI.StatusFile(podName, podFileWithPath, sessionId)
	if err != nil {
		if err == dfs.ErrPodNotOpen {
			h.logger.Errorf("status: %v", err)
			jsonhttp.BadRequest(w, "status: "+err.Error())
			return
		}
		if err == file.ErrFileNotFound {
			h.logger.Errorf("status: %v", err)
			jsonhttp.NotFound(w, "status: "+err.Error())
			return
		}
		h.logger.Errorf("status: %v", err)
		jsonhttp.InternalServerError(w, "status: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &StatusResponse{
		Total:     t,
		Processed: p,
		Synced:    s,
	})
}
