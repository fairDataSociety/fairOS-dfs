package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupUpdatePermissionHandler godoc
//
//	@Summary      Update group permission
//	@Description  GroupUpdatePermissionHandler is the api handler to update a group permission
//	@ID           group-update-permission-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupAddMemberRequest true "group name, member name and permission"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/update-permission [post]
func (h *Handler) GroupUpdatePermissionHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group update permission: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group update permission: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupAddMemberRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group update permission: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group update permission: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group update permission: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group update permission: \"groupName\" argument missing"})
		return
	}

	member := req.Member
	if member == "" {
		h.logger.Errorf("group update permission: \"member\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group update permission: \"member\" argument missing"})
		return
	}

	permission := req.Permission
	if permission == 0 {
		h.logger.Errorf("group update permission: \"permission\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group update permission: \"permission\" argument missing"})
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

	err = h.dfsAPI.UpdatePermission(sessionId, group, member, permission)
	if err != nil {
		h.logger.Errorf("group update permission: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "group permission updated successfully"})
}
