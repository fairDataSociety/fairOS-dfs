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
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"io"
	"net/http"
	"resenje.org/jsonhttp"
	"strings"
)

func (h *Handler) DocLoadJsonHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("doc loadjson: \"name\" argument missing")
		jsonhttp.BadRequest(w, "doc loadjson: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("doc loadjson: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("doc loadjson: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "doc loadjsonv: \"cookie-id\" parameter missing in cookie")
		return
	}

	//  get the files parameter from the multi part
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.BadRequest(w, "doc loadjson: "+err.Error())
		return
	}
	files := r.MultipartForm.File["json"]
	if len(files) == 0 {
		h.logger.Errorf("doc loadjson: parameter \"csv\" missing")
		jsonhttp.BadRequest(w, "doc loadjson: parameter \"csv\" missing")
		return
	}

	file := files[0]
	fd, err := file.Open()
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.InternalServerError(w, "doc loadjson: "+err.Error())
		return
	}

	reader := bufio.NewReader(fd)
	rowCount := 0
	successCount := 0
	failureCount := 0
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

		err = h.dfsAPI.DocPut(sessionId, name, []byte(record))
		if err != nil {
			failureCount++
			continue
		}
		successCount++
	}
	err = fd.Close()
	if err != nil {
		h.logger.Errorf("doc loadjson: %v", err)
		jsonhttp.InternalServerError(w, "doc loadjson: "+err.Error())
		return
	}

	sendStr := fmt.Sprintf("json file loaded in to document db (%s) with total:%d, success: %d, failure: %d rows", name, rowCount, successCount, failureCount)
	jsonhttp.OK(w, sendStr)
}
