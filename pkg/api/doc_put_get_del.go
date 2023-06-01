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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DocPutRequest is used to put entry in doc store
type DocPutRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	Document  string `json:"doc,omitempty"`
}

// DocDeleteRequest is used to delete entry in doc store
type DocDeleteRequest struct {
	PodName   string `json:"podName,omitempty"`
	TableName string `json:"tableName,omitempty"`
	ID        string `json:"id,omitempty"`
}

// DocGetResponse represents a single document row
type DocGetResponse struct {
	Doc []byte `json:"doc"`
}

// DocGet represents a single document row
type DocGet struct {
	Doc string `json:"doc"`
}

// DocEntryPutHandler godoc
//
//	@Summary      Add a record in document datastore
//	@Description  DocEntryPutHandler is the api handler add a document in to a document datastore
//	@ID		      doc-entry-put
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_entry_put_request query DocPutRequest true "doc put request"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/entry/put [post]
func (h *Handler) DocEntryPutHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc put: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc put: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocPutRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc put: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc put: could not decode arguments"})
		return
	}
	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc put: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc put: \"podName\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc put: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc put: \"tableName\" argument missing"})
		return
	}

	doc := docReq.Document
	if doc == "" {
		h.logger.Errorf("doc put: \"doc\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc put: \"doc\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc put: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc put: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc put: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.DocPut(sessionId, podName, name, []byte(doc))
	if err != nil {
		h.logger.Errorf("doc put: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc put: " + err.Error()})
		return
	}
	jsonhttp.OK(w, &response{Message: "added document to db"})
}

// DocEntryGetHandler godoc
//
//	@Summary      Get a document from a document datastore
//	@Description  DocEntryGetHandler is the api handler to get a document from a document datastore
//	@ID		      doc-entry-get
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      id query string true "id to search for"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  DocGet "base64 encoded string"
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/entry/get [get]
func (h *Handler) DocEntryGetHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["podName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"podName\" argument missing"})
		return
	}
	podName := keys[0]
	if podName == "" {
		h.logger.Errorf("doc get: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"podName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["tableName"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"tableName\" argument missing"})
		return
	}
	name := keys[0]
	if name == "" {
		h.logger.Errorf("doc get: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"tableName\" argument missing"})
		return
	}

	keys, ok = r.URL.Query()["id"]
	if !ok || len(keys[0]) < 1 {
		h.logger.Errorf("doc get: \"id\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"id\" argument missing"})
		return
	}
	id := keys[0]
	if id == "" {
		h.logger.Errorf("doc get: \"id\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"id\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc get: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc get: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc get: \"cookie-id\" parameter missing in cookie"})
		return
	}

	data, err := h.dfsAPI.DocGet(sessionId, podName, name, id)
	if err != nil {
		h.logger.Errorf("doc get: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc get: " + err.Error()})
		return
	}

	var getResponse DocGetResponse
	getResponse.Doc = data

	w.Header().Set("Content-Type", "application/json")
	jsonhttp.OK(w, &getResponse)
}

// DocEntryDelHandler godoc
//
//	@Summary      Delete a document from a document datastore
//	@Description  DocEntryDelHandler is the api handler to delete a document from a document datastore
//	@ID		      doc-entry-del
//	@Tags         doc
//	@Accept       json
//	@Produce      json
//	@Param	      doc_entry_delete_request query DocDeleteRequest true "doc entry delete"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/entry/del [delete]
func (h *Handler) DocEntryDelHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonContentType {
		h.logger.Errorf("doc del: invalid request body type")
		jsonhttp.BadRequest(w, &response{Message: "doc del: invalid request body type"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var docReq DocDeleteRequest
	err := decoder.Decode(&docReq)
	if err != nil {
		h.logger.Errorf("doc del: could not decode arguments")
		jsonhttp.BadRequest(w, &response{Message: "doc del: could not decode arguments"})
		return
	}

	podName := docReq.PodName
	if podName == "" {
		h.logger.Errorf("doc del: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc del: \"podName\" argument missing"})
		return
	}

	name := docReq.TableName
	if name == "" {
		h.logger.Errorf("doc del: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc del: \"tableName\" argument missing"})
		return
	}

	id := docReq.ID
	if id == "" {
		h.logger.Errorf("doc del: \"id\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc del: \"id\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc del: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc del: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc del: \"cookie-id\" parameter missing in cookie"})
		return
	}

	err = h.dfsAPI.DocDel(sessionId, podName, name, id)
	if err != nil {
		h.logger.Errorf("doc del: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc del: " + err.Error()})
		return
	}

	jsonhttp.OK(w, &response{Message: "deleted document from db"})
}
