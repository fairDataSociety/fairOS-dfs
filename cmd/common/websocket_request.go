package common

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Event string

var (
	UserSignup      Event = "/user/signup"
	UserLogin       Event = "/user/login"
	UserImport      Event = "/user/import"
	UserPresent     Event = "/user/present"
	UserIsLoggedin  Event = "/user/isloggedin"
	UserLogout      Event = "/user/logout"
	UserExport      Event = "/user/export"
	UserDelete      Event = "/user/delete"
	UserStat        Event = "/user/stat"
	PodNew          Event = "/pod/new"
	PodOpen         Event = "/pod/open"
	PodClose        Event = "/pod/close"
	PodSync         Event = "/pod/sync"
	PodDelete       Event = "/pod/delete"
	PodLs           Event = "/pod/ls"
	PodStat         Event = "/pod/stat"
	PodShare        Event = "/pod/share"
	PodReceive      Event = "/pod/receive"
	PodReceiveInfo  Event = "/pod/receiveinfo"
	DirIsPresent    Event = "/dir/present"
	DirMkdir        Event = "/dir/mkdir"
	DirRmdir        Event = "/dir/rmdir"
	DirLs           Event = "/dir/ls"
	DirStat         Event = "/dir/stat"
	FileDownload    Event = "/file/download"
	FileUpload      Event = "/file/upload"
	FileShare       Event = "/file/share"
	FileReceive     Event = "/file/receive"
	FileReceiveInfo Event = "/file/receiveinfo"
	FileDelete      Event = "/file/delete"
	FileStat        Event = "/file/stat"
	KVCreate        Event = "/kv/new"
	KVList          Event = "/kv/ls"
	KVOpen          Event = "/kv/open"
	KVDelete        Event = "/kv/delete"
	KVCount         Event = "/kv/count"
	KVEntryPut      Event = "/kv/entry/put"
	KVEntryGet      Event = "/kv/entry/get"
	KVEntryDelete   Event = "/kv/entry/del"
	KVLoadCSV       Event = "/kv/loadcsv"
	KVSeek          Event = "/kv/seek"
	KVSeekNext      Event = "/kv/seek/next"
	DocCreate       Event = "/doc/new"
	DocList         Event = "/doc/ls"
	DocOpen         Event = "/doc/open"
	DocCount        Event = "/doc/count"
	DocDelete       Event = "/doc/delete"
	DocFind         Event = "/doc/find"
	DocEntryPut     Event = "/doc/entry/put"
	DocEntryGet     Event = "/doc/entry/get"
	DocEntryDel     Event = "/doc/entry/del"
	DocLoadJson     Event = "/doc/loadjson"
	DocIndexJson    Event = "/doc/indexjson"
)

type WebsocketRequest struct {
	Event  Event       `json:"event"`
	Params interface{} `json:"params,omitempty"`
}

type FileRequest struct {
	PodName   string `json:"pod_name,omitempty"`
	TableName string `json:"table_name,omitempty"`
	DirPath   string `json:"dir_path,omitempty"`
	BlockSize string `json:"block_size,omitempty"`
	FileName  string `json:"file_name,omitempty"`
}

type FileDownloadRequest struct {
	PodName  string `json:"pod_name,omitempty"`
	Filepath string `json:"file_path,omitempty"`
}

type WebsocketResponse struct {
	Event      Event       `json:"event"`
	StatusCode int         `json:"code"`
	Params     interface{} `json:"params,omitempty"`
	header     http.Header
	buf        bytes.Buffer
}

func NewWebsocketResponse() *WebsocketResponse {
	return &WebsocketResponse{
		header: map[string][]string{},
	}
}

func (w *WebsocketResponse) Header() http.Header {
	return w.header
}

func (w *WebsocketResponse) Write(bytes []byte) (int, error) {
	if w.Header().Get("Content-Type") == "application/json; charset=utf-8" ||
		w.Header().Get("Content-Type") == "application/json" {
		body := map[string]interface{}{}
		err := json.Unmarshal(bytes, &body)
		if err != nil {
			return 0, err
		}
		w.Params = body
		return len(bytes), nil
	}
	if w.Header().Get("Content-Length") != "" || w.Header().Get("Content-Length") != "0" {
		return w.buf.Write(bytes)
	}
	return 0, nil
}

func (w *WebsocketResponse) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *WebsocketResponse) Marshal() []byte {
	if w.Header().Get("Content-Type") == "application/json; charset=utf-8" ||
		w.Header().Get("Content-Type") == "application/json" {
		data, _ := json.Marshal(w)
		return data
	}
	if w.Header().Get("Content-Length") != "" {
		return w.buf.Bytes()
	}
	return nil
}
