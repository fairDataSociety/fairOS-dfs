package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

const maxExportLimit = 100

// KVExportRequest is the request for kv export
type KVExportRequest struct {
	PodName     string `json:"podName,omitempty"`
	TableName   string `json:"tableName,omitempty"`
	StartPrefix string `json:"startPrefix,omitempty"`
	EndPrefix   string `json:"endPrefix,omitempty"`
	Limit       string `json:"limit,omitempty"`
}

// KVExportHandler godoc
//
//	@Summary      Export from a particular key with the given prefix
//	@Description  KVExportHandler is the api handler to export from a particular key with the given prefix
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      export_request body KVExportRequest true "kv export info"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  []map[string]interface{}
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/export [Post]
func (h *Handler) KVExportHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv export: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv export: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVExportRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv export: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv export: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv export: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"podName\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv export: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"tableName\" argument missing"})
		return
	}

	start := kvReq.StartPrefix
	if start == "" {
		h.logger.Errorf("kv export: \"start\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"start\" argument missing"})
		return
	}

	end := kvReq.EndPrefix
	limit := kvReq.Limit
	if limit == "" {
		limit = fmt.Sprintf("%d", maxExportLimit)
	}
	noOfRows, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		h.logger.Errorf("kv export: invalid limit")
		jsonhttp.BadRequest(w, &response{Message: "kv export: invalid limit"})
		return
	}

	if noOfRows > maxExportLimit {
		h.logger.Errorf("kv export: maximum limit is %d", maxExportLimit)
		jsonhttp.BadRequest(w, &response{Message: fmt.Sprintf("kv export: maximum limit is %d", maxExportLimit)})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv export: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv export: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"cookie-id\" parameter missing in cookie"})
		return
	}

	itr, err := h.dfsAPI.KVSeek(sessionId, podName, name, start, end, noOfRows)
	if err != nil {
		h.logger.Errorf("kv export: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv export: " + err.Error()})
		return
	}
	items := []map[string]interface{}{}
	var i int64
	for i = 0; i < noOfRows; i++ {
		if itr == nil {
			break
		}
		ok := itr.Next()
		if !ok {
			break
		}
		item := map[string]interface{}{}
		item[itr.StringKey()] = string(itr.Value())
		items = append(items, item)
	}
	resp := map[string]interface{}{}
	resp["items"] = items
	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &resp)
}
