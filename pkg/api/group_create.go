package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"resenje.org/jsonhttp"
)

// GroupNameRequest is the request to create a group
type GroupNameRequest struct {
	GroupName string `json:"groupName,omitempty"`
}

// GroupCreateHandler godoc
//
//	@Summary      Create group
//	@Description  GroupCreateHandler is the api handler to create a new group
//	@ID           group-create-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      group_request body GroupNameRequest true "group name"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/new [post]
func (h *Handler) GroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group new: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group new: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupNameRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group new: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group new: could not decode arguments"})
		return
	}

	group := req.GroupName
	if group == "" {
		h.logger.Errorf("group new: \"groupName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group new: \"groupName\" argument missing"})
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

	_, err = h.dfsAPI.CreateGroup(sessionId, group)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, dfs.ErrUserNotLoggedIn) ||
			errors.Is(err, p.ErrInvalidPodName) ||
			errors.Is(err, p.ErrTooLongPodName) ||
			errors.Is(err, p.ErrPodAlreadyExists) ||
			errors.Is(err, p.ErrMaxPodsReached) {
			h.logger.Errorf("group new: %v", err)
			jsonhttp.BadRequest(w, &response{Message: "group new: " + err.Error()})
			return
		}
		h.logger.Errorf("group new: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group new: " + err.Error()})
		return
	}

	jsonhttp.Created(w, &response{Message: "group created successfully"})
}
