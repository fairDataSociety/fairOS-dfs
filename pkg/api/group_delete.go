package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupDeleteHandler godoc
//
//	@Summary      Delete group
//	@Description  GroupDeleteHandler is the api handler to delete a new group
//	@ID           group-delete-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupNameRequest true "group name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/delete [delete]
func (h *Handler) GroupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group delete: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group delete: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupNameRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group delete: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group delete: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group delete: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group delete: \"groupName\" argument missing"})
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

	err = h.dfsAPI.RemoveGroup(sessionId, group)
	if err != nil {
		h.logger.Errorf("group delete: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group delete: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "group deleted successfully"})
}
