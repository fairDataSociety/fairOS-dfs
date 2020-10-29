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
	"errors"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"net/http"
	"resenje.org/jsonhttp"
	"strconv"
)

func (h *Handler) KVSeekHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv seek: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv seek: \"name\" argument missing")
		return
	}

	start := r.FormValue("start")
	if start == "" {
		h.logger.Errorf("kv seek: \"start\" argument missing")
		jsonhttp.BadRequest(w, "kv seek: \"start\" argument missing")
		return
	}

	end := r.FormValue("end")
	if end == "" {
		h.logger.Errorf("kv seek: \"end\" argument missing")
		jsonhttp.BadRequest(w, "kv seek: \"end\" argument missing")
		return
	}

	limit := r.FormValue("limit")
	if limit == "" {
		h.logger.Errorf("kv seek: \"limit\" argument missing")
		jsonhttp.BadRequest(w, "kv limit: \"start\" argument missing")
		return
	}
	noOfRows, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		h.logger.Errorf("kv seek: invalid limit")
		jsonhttp.BadRequest(w, "kv seek: invalid limit")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv seek: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv seek: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv seek: \"cookie-id\" parameter missing in cookie")
		return
	}

	_, err = h.dfsAPI.KVSeek(sessionId, name, start, end, noOfRows)
	if err != nil {
		h.logger.Errorf("kv seek: %v", err)
		jsonhttp.InternalServerError(w, "kv seek: "+err.Error())
		return
	}
	jsonhttp.OK(w, "seeked closest to the start key")
}

func (h *Handler) KVGetNextHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv get_next: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv get_next: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv get_next: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv get_next: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv get_next: \"cookie-id\" parameter missing in cookie")
		return
	}

	columns, key, data, err := h.dfsAPI.KVGetNext(sessionId, name)
	if err != nil && !errors.Is(err, collection.ErrNoNextElement) {
		h.logger.Errorf("kv get_next: %v", err)
		jsonhttp.InternalServerError(w, "kv get_next: "+err.Error())
		return
	}

	if errors.Is(err, collection.ErrNoNextElement) {
		jsonhttp.Respond(w, http.StatusNoContent, nil)
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
