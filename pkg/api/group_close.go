package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupCloseHandler godoc
//
//	@Summary      Close group
//	@Description  GroupCloseHandler is the api handler to close a group
//	@ID           group-close-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      groupRequest body GroupNameRequest true "group name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/close [post]
func (h *Handler) GroupCloseHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group close: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group close: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupNameRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group close: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group close: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group close: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group close: \"groupName\" argument missing"})
		return
	}

	// get sessionId from request
	sessionId, err := auth.GetSessionIdFromRequest(r)
	if err != nil {
		h.logger.Errorf("sessionId parse failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Error("sessionId not set: ", err)
		jsonhttp.BadRequest(w, &response{Message: ErrUnauthorized.Error()})
		return
	}

	err = h.dfsAPI.CloseGroup(group, sessionId)
	if err != nil {
		h.logger.Errorf("group close: ", err)
		jsonhttp.InternalServerError(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "group closed"})
}
