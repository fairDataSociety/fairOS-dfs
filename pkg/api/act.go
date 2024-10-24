package api

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/act"

	"github.com/btcsuite/btcd/btcec/v2"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"github.com/gorilla/mux"
	"resenje.org/jsonhttp"
)

// CreateGranteeHandler godoc
//
//	@Summary      Create ACT with grantee public key
//	@Description  CreateGranteeHandler is the api handler for creating act with grantee public key.
//	@ID		      create-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      grantee query string true "grantee public key"
//	@Success      201  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/grantee/{actName} [post]
func (h *Handler) CreateGranteeHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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

	vars := mux.Vars(r)
	actName := vars["actName"]

	userPublicKeyString := r.URL.Query().Get("grantee")
	pubk, err := hex.DecodeString(userPublicKeyString)
	if err != nil {
		h.logger.Error("create grantee failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	pub, err := btcec.ParsePubKey(pubk)
	if err != nil {
		h.logger.Error("create grantee failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	err = h.dfsAPI.CreateGranteePublicKey(sessionId, actName, pub.ToECDSA())
	if err != nil {
		h.logger.Error("create grantee failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	jsonhttp.Created(w, &response{Message: "act created"})
}

// GrantRevokeHandler godoc
//
//	@Summary      Grant ACT with grantee public key
//	@Description  GrantRevokeHandler is the api handler for granting and revoking access in an existing act.
//	@ID		      grant-revoke-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      grant query string false "grantee public key"
//	@Param	      revoke query string false "revoke public key"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/grantee/{actName} [patch]
func (h *Handler) GrantRevokeHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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

	vars := mux.Vars(r)
	actName := vars["actName"]

	grantUser := r.URL.Query().Get("grant")
	revokeUser := r.URL.Query().Get("revoke")
	if grantUser == "" && revokeUser == "" {
		h.logger.Error("grant revoke failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: "grant and revoke user public key required"})
		return
	}
	if grantUser == revokeUser {
		h.logger.Error("grant revoke failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: "grant and revoke user public key cannot be the same"})
	}
	var (
		granteePubKey *ecdsa.PublicKey
		removePubKey  *ecdsa.PublicKey
	)
	if grantUser != "" {
		pubkg, err := hex.DecodeString(grantUser)
		if err != nil {
			h.logger.Error("grant revoke failed: ", err)
			jsonhttp.BadRequest(w, &response{Message: err.Error()})
			return
		}
		pubg, err := btcec.ParsePubKey(pubkg)
		if err != nil {
			h.logger.Error("grant revoke failed: ", err)
			jsonhttp.BadRequest(w, &response{Message: err.Error()})
			return
		}
		granteePubKey = pubg.ToECDSA()
	}
	if revokeUser != "" {
		pubkr, err := hex.DecodeString(revokeUser)
		if err != nil {
			h.logger.Error("grant revoke failed: ", err)
			jsonhttp.BadRequest(w, &response{Message: err.Error()})
			return
		}
		pubr, err := btcec.ParsePubKey(pubkr)
		if err != nil {
			h.logger.Error("grant revoke failed: ", err)
			jsonhttp.BadRequest(w, &response{Message: err.Error()})
			return
		}
		removePubKey = pubr.ToECDSA()
	}

	err = h.dfsAPI.GrantRevokeGranteePublicKey(sessionId, actName, granteePubKey, removePubKey)
	if err != nil {
		h.logger.Error("grant revoke failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ListGranteesHandler godoc
//
//	@Summary      List grantees in ACT
//	@Description  ListGranteesHandler is the api handler for listing grantees in an existing act.
//	@ID		      list-grantee-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  []string
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/grantee/{actName} [get]
func (h *Handler) ListGranteesHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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

	vars := mux.Vars(r)
	actName := vars["actName"]

	grantees, err := h.dfsAPI.ListGrantees(sessionId, actName)
	if err != nil {
		h.logger.Error("list grantees failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, grantees)
}

// ACTPodShareHandler godoc
//
//	@Summary      share a pod in act
//	@Description  ACTPodShareHandler is the api handler for adding a pod in act.
//	@ID		      share-pod-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      podname path string true "pod to share in act"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/share-pod/{actName}/{podname} [post]
func (h *Handler) ACTPodShareHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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

	vars := mux.Vars(r)
	actName := vars["actName"]
	podName := vars["podName"]

	err = h.dfsAPI.ACTPodShare(sessionId, podName, actName)
	if err != nil {
		h.logger.Error("create grantee failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "pod shared"})
}

// ACTListHandler godoc
//
//	@Summary      List acts
//	@Description  ACTListHandler is the api handler for listing acts.
//	@ID		      list-act
//	@Tags         act
//	@Produce      json
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  act.List
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/list [get]
func (h *Handler) ACTListHandler(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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

	list, err := h.dfsAPI.GetACTs(sessionId)
	if err != nil {
		h.logger.Error("list acts failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, list)
}

// ACTSharedPods godoc
//
//	@Summary      List pods in act
//	@Description  ACTSharedPods is the api handler for listing pods shared in act.
//	@ID		      list-shared-pod-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  []act.Content
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/act-shared-pods [get]
func (h *Handler) ACTSharedPods(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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
	vars := mux.Vars(r)
	actName := vars["actName"]

	list, err := h.dfsAPI.GetACTContents(sessionId, actName)
	if err != nil {
		h.logger.Error("list contents failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}

	jsonhttp.OK(w, list)
}

// ACTOpenPod godoc
//
//	@Summary      Open Act pod
//	@Description  ACTOpenPod is the api handler for opening pod in act.
//	@ID		      open-pod-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/open-act-pod/{actName} [post]
func (h *Handler) ACTOpenPod(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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
	vars := mux.Vars(r)
	actName := vars["actName"]

	err = h.dfsAPI.OpenACTPod(sessionId, actName)
	if err != nil {
		h.logger.Error("open act pod failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "pod opened"})
}

// ACTSavePod godoc
//
//	@Summary      Save shared acted pod in act list
//	@Description  ACTSavePod is the api handler for saving shared act pod.
//	@ID		      save-pod-act
//	@Tags         act
//	@Produce      json
//	@Param	      actName path string true "unique act identifier"
//	@Param	      content body act.Content true "acted pod info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/act/save-act-pod/{actName} [post]
func (h *Handler) ACTSavePod(w http.ResponseWriter, r *http.Request) {
	sessionId, err := auth.GetSessionIdFromGitRequest(r)
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
	vars := mux.Vars(r)
	actName := vars["actName"]

	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("save act pod: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "save act pod: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var saveReq act.Content
	err = decoder.Decode(&saveReq)
	if err != nil {
		h.logger.Errorf("save act pod: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "save act pod: could not decode arguments"})
		return
	}

	err = h.dfsAPI.SaveACTPod(sessionId, actName, &saveReq)
	if err != nil {
		h.logger.Error("save act pod failed: ", err)
		jsonhttp.BadRequest(w, &response{Message: err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "pod saved"})
}
