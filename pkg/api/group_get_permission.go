package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupPermissionResponse represents the response of the group permission request
type GroupPermissionResponse struct {
	Permission uint8 `json:"permission"`
}

// GroupGetPermission returns the permission of the loggedin user in the group
//
// @Summary 	Get the permission of the user in the group
// @Description Get the permission of the user in the group
// @Tags 		group
// @Accept 		json
// @Produce 	json
// @Param 		groupName query string true "Group name"
// @Param	    Cookie header string true "cookie parameter"
// @Success 200 {object} GroupPermissionResponse
// @Failure 400 {object} response
// @Failure 500 {object} response
// @Router /v1/group/permission [get]
func (h *Handler) GroupGetPermission(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["groupName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("group permission: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group permission: \"groupName\" argument missing"})
		return
	}
	groupName := keys[0]
	if groupName == "" {
		h.logger.Errorf("group permission: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group permission: \"groupName\" argument missing"})
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

	permission, err := h.dfsAPI.GetPermission(sessionId, groupName)
	if err != nil {
		h.logger.Errorf("group permission: failed to get group permission: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group permission: failed to get group permission"})
		return
	}
	jsonhttp.OK(w, &GroupPermissionResponse{
		Permission: permission,
	})
}
