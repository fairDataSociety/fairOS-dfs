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

	"github.com/gorilla/mux"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

type DocGetResponse struct {
	Doc []byte `json:"doc"`
}

func (h *Handler) DocPutHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc put: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc put: \"name\" argument missing")
		return
	}

	doc := r.FormValue("doc")
	if doc == "" {
		h.logger.Errorf("doc put: \"doc\" argument missing")
		jsonhttp.BadRequest(w, "doc put: \"doc\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc put: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc put: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc put: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocPut(sessionId, name, []byte(doc))
	if err != nil {
		h.logger.Errorf("doc put: %v", err)
		jsonhttp.InternalServerError(w, "doc put: "+err.Error())
		return
	}
	jsonhttp.OK(w, "added document to db")
}

func (h *Handler) DocGetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc get: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"name\" argument missing")
		return
	}

	id := r.FormValue("id")
	if id == "" {
		h.logger.Errorf("doc get: \"id\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"id\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc get: \"cookie-id\" parameter missing in cookie")
		return
	}

	data, err := h.dfsAPI.DocGet(sessionId, name, id)
	if err != nil {
		h.logger.Errorf("doc get: %v", err)
		jsonhttp.InternalServerError(w, "doc get: "+err.Error())
		return
	}

	var getResponse DocGetResponse
	getResponse.Doc = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &getResponse)
}

func (h *Handler) DocNewGetHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		h.logger.Errorf("doc get: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"name\" argument missing")
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		h.logger.Errorf("doc get: \"id\" argument missing")
		jsonhttp.BadRequest(w, "doc get: \"id\" argument missing")
		return
	}

	// get values from cookie
	cookieStr := r.FormValue("fairOS-dfs")
	sessionId, err := cookie.GetSessionIdFromRawCookie(cookieStr)
	if err != nil {
		h.logger.Errorf("doc get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc get: \"cookie-id\" parameter missing in cookie")
		return
	}

	data, err := h.dfsAPI.DocGet(sessionId, name, id)
	if err != nil {
		h.logger.Errorf("doc get: %v", err)
		jsonhttp.InternalServerError(w, "doc get: "+err.Error())
		return
	}

	//var getResponse DocGetResponse
	//getResponse.Doc = data

	//w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, data)
}

func (h *Handler) DocDelHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc del: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc del: \"name\" argument missing")
		return
	}

	id := r.FormValue("id")
	if id == "" {
		h.logger.Errorf("doc del: \"id\" argument missing")
		jsonhttp.BadRequest(w, "doc del: \"id\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc del: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc del: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc del: \"cookie-id\" parameter missing in cookie")
		return
	}

	err = h.dfsAPI.DocDel(sessionId, name, id)
	if err != nil {
		h.logger.Errorf("doc del: %v", err)
		jsonhttp.InternalServerError(w, "doc del: "+err.Error())
		return
	}

	jsonhttp.OK(w, "deleted document from db")
}
