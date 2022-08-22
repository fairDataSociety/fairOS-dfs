package common

import (
	"bytes"
	"encoding/json"
)

type Event string

var (
	UserSignup         Event = "/user/signup"
	UserSignupV2       Event = "/user/signupV2"
	UserLogin          Event = "/user/login"
	UserLoginV2        Event = "/user/loginV2"
	UserImport         Event = "/user/import"
	UserPresent        Event = "/user/present"
	UserPresentV2      Event = "/user/presentV2"
	UserIsLoggedin     Event = "/user/isloggedin"
	UserLogout         Event = "/user/logout"
	UserExport         Event = "/user/export"
	UserDelete         Event = "/user/delete"
	UserStat           Event = "/user/stat"
	PodNew             Event = "/pod/new"
	PodOpen            Event = "/pod/open"
	PodClose           Event = "/pod/close"
	PodSync            Event = "/pod/sync"
	PodDelete          Event = "/pod/delete"
	PodLs              Event = "/pod/ls"
	PodStat            Event = "/pod/stat"
	PodShare           Event = "/pod/share"
	PodReceive         Event = "/pod/receive"
	PodReceiveInfo     Event = "/pod/receiveinfo"
	DirIsPresent       Event = "/dir/present"
	DirMkdir           Event = "/dir/mkdir"
	DirRmdir           Event = "/dir/rmdir"
	DirLs              Event = "/dir/ls"
	DirStat            Event = "/dir/stat"
	FileDownload       Event = "/file/download"
	FileDownloadStream Event = "/file/download/stream"
	FileUpload         Event = "/file/upload"
	FileUploadStream   Event = "/file/upload/stream"
	FileShare          Event = "/file/share"
	FileReceive        Event = "/file/receive"
	FileReceiveInfo    Event = "/file/receiveinfo"
	FileDelete         Event = "/file/delete"
	FileStat           Event = "/file/stat"
	KVCreate           Event = "/kv/new"
	KVList             Event = "/kv/ls"
	KVOpen             Event = "/kv/open"
	KVDelete           Event = "/kv/delete"
	KVCount            Event = "/kv/count"
	KVEntryPut         Event = "/kv/entry/put"
	KVEntryGet         Event = "/kv/entry/get"
	KVEntryDelete      Event = "/kv/entry/del"
	KVLoadCSV          Event = "/kv/loadcsv"
	KVLoadCSVStream    Event = "/kv/loadcsv/stream"
	KVSeek             Event = "/kv/seek"
	KVSeekNext         Event = "/kv/seek/next"
	DocCreate          Event = "/doc/new"
	DocList            Event = "/doc/ls"
	DocOpen            Event = "/doc/open"
	DocCount           Event = "/doc/count"
	DocDelete          Event = "/doc/delete"
	DocFind            Event = "/doc/find"
	DocEntryPut        Event = "/doc/entry/put"
	DocEntryGet        Event = "/doc/entry/get"
	DocEntryDel        Event = "/doc/entry/del"
	DocLoadJson        Event = "/doc/loadjson"
	DocLoadJsonStream  Event = "/doc/loadjson/stream"
	DocIndexJson       Event = "/doc/indexjson"
)

type WebsocketRequest struct {
	Id     string      `json:"_id"`
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
	Id          string      `json:"_id"`
	Event       Event       `json:"event"`
	Params      interface{} `json:"params,omitempty"`
	StatusCode  int         `json:"code,omitempty"`
	buf         bytes.Buffer
	contentType string
}

func NewWebsocketResponse() *WebsocketResponse {
	return &WebsocketResponse{}
}

func (w *WebsocketResponse) Write(bytes []byte) (int, error) {
	return w.buf.Write(bytes)
}

func (w *WebsocketResponse) WriteJson(bytes []byte) (int, error) {
	w.contentType = "json"
	body := map[string]interface{}{}
	err := json.Unmarshal(bytes, &body)
	if err != nil {
		return 0, err
	}
	w.Params = body
	return len(bytes), nil
}

func (w *WebsocketResponse) Marshal() []byte {
	if w.contentType == "json" {
		data, _ := json.Marshal(w)
		return data
	}
	return w.buf.Bytes()
}
