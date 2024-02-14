package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupMembersResponse is the response to get group members
type GroupMembersResponse struct {
	Members map[string]uint8 `json:"members,omitempty"`
}

// GroupGetMembers gets the members of a group
//
// @Summary      Get group members
// @Description  GroupGetMembers is the api handler to get the members of a group
// @ID           group-get-members
// @Tags         group
// @Accept       json
// @Produce      json
// @Param	     groupName query string true "group name"
// @Param	     Cookie header string true "cookie parameter"
// @Success      200  {object}  GroupMembersResponse
// @Failure      400  {object}  response
// @Failure      500  {object}  response
// @Router       /v1/group/members [get]
func (h *Handler) GroupGetMembers(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["groupName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("group members: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group members: \"groupName\" argument missing"})
		return
	}
	groupName := keys[0]
	if groupName == "" {
		h.logger.Errorf("group members: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group members: \"groupName\" argument missing"})
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

	members, err := h.dfsAPI.GetGroupMembers(sessionId, groupName)
	if err != nil {
		h.logger.Errorf("group members: failed to get group members: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group members: failed to get group members"})
		return
	}

	jsonhttp.OK(w, &GroupMembersResponse{
		Members: members,
	})
}
