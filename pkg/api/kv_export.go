package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

const MaxExportLimit = 100

// KVExportHandler is the api handler to export from a particular key with the given prefix
// it takes four arguments, 2 mandatory and two optional
// - table_name: the name of the kv table
// - start_prefix: the prefix of the key to seek
// * end_prefix: the prefix of the end key
// * limit: the threshold for the number of keys to go when get_next is called
func (h *Handler) KVExportHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv export: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv export: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv export: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv export: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv export: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"pod_name\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv export: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv export: \"table_name\" argument missing"})
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
		limit = fmt.Sprintf("%d", MaxExportLimit)
	}
	noOfRows, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		h.logger.Errorf("kv export: invalid limit")
		jsonhttp.BadRequest(w, &response{Message: "kv export: invalid limit"})
		return
	}

	if noOfRows > MaxExportLimit {
		h.logger.Errorf("kv export: maximum limit is %d", MaxExportLimit)
		jsonhttp.BadRequest(w, &response{Message: fmt.Sprintf("kv export: maximum limit is %d", MaxExportLimit)})
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
