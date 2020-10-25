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
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"resenje.org/jsonhttp"
)

func (h *Handler) KVLoadCSVHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		h.logger.Errorf("kv loadcsv: \"name\" argument missing")
		jsonhttp.BadRequest(w, "kv loadcsv: \"name\" argument missing")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("kv loadcsv: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("kv loadcsv: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "kv loadcsv: \"cookie-id\" parameter missing in cookie")
		return
	}

	//  get the files parameter from the multi part
	err = r.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.BadRequest(w, "kv loadcsv: "+err.Error())
		return
	}
	files := r.MultipartForm.File["csv"]
	if len(files) == 0 {
		h.logger.Errorf("kv loadcsv: parameter \"csv\" missing")
		jsonhttp.BadRequest(w, "kv loadcsv: parameter \"csv\" missing")
		return
	}

	file := files[0]
	fd, err := file.Open()
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
		return
	}

	batch, err := h.dfsAPI.KVBatch(sessionId, name)
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
		return
	}

	reader := csv.NewReader(fd)
	readHeader := false
	columns := make(map[string]string)
	for {
		// read one row from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.Errorf("kv loadcsv: %v", err)
			jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
			return
		}
		if !readHeader {
			for i, colName := range record {
				columns[strconv.Itoa(i)] = colName
			}
			readHeader = true
			continue
		}

		row := make(map[string]string)
		for i, colValue := range record {
			colName := columns[strconv.Itoa(i)]
			row[colName] = colValue
		}
		jsonString, err := json.Marshal(row)
		if err != nil {
			h.logger.Errorf("kv loadcsv: %v", err)
			jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
			return
		}

		err = batch.Put(record[0], jsonString)
		if err != nil {
			h.logger.Errorf("kv loadcsv: %v", err)
			jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
			return
		}
	}
	err = batch.Write()
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
		return
	}

	err = fd.Close()
	if err != nil {
		h.logger.Errorf("kv loadcsv: %v", err)
		jsonhttp.InternalServerError(w, "kv loadcsv: "+err.Error())
		return
	}

	jsonhttp.OK(w, "file loaded in to kv store")
}
