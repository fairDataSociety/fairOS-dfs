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

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
)

func docNew(tableName, simpleIndex, mutableStr string) {
	mutable := true
	if mutableStr != "" {
		mut, err := strconv.ParseBool(mutableStr)
		if err != nil {
			fmt.Println("doc new: error parsing \"mutable\" string")
			return
		}
		mutable = mut
	}

	docNewReq := common.DocRequest{
		TableName:   tableName,
		SimpleIndex: simpleIndex,
		Mutable:     mutable,
	}
	jsonData, err := json.Marshal(docNewReq)
	if err != nil {
		fmt.Println("doc new: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDocCreate, jsonData)
	if err != nil {
		fmt.Println("doc new: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docList() {
	data, err := fdfsAPI.postReq(http.MethodGet, apiDocList, nil)
	if err != nil {
		fmt.Println("doc ls: ", err)
		return
	}
	var resp api.DocumentDBs
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("doc ls: ", err)
		return
	}
	for _, table := range resp.Tables {
		fmt.Println("<DOC>: ", table.Name)
		for fn, ft := range table.IndexedColumns {
			fmt.Println("     SI:", fn, ft)
		}
	}
}

func docOpen(tableName string) {
	docOpenReq := common.DocRequest{
		TableName: tableName,
	}
	jsonData, err := json.Marshal(docOpenReq)
	if err != nil {
		fmt.Println("doc open: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDocOpen, jsonData)
	if err != nil {
		fmt.Println("doc open: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docCount(tableName, expression string) {
	docCountReq := common.DocRequest{
		TableName:  tableName,
		Expression: expression,
	}
	jsonData, err := json.Marshal(docCountReq)
	if err != nil {
		fmt.Println("doc count: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDocCount, jsonData)
	if err != nil {
		fmt.Println("doc count: ", err)
		return
	}
	count, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		fmt.Println("doc count: ", err)
		return
	}
	fmt.Println("Count = ", count)
}

func docDelete(tableName string) {
	docDeleteReq := common.DocRequest{
		TableName: tableName,
	}
	jsonData, err := json.Marshal(docDeleteReq)
	if err != nil {
		fmt.Println("doc delete: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiDocDelete, jsonData)
	if err != nil {
		fmt.Println("doc del: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docFind(tableName, expression, limit string) {
	argString := fmt.Sprintf("table_name=%s&expr=%s&limit=%s", tableName, expression, limit)
	data, err := fdfsAPI.getReq(apiDocFind, argString)
	if err != nil {
		fmt.Println("doc find: ", err)
		return
	}
	var docs api.DocFindResponse
	err = json.Unmarshal(data, &docs)
	if err != nil {
		fmt.Println("doc find: ", err)
		return
	}
	for i, doc := range docs.Docs {
		fmt.Println("--- doc ", i)
		var d map[string]interface{}
		err = json.Unmarshal(doc, &d)
		if err != nil {
			fmt.Println("doc find: ", err)
			return
		}
		for k, v := range d {
			var valStr string
			switch val := v.(type) {
			case string:
				fmt.Println(k, "=", val)
			case float64:
				valStr = fmt.Sprintf("%g", val)
				fmt.Println(k, "=", valStr)
			case map[string]interface{}:
				fmt.Println(k + ":")
				for k1, v1 := range val {
					switch val1 := v1.(type) {
					case string:
						fmt.Println("   "+k1+" = ", val1)
					case float64:
						valStr = fmt.Sprintf("%g", val1)
						fmt.Println("   "+k1+" = ", valStr)
					default:
						fmt.Println("   "+k1+" = ", val1)
					}
				}
			default:
				fmt.Println(k, "=", val)
			}
		}
	}
}

func docPut(tableName, document string) {
	docPutReq := common.DocRequest{
		TableName: tableName,
		Document:  document,
	}
	jsonData, err := json.Marshal(docPutReq)
	if err != nil {
		fmt.Println("doc put: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDocEntryPut, jsonData)
	if err != nil {
		fmt.Println("doc put: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docGet(tableName, id string) {
	argString := fmt.Sprintf("table_name=%s&id=%s", tableName, id)
	data, err := fdfsAPI.getReq(apiDocEntryGet, argString)
	if err != nil {
		fmt.Println("doc get: ", err)
		return
	}

	var doc api.DocGetResponse
	err = json.Unmarshal(data, &doc)
	if err != nil {
		fmt.Println("doc get: ", err)
		return
	}
	var d map[string]interface{}
	err = json.Unmarshal(doc.Doc, &d)
	if err != nil {
		fmt.Println("doc get: ", err)
		return
	}
	for k, v := range d {
		fmt.Println(k, "=", v)
	}
}

func docDel(tableName, id string) {
	docDelReq := common.DocRequest{
		TableName: tableName,
		ID:        id,
	}
	jsonData, err := json.Marshal(docDelReq)
	if err != nil {
		fmt.Println("doc del: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiDocEntryDel, jsonData)
	if err != nil {
		fmt.Println("doc del: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docLoadJson(localJsonFile, tableName, fileName string) {
	fd, err := os.Open(localJsonFile)
	if err != nil {
		fmt.Println("loadjson failed: ", err)
		return
	}
	fi, err := fd.Stat()
	if err != nil {
		fmt.Println("loadjson failed: ", err)
		return
	}

	args := make(map[string]string)
	args["name"] = tableName
	data, err := fdfsAPI.uploadMultipartFile(apiDocLoadJson, fileName, fi.Size(), fd, args, "json", "false")
	if err != nil {
		fmt.Println("loadjson: ", err)
		return
	}
	var resp api.UploadFileResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("loadjson: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func docIndexJson(tableName, fileName string) {
	docIndexJsonReq := common.DocRequest{
		TableName: tableName,
		FileName:  fileName,
	}
	jsonData, err := json.Marshal(docIndexJsonReq)
	if err != nil {
		fmt.Println("index json: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDocIndexJson, jsonData)
	if err != nil {
		fmt.Println("index json: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}
