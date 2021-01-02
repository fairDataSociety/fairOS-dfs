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
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"net/http"
	"resenje.org/jsonhttp"
)

type KVResponse struct {
	Names  []string `json:"names,omitempty"`
	Values []byte   `json:"values"`
}

func (h *Handler) KVPutHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv put: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
	if name == "" {
		h.logger.Errorf("kv put: \"key\" argument missing")
		jsonhttp.BadRequest(w, "kv put: \"key\" argument missing")
		return
	}

	value := r.FormValue("value")
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

	err = h.dfsAPI.KVPut(sessionId, name, key, []byte(value))
	if err != nil {
		h.logger.Errorf("kv put: %v", err)
		jsonhttp.InternalServerError(w, "kv put: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key added")
}

func (h *Handler) KVGetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv get: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv get: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
	if name == "" {
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

	columns, data, err := h.dfsAPI.KVGet(sessionId, name, key)
	if err != nil {
		h.logger.Errorf("kv get: %v", err)
		jsonhttp.InternalServerError(w, "kv get: "+err.Error())
		return
	}

	var resp KVResponse
	if columns != nil {
		resp.Names = columns
	} else {
		resp.Names = []string{key}
	}
	resp.Values = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &resp)
}

func (h *Handler) KVDelHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv del: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv del: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
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

	_, err = h.dfsAPI.KVDel(sessionId, name, key)
	if err != nil {
		h.logger.Errorf("kv del: %v", err)
		jsonhttp.InternalServerError(w, "kv del: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key deleted")
}
