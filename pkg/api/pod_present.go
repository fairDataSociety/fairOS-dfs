package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"

	"resenje.org/jsonhttp"
)

// PodPresentHandler godoc
//
//	@Summary      Is pod present
//	@Description  PodPresentHandler is the api handler to check if a pod is present
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/present [get]
func (h *Handler) PodPresentHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc ls: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc ls: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc ls: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc ls: \"podName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "pod open: \"cookie-id\" parameter missing in cookie"})
		return
	}
	if h.dfsAPI.IsPodExist(podName, sessionId) {
		jsonhttp.OK(w, &PresentResponse{
			Present: true,
		})
	} else {
		jsonhttp.OK(w, &PresentResponse{
			Present: false,
		})
	}
}
