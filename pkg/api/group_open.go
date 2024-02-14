package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupOpenHandler godoc
//
//	@Summary      Open group
//	@Description  GroupOpenHandler is the api handler to open a group
//	@ID 		  group-open
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupNameRequest true "group name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/open [post]
func (h *Handler) GroupOpenHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group open: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group open: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupNameRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group open: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group open: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group open: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group open: \"groupName\" argument missing"})
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

	_, err = h.dfsAPI.OpenGroup(group, sessionId)
	if err != nil {
		h.logger.Errorf("group open failed: ", err)
		jsonhttp.InternalServerError(w, &response{Message: "group open failed"})
		return
	}

	jsonhttp.OK(w, &response{Message: "group open successful"})
}
