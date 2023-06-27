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
	"fmt"
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"resenje.org/jsonhttp"
)

// KVEntryRequest is the request to put a key-value in the kv table
type KVEntryRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
}

// KVEntryDeleteRequest is the request to delete a key-value in the kv table
type KVEntryDeleteRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	Key       string `json:"key,omitempty"`
}

// KVResponse is the response to get a key-value from the kv table
type KVResponse struct {
	Keys   []string `json:"keys,omitempty"`
	Values []byte   `json:"values"`
}

// KVResponseRaw is the response to get a key-value from the kv table
type KVResponseRaw struct {
	Keys   []string `json:"keys,omitempty"`
	Values string   `json:"values"`
}

// KVPutHandler godoc
//
//	@Summary      put key and value in the kv table
//	@Description  KVPutHandler is the api handler to put a key-value  in the kv table
//	@ID		      kv-put
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      kv_entry body KVEntryRequest true "kv entry"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/entry/put [post]
func (h *Handler) KVPutHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv put: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "kv put: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVEntryRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv put: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "kv put: could not decode arguments"})
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv put: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv put: \"podName\" argument missing"})
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv put: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv put: \"tableName\" argument missing"})
		return
	}

	key := kvReq.Key
	if name == "" {
		h.logger.Errorf("kv put: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv put: \"key\" argument missing"})
		return
	}

	value := kvReq.Value
	if value == "" {
		h.logger.Errorf("kv put: \"value\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv put: \"value\" argument missing"})
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

	err = h.dfsAPI.KVPut(sessionId, podName, name, key, []byte(value))
	if err != nil {
		h.logger.Errorf("kv put: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv put: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "key added"})
}

// KVGetHandler godoc
//
//	@Summary      get value from the kv table
//	@Description  KVGetHandler is the api handler to get a value from the kv table
//	@ID		      kv-get
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      key query string true "key"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  KVResponse
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/entry/get [get]
func (h *Handler) KVGetHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"tableName\" argument missing"})
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"tableName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"key\" argument missing"})
		return
	}
	key := keys[0]
	if key == "" {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"key\" argument missing"})
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

	columns, data, err := h.dfsAPI.KVGet(sessionId, podName, name, key)
	if err != nil {
		h.logger.Errorf("kv get: %v", err)
		if err == collection.ErrEntryNotFound {
			jsonhttp.NotFound(w, &response{Message: "kv get: " + err.Error()})
			return
		}
		jsonhttp.InternalServerError(w, &response{Message: "kv get: " + err.Error()})
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

// KVGetDataHandler godoc
//
//	@Summary      get value from the kv table
//	@Description  KVGetDataHandler is the api handler to get raw value from the kv table
//	@ID		      kv-get-data
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      key query string true "key"
//	@Param	      format query string false "format of the value" example(byte-string, string)
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  KVResponseRaw
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/entry/get-data [get]
func (h *Handler) KVGetDataHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"tableName\" argument missing"})
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"tableName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"key\" argument missing"})
		return
	}
	key := keys[0]
	if key == "" {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"key\" argument missing"})
		return
	}

	formats, ok := r.URL.Query()["format"]
	if !ok || len(formats[0]) < 1 {
		h.logger.Errorf("kv get: \"format\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"format\" argument missing"})
		return
	}
	format := formats[0]
	if format == "" {
		h.logger.Errorf("kv get: \"format\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"format\" argument missing"})
		return
	}

	if format != "string" && format != "byte-string" {
		h.logger.Errorf("kv get: \"format\" argument is unknown")
		jsonhttp.BadRequest(w, &response{Message: "kv get: \"format\" argument is unknown"})
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

	columns, data, err := h.dfsAPI.KVGet(sessionId, podName, name, key)
	if err != nil {
		h.logger.Errorf("kv get: %v", err)
		if err == collection.ErrEntryNotFound {
			jsonhttp.NotFound(w, &response{Message: "kv get: " + err.Error()})
			return
		}
		jsonhttp.InternalServerError(w, &response{Message: "kv get: " + err.Error()})
		return
	}

	var resp KVResponseRaw
	if columns != nil {
		resp.Keys = columns
	} else {
		resp.Keys = []string{key}
	}

	if format == "string" {
		resp.Values = string(data)
	} else {
		resp.Values = fmt.Sprintf("%v", data)
	}

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &resp)
}

// KVDelHandler godoc
//
//	@Summary      Delete key-value from the kv table
//	@Description  KVDelHandler is the api handler to delete a key and value from the kv table
//	@ID		      kv-del
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      delete_request body KVEntryDeleteRequest true "delete request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  KVResponseRaw
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/entry/del [delete]
func (h *Handler) KVDelHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("kv delete: invalid request body type")
		jsonhttp.BadRequest(w, "kv delete: invalid request body type")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var kvReq KVEntryDeleteRequest
	err := decoder.Decode(&kvReq)
	if err != nil {
		h.logger.Errorf("kv delete: could not decode arguments")
		jsonhttp.BadRequest(w, "kv delete: could not decode arguments")
		return
	}

	podName := kvReq.PodName
	if podName == "" {
		h.logger.Errorf("kv del: \"podName\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"podName\" argument missing")
		return
	}

	name := kvReq.TableName
	if name == "" {
		h.logger.Errorf("kv del: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"tableName\" argument missing")
		return
	}

	key := kvReq.Key
	if name == "" {
		h.logger.Errorf("kv del: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"key\" argument missing")
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

	_, err = h.dfsAPI.KVDel(sessionId, podName, name, key)
	if err != nil {
		h.logger.Errorf("kv del: %v", err)
		jsonhttp.InternalServerError(w, "kv del: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key deleted")
}

// KVPresentHandler godoc
//
//	@Summary      Check if a value exists in the kv table
//	@Description  KVPresentHandler is the api handler to check if a value exists in the kv table
//	@ID           kv-present-handler
//	@Tags         kv
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      key query string true "key"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/entry/present [get]
func (h *Handler) KVPresentHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"podName\" argument missing")
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("kv get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"podName\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"tableName\" argument missing")
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("kv get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"tableName\" argument missing")
		return
	}

	keys, ok = r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"key\" argument missing")
		return
	}
	key := keys[0]
	if key == "" {
		h.logger.Errorf("kv get: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"key\" argument missing")
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
	w.Header().Set("Content-Type", "application/json")

	_, _, err = h.dfsAPI.KVGet(sessionId, podName, name, key)
	if err != nil {
		jsonhttp.OK(w, &PresentResponse{
			Present: false,
		})
		return
	}

	jsonhttp.OK(w, &PresentResponse{
		Present: true,
	})
}
