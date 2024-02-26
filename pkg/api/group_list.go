package api

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// GroupListResponse represents the response of the group list request
type GroupListResponse pod.GroupList

// GroupListHandler is the handler for group list API
//
// @Summary 	List groups
// @Description List groups
// @ID 			group_list
// @Tags 		group
// @Accept  	json
// @Produce  	json
// @Param	    Cookie header string true "cookie parameter"
// @Success 200 {object} GroupListResponse
// @Failure 400 {object} response
// @Failure 500 {object} response
// @Router /v1/group/ls [get]
func (h *Handler) GroupListHandler(w http.ResponseWriter, r *http.Request) {
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

	groupList, err := h.dfsAPI.ListGroups(sessionId)
	if err != nil {
		h.logger.Errorf("ListGroups failed: ", err)
		jsonhttp.InternalServerError(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, groupList)
}
