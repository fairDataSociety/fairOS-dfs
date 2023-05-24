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

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

// KVLoadCSVHandler godoc
//
//	@Summary      Upload a csv file in kv table
//	@Description  KVLoadCSVHandler is the api handler to load a csv file as key and value in a KV table
//	@ID		      kv-loadcsv
//	@Tags         kv
//	@Accept       mpfd
//	@Produce      json
//	@Param	      podName formData string true "pod name"
//	@Param	      tableName formData string true "table name"
//	@Param	      memory formData string false "keep in memory"
//	@Param	      csv formData file true "file to upload"
//	@Param	      Cookie header string true "cookie parameter"
//	@Success      200  {object}  response
//	@Failure      400  {object}  response
//	@Failure      500  {object}  response
//	@Router       /v1/kv/loadcsv [Post]
func (h *Handler) KVLoadCSVHandler(w http.ResponseWriter, r *http.Request) {
	podName := r.FormValue("podName")
	if podName == "" {
		h.logger.Errorf("kv loadcsv: \"podName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv loadcsv: \"podName\" argument missing"})
		return
	}

	name := r.FormValue("tableName")
	if name == "" {
		h.logger.Errorf("kv loadcsv: \"tableName\" argument missing")
		jsonhttp.BadRequest(w, &response{Message: "kv loadcsv: \"tableName\" argument missing"})
		return
	}

	mem := r.FormValue("memory")
	memory := true
	if mem == "" {
		memory = false
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv loadcsv: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, &response{Message: ErrInvalidCookie.Error()})
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv loadcsv: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, &response{Message: "kv loadcsv: \"cookie-id\" parameter missing in cookie"})
		return
	}

	//  get the files parameter from the multipart
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.BadRequest(w, &response{Message: "kv loadcsv: " + err.Error()})
		return
	}
	files := r.MultipartForm.File["csv"]
	if len(files) == 0 {
		h.logger.Errorf("kv loadcsv: parameter \"csv\" missing")
		jsonhttp.BadRequest(w, &response{Message: "kv loadcsv: parameter \"csv\" missing"})
		return
	}

	file := files[0]
	fd, err := file.Open()
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv loadcsv: " + err.Error()})
		return
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)
	readHeader := false
	rowCount := 0
	successCount := 0
	failureCount := 0
	var batch *collection.Batch
	for {
		// read one row from csv (assuming
		record, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		rowCount++
		if err != nil {
			h.logger.Errorf("kv loadcsv: error loading row %d: %v", rowCount, err)
			failureCount++
			continue
		}

		record = strings.TrimSuffix(record, "\n")
		record = strings.TrimSuffix(record, "\r")
		if !readHeader {
			columns := strings.Split(record, ",")
			batch, err = h.dfsAPI.KVBatch(sessionId, podName, name, columns)
			if err != nil {
				h.logger.Errorf("kv loadcsv: %v", err)
				jsonhttp.InternalServerError(w, &response{Message: "kv loadcsv: " + err.Error()})
				return
			}

			err = batch.Put(collection.CSVHeaderKey, []byte(record), false, memory)
			if err != nil {
				h.logger.Errorf("kv loadcsv: error adding header %d: %v", rowCount, err)
				failureCount++
				readHeader = true
				continue
			}
			readHeader = true
			successCount++
			continue
		}

		key := strings.Split(record, ",")[0]
		err = batch.Put(key, []byte(record), false, memory)
		if err != nil {
			h.logger.Errorf("kv loadcsv: error adding row %d: %v", rowCount, err)
			failureCount++
			continue
		}
		successCount++
	}
	_, err = batch.Write("")
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, &response{Message: "kv loadcsv: " + err.Error()})
		return
	}

	sendStr := fmt.Sprintf("csv file loaded in to kv table (%s) with total:%d, success: %d, failure: %d rows", name, rowCount, successCount, failureCount)
	jsonhttp.OK(w, &response{Message: sendStr})
}
