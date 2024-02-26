package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupDeleteSharedHandler godoc
//
//	@Summary      Delete shared group
//	@Description  GroupDeleteSharedHandler is the api handler to delete a shared group
//	@ID           group-delete-shared-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupNameRequest true "group name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/delete-shared [delete]
func (h *Handler) GroupDeleteSharedHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group delete shared: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group delete shared: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupNameRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group delete shared: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group delete shared: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group delete shared: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group delete shared: \"groupName\" argument missing"})
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

	err = h.dfsAPI.RemoveSharedGroup(sessionId, group)
	if err != nil {
		h.logger.Errorf("group delete shared: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group delete shared: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "group deleted successfully"})
}
