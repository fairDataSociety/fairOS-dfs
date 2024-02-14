package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupRemoveMemberRequest is the request to remove a member from a group
type GroupRemoveMemberRequest struct {
	GroupName string `json:"groupName,omitempty"`
	Member    string `json:"member,omitempty"`
}

// GroupRemoveMemberHandler godoc
//
//	@Summary      Remove member from group
//	@Description  GroupRemoveMemberHandler is the api handler to remove a member from a group
//	@ID           group-remove-member-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupRemoveMemberRequest true "group name and member name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/remove [post]
func (h *Handler) GroupRemoveMemberHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group remove: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group remove: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupRemoveMemberRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group remove: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group remove: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group remove: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group remove: \"groupName\" argument missing"})
		return
	}

	member := req.Member
	if member == "" {
		h.logger.Errorf("group remove: \"member\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group remove: \"member\" argument missing"})
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

	err = h.dfsAPI.RemoveMember(group, member, sessionId)
	if err != nil {
		h.logger.Errorf("group remove: failed to remove member from group: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group remove: failed to remove member from group"})
		return
	}

	jsonhttp.OK(w, &response{Message: "member removed from group"})
}
