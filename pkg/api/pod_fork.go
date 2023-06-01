package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// PodForkRequest is the api request to fork a pod
type PodForkRequest struct {
	PodName  string `json:"podName,omitempty"`
	ForkName string `json:"forkName,omitempty"`
}

// PodForkFromReferenceRequest is the api request to fork a pod from a reference
type PodForkFromReferenceRequest struct {
	ForkName  string `json:"forkName,omitempty"`
	Reference string `json:"podSharingReference"`
}

// PodForkHandler godoc
//
//	@Summary      Fork a pod
//	@Description  PodForkHandler is the api handler to fork a pod
//	@ID           pod-fork-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodForkRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/fork [post]
func (h *Handler) PodForkHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod fork: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq PodForkRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod fork: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: could not decode arguments"})
		return
	}

	pod := podReq.PodName
	if pod == "" {
		h.logger.Errorf("pod fork: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: \"podName\" argument missing"})
		return
	}

	forkName := podReq.ForkName
	if forkName == "" {
		h.logger.Errorf("pod fork: \"forkName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: \"forkName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod fork: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod fork: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "pod stat: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.ForkPod(pod, forkName, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod fork: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "pod fork: " + err.Error()})
			return
		}
		h.logger.Errorf("pod fork: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod fork: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &response{Message: "pod forked successfully"})
}

// PodForkFromReferenceHandler godoc
//
//	@Summary      Fork a pod from sharing reference
//	@Description  PodForkFromReferenceHandler is the api handler to fork a pod from a given sharing reference
//	@ID           pod-fork-from-reference-handler
//	@Tags         pod
//	@Accept       json
//	@Produce      json
//	@Param	      pod_request body PodForkFromReferenceRequest true "pod name and user password"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/pod/fork-from-reference [post]
func (h *Handler) PodForkFromReferenceHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("pod fork: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var podReq PodForkFromReferenceRequest
	err := decoder.Decode(&podReq)
	if err != nil {
		h.logger.Errorf("pod fork: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "pod fork: could not decode arguments"})
		return
	}

	forkName := podReq.ForkName

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("pod fork: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("pod fork: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "pod stat: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.ForkPodFromRef(forkName, podReq.Reference, sessionId)
	if err != nil {
		if err == dfs.ErrUserNotLoggedIn ||
			err == p.ErrInvalidPodName {
			h.logger.Errorf("pod fork: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "pod fork: " + err.Error()})
			return
		}
		h.logger.Errorf("pod fork: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "pod fork: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &response{Message: "pod forked successfully"})
}
