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
func (h *Handler) ACTOpenPod(w http.ResponseWriter, r *http.Request) {}
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
	w.WriteHeader(http.StatusOK)
}
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
	w.WriteHeader(http.StatusOK)
}

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
	w.WriteHeader(http.StatusOK)
}
