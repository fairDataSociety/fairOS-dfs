package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupAddMemberRequest is the request to add a member to a group
type GroupAddMemberRequest struct {
	GroupName  string `json:"groupName,omitempty"`
	Member     string `json:"member,omitempty"`
	Permission uint8  `json:"permission,omitempty"`
}

// GroupAddMemberResponse is the response to add a member to a group
type GroupAddMemberResponse struct {
	Reference string `json:"invite,omitempty"`
}

// GroupAddMemberHandler godoc
//
//	@Summary      Add member to group
//	@Description  GroupAddMemberHandler is the api handler to add a member to a group
//	@ID           group-add-member-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupAddMemberRequest true "group name, member name and permission"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  GroupAddMemberResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/add [post]
func (h *Handler) GroupAddMemberHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group add: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group add: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupAddMemberRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group add: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group add: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group add: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group add: \"groupName\" argument missing"})
		return
	}

	member := req.Member
	if member == "" {
		h.logger.Errorf("group add: \"member\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group add: \"member\" argument missing"})
		return
	}

	permission := req.Permission
	if permission == 0 {
		h.logger.Errorf("group add: \"permission\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group add: \"permission\" argument missing"})
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

	ref, err := h.dfsAPI.AddMember(sessionId, group, member, permission)
	if err != nil {
		h.logger.Errorf("group add: ", err)
		jsonhttp.InternalServerError(w, &response{Message: "group add: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", " application/json")
	jsonhttp.OK(w, &GroupAddMemberResponse{
		Reference: string(ref),
	})
}
