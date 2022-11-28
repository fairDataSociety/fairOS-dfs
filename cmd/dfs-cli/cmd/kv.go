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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/tinygrasshopper/bettercsv"
)

func kvNew(podName, tableName, indexType string) {
	kvNewReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
		IndexType: indexType,
	}
	jsonData, err := json.Marshal(kvNewReq)
	if err != nil {
		fmt.Println("kv new: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiKVCreate, jsonData)
	if err != nil {
		fmt.Println("kv new: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvDelete(podName, tableName string) {
	kvDelReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
	}
	jsonData, err := json.Marshal(kvDelReq)
	if err != nil {
		fmt.Println("kv del: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiKVDelete, jsonData)
	if err != nil {
		fmt.Println("kv del: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvList(podName string) {
	argString := fmt.Sprintf("podName=%s", podName)
	data, err := fdfsAPI.getReq(apiKVList, argString)
	if err != nil {
		fmt.Println("kv ls: ", err)
		return
	}
	var resp api.Collections
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("kv ls: ", err)
		return
	}
	for _, table := range resp.Tables {
		fmt.Println("<KV>: ", table.Name)
	}
}

func kvOpen(podName, tableName string) {
	kvOpenReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
	}
	jsonData, err := json.Marshal(kvOpenReq)
	if err != nil {
		fmt.Println("kv open: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiKVOpen, jsonData)
	if err != nil {
		fmt.Println("kv open: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvCount(podName, tableName string) {
	kvCountReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
	}
	jsonData, err := json.Marshal(kvCountReq)
	if err != nil {
		fmt.Println("kv count: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiKVCount, jsonData)
	if err != nil {
		fmt.Println("kv count: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvPut(podName, tableName, key, value string) {
	kvPutReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
		Key:       key,
		Value:     value,
	}
	jsonData, err := json.Marshal(kvPutReq)
	if err != nil {
		fmt.Println("kv count: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiKVEntryPut, jsonData)
	if err != nil {
		fmt.Println("kv put: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvget(podName, tableName, key string) {
	argString := fmt.Sprintf("podName=%s&tableName=%s&key=%s", podName, tableName, key)
	data, err := fdfsAPI.getReq(apiKVEntryGet, argString)
	if err != nil {
		fmt.Println("kv get: ", err)
		return
	}
	var resp api.KVResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("kv get: ", err)
		return
	}

	rdr := bytes.NewReader(resp.Values)
	csvReader := bettercsv.NewReader(rdr)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true
	csvReader.Quote = '"'
	content, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println("kv get: ", err)
		return
	}
	values := content[0]
	for i, name := range resp.Keys {
		fmt.Println(name + " : " + values[i])
	}
}

func kvDel(podName, tableName, key string) {
	kvDelReq := common.KVRequest{
		PodName:   podName,
		TableName: tableName,
		Key:       key,
	}
	jsonData, err := json.Marshal(kvDelReq)
	if err != nil {
		fmt.Println("kv count: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiKVEntryDelete, jsonData)
	if err != nil {
		fmt.Println("kv del: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func loadcsv(podName, tableName, fileName, localCsvFile, memory string) {
	fd, err := os.Open(localCsvFile)
	if err != nil {
		fmt.Println("loadcsv failed: ", err)
		return
	}
	fi, err := fd.Stat()
	if err != nil {
		fmt.Println("loadcsv failed: ", err)
		return
	}
	args := make(map[string]string)
	args["podName"] = podName
	args["tableName"] = tableName
	if memory == "true" {
		args["memory"] = tableName
	}
	data, err := fdfsAPI.uploadMultipartFile(apiKVLoadCSV, fileName, fi.Size(), fd, args, "csv", "false")
	if err != nil {
		fmt.Println("loadcsv: ", err)
		return
	}
	var resp api.UploadFileResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("loadcsv: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvSeek(podName, tableName, start, end, limit string) {
	kvSeekReq := common.KVRequest{
		PodName:     podName,
		TableName:   tableName,
		StartPrefix: start,
		EndPrefix:   end,
		Limit:       limit,
	}
	jsonData, err := json.Marshal(kvSeekReq)
	if err != nil {
		fmt.Println("kv seek: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiKVSeek, jsonData)
	if err != nil {
		fmt.Println(err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func kvGetNext(podName, tableName string) {
	argString := fmt.Sprintf("podName=%s&tableName=%s", podName, tableName)
	data, err := fdfsAPI.getReq(apiKVSeekNext, argString)
	if err != nil && !errors.Is(err, collection.ErrNoNextElement) {
		fmt.Println("kv get_next: ", err)
		return
	}

	if errors.Is(err, collection.ErrNoNextElement) {
		fmt.Println("no next element")
	} else {
		var resp api.KVResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("kv get_next: ", err)
			return
		}

		rdr := bytes.NewReader(resp.Values)
		csvReader := bettercsv.NewReader(rdr)
		csvReader.Comma = ','
		csvReader.Quote = '"'
		content, err := csvReader.ReadAll()
		if err != nil {
			fmt.Println("kv get_next: ", err)
			return
		}
		values := content[0]
		for i, name := range resp.Keys {
			fmt.Println(name + " : " + values[i])
		}
	}
}
