package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"

	"resenje.org/jsonhttp"
)

// PodPresentHandler is the api handler to check if a pod is present
// it takes pod_name as query parameter
func (h *Handler) PodPresentHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc ls: \"pod_name\" argument missing")
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc ls: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "doc ls: \"pod_name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod open: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod open: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "pod open: \"cookie-id\" parameter missing in cookie")
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
