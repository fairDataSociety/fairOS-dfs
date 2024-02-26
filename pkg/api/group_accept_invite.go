package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"resenje.org/jsonhttp"
)

// GroupInviteRequest is the request to accept a group invite
type GroupInviteRequest struct {
	Reference string `json:"reference,omitempty"`
}

// GroupAcceptInviteHandler godoc
//
//	@Summary      Accept group membersion
//	@Description  GroupAcceptInviteHandler is the api handler to accept a group invite
//	@ID           group-accept-invite-handler
//	@Tags         group
//	@Accept       json
//	@Produce      json
//	@Param	      reference body GroupInviteRequest true "reference of the invite"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/group/accept [post]
func (h *Handler) GroupAcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("group accept: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "group accept: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req GroupInviteRequest
	err := decoder.Decode(&req)
	if err != nil {
		h.logger.Errorf("group accept: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "group accept: could not decode arguments"})
		return
	}

	reference := req.Reference
	if reference == "" {
		h.logger.Errorf("group accept: \"reference\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "group accept: \"reference\" argument missing"})
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

	err = h.dfsAPI.AcceptGroupInvite(sessionId, []byte(reference))
	if err != nil {
		h.logger.Errorf("group accept: failed to accept group invite: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "group accept: failed to accept group invite"})
		return
	}

	jsonhttp.OK(w, &response{Message: "group invite accepted"})
}
