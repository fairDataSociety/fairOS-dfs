/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"encoding/json"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type KVResponse struct {
	Keys   []string `json:"keys,omitempty"`
	Values []byte   `json:"values"`
}

// KVPutHandler is the api handler to insert a key and value in to the kv table
// it takes three arguments
// - table_name: the name of the kv table
// - key: the key string
// - value: the value to insert in bytes
func (h *Handler) KVPutHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv put: invalid request body type")
		jsonhttp.BadRequest(w, "kv put: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv put: could not decode arguments")
		jsonhttp.BadRequest(w, "kv put: could not decode arguments")
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv put: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"pod_name\" argument missing")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv put: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"name\" argument missing")
		return
	}

	key := kvReq.Key
	if name == "" {
		h.logger.Errorf("kv put: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"key\" argument missing")
		return
	}

	value := kvReq.Value
	if value == "" {
		h.logger.Errorf("kv put: \"value\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"value\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv put: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv put: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv put: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.KVPut(sessionId, podName, name, key, []byte(value))
	if err != nil {
		h.logger.Errorf("kv put: %v", err)
		jsonhttp.InternalServerError(w, "kv put: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key added")
}

// KVGetHandler is the api handler to get a value from the kv table
// it takes two arguments
// - table_name: the name of the kv table
// - key: the key string
func (h *Handler) KVGetHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["pod_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"pod_name\" argument missing")
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("kv get: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"pod_name\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["table_name"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"table_name\" argument missing")
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("kv get: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"table_name\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"sharing_ref\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"sharing_ref\" argument missing")
		return
	}
	key := keys[0]
	if key == "" {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"key\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv get: \"cookie-id\" parameter missing in cookie")
		return
	}

	columns, data, err := h.dfsAPI.KVGet(sessionId, podName, name, key)
	if err != nil {
		h.logger.Errorf("kv get: %v", err)
		jsonhttp.InternalServerError(w, "kv get: "+err.Error())
		return
	}

	var resp KVResponse
	if columns != nil {
		resp.Keys = columns
	} else {
		resp.Keys = []string{key}
	}
	resp.Values = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &resp)
}

// KVDelHandler is the api handler to delete a key and value from the kv table
// it takes two arguments
// - table_name: the name of the kv table
// - key: the key string
func (h *Handler) KVDelHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv delete: invalid request body type")
		jsonhttp.BadRequest(w, "kv delete: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq common.KVRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv delete: could not decode arguments")
		jsonhttp.BadRequest(w, "kv delete: could not decode arguments")
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv del: \"pod_name\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"pod_name\" argument missing")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv del: \"table_name\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"table_name\" argument missing")
		return
	}

	key := kvReq.Key
	if name == "" {
		h.logger.Errorf("kv del: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"key\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv del: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv del: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv del: \"cookie-id\" parameter missing in cookie")
		return
	}

	_, err = h.dfsAPI.KVDel(sessionId, podName, name, key)
	if err != nil {
		h.logger.Errorf("kv del: %v", err)
		jsonhttp.InternalServerError(w, "kv del: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key deleted")
}
