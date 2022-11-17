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
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// DocLoadJsonHandler godoc
//
//	@Summary      Load json file from local file system
//	@Description  DocLoadJsonHandler is the api handler that indexes a json file that is present in the local file system
//	@Tags         doc
//	@Accept       mpfd
//	@Produce      json
//	@Param	      podName query string true "pod name"
//	@Param	      tableName query string true "table name"
//	@Param	      json formData file true "json to index"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/doc/loadjson [post]
func (h *Handler) DocLoadJsonHandler(w http.ResponseWriter, r *http.Request) {
	podName := r.FormValue("podName")
	if podName == "" {
		h.logger.Errorf("doc loadjson: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc loadjson: \"pod_name\" argument missing"})
		return
	}

	name := r.FormValue("tableName")
	if name == "" {
		h.logger.Errorf("doc loadjson: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "doc loadjson: \"tableName\" argument missing"})
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc loadjson: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc loadjson: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "doc loadjsonv: \"cookie-id\" parameter missing in cookie"})
		return
	}

	//  get the files parameter from the multi part
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.BadRequest(w, &response{Message: "doc loadjson: " + err.Error()})
		return
	}
	files := r.MultipartForm.File["json"]
	if len(files) == 0 {
		h.logger.Errorf("doc loadjson: parameter \"csv\" missing")
		jsonhttp.BadRequest(w, &response{Message: "doc loadjson: parameter \"csv\" missing"})
		return
	}

	file := files[0]
	fd, err := file.Open()
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc loadjson: " + err.Error()})
		return
	}
	defer fd.Close()
	reader := bufio.NewReader(fd)
	rowCount := 0
	successCount := 0
	failureCount := 0
	docBatch, err := h.dfsAPI.DocBatch(sessionId, podName, name)
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc loadjson: " + err.Error()})
		return
	}

	for {
		// read one row from csv (assuming
		record, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		rowCount++
		if err != nil {
			h.logger.Errorf("doc loadjson: error loading row %d: %v", rowCount, err)
			failureCount++
			continue
		}

		record = strings.TrimSuffix(record, "\n")
		record = strings.TrimSuffix(record, "\r")

		err = h.dfsAPI.DocBatchPut(sessionId, podName, []byte(record), docBatch)
		if err != nil {
			failureCount++
			continue
		}
		successCount++

		if (rowCount % 10000) == 0 {
			h.logger.Info("uploaded ", rowCount)
		}
	}
	err = h.dfsAPI.DocBatchWrite(sessionId, podName, docBatch)
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "doc loadjson: " + err.Error()})
		return
	}

	sendStr := fmt.Sprintf("json file loaded in to document db (%s) with total:%d, success: %d, failure: %d rows", name, rowCount, successCount, failureCount)
	jsonhttp.OK(w, &response{Message: sendStr})
}
