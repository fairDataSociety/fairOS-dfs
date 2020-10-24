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
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

func (h *Handler) CollectionPutHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("collection put: \"name\" argument missing")
		jsonhttp.BadRequest(w, "collection put: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
	if name == "" {
		h.logger.Errorf("collection put: \"key\" argument missing")
		jsonhttp.BadRequest(w, "collection put: \"key\" argument missing")
		return
	}

	value := r.FormValue("value")
	if value == "" {
		h.logger.Errorf("collection put: \"value\" argument missing")
		jsonhttp.BadRequest(w, "collection put: \"value\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("collection put: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("collection put: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "collection put: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.Put(sessionId, name, key, []byte(value))
	if err != nil {
		h.logger.Errorf("collection put: %v", err)
		jsonhttp.InternalServerError(w, "collection put: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key added")
}

func (h *Handler) CollectionGetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("collection get: \"name\" argument missing")
		jsonhttp.BadRequest(w, "collection get: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
	if name == "" {
		h.logger.Errorf("collection get: \"key\" argument missing")
		jsonhttp.BadRequest(w, "collection get: \"key\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("collection get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("collection get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "collection get: \"cookie-id\" parameter missing in cookie")
		return
	}

	data, err := h.dfsAPI.Get(sessionId, name, key)
	if err != nil {
		h.logger.Errorf("collection get: %v", err)
		jsonhttp.InternalServerError(w, "collection get: "+err.Error())
		return
	}
	jsonhttp.OK(w, string(data))
}

func (h *Handler) CollectionDelHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("collection del: \"name\" argument missing")
		jsonhttp.BadRequest(w, "collection del: \"name\" argument missing")
		return
	}

	key := r.FormValue("key")
	if name == "" {
		h.logger.Errorf("collection del: \"key\" argument missing")
		jsonhttp.BadRequest(w, "collection del: \"key\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("collection del: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("collection del: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "collection del: \"cookie-id\" parameter missing in cookie")
		return
	}

	_, err = h.dfsAPI.Delete(sessionId, name, key)
	if err != nil {
		h.logger.Errorf("collection del: %v", err)
		jsonhttp.InternalServerError(w, "collection del: "+err.Error())
		return
	}
	jsonhttp.OK(w, "key deleted")
}
