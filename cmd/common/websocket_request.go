package common

import (
	"bytes"
	"encoding/json"
)

// Event
type Event string

var (
	//UserSignup
	UserSignup Event = "/user/signup"
	//UserLogin
	UserLogin Event = "/user/login"
	//UserLoginV2
	UserLoginV2 Event = "/user/loginV2"
	//UserPresentV2
	UserPresentV2 Event = "/user/presentV2"
	//UserIsLoggedin
	UserIsLoggedin Event = "/user/isloggedin"
	//UserLogout
	UserLogout Event = "/user/logout"
	//UserDelete
	UserDelete Event = "/user/delete"
	//UserStat
	UserStat Event = "/user/stat"
	//PodNew
	PodNew Event = "/pod/new"
	//PodOpen
	PodOpen Event = "/pod/open"
	//PodClose
	PodClose Event = "/pod/close"
	//PodSync
	PodSync Event = "/pod/sync"
	//PodDelete
	PodDelete Event = "/pod/delete"
	//PodLs
	PodLs Event = "/pod/ls"
	//PodStat
	PodStat Event = "/pod/stat"
	//PodShare
	PodShare Event = "/pod/share"
	//PodReceive
	PodReceive Event = "/pod/receive"
	//PodReceiveInfo
	PodReceiveInfo Event = "/pod/receiveinfo"
	//DirIsPresent
	DirIsPresent Event = "/dir/present"
	//DirMkdir
	DirMkdir Event = "/dir/mkdir"
	//DirRename
	DirRename Event = "/dir/rename"
	//DirRmdir
	DirRmdir Event = "/dir/rmdir"
	//DirLs
	DirLs Event = "/dir/ls"
	//DirStat
	DirStat Event = "/dir/stat"
	//FileDownload
	FileDownload Event = "/file/download"
	//FileDownloadStream
	FileDownloadStream Event = "/file/download/stream"
	//FileUpload
	FileUpload Event = "/file/upload"
	//FileUploadStream
	FileUploadStream Event = "/file/upload/stream"
	//FileShare
	FileShare Event = "/file/share"
	//FileReceive
	FileReceive Event = "/file/receive"
	//FileRename
	FileRename Event = "/file/rename"
	//FileReceiveInfo
	FileReceiveInfo Event = "/file/receiveinfo"
	//FileDelete
	FileDelete Event = "/file/delete"
	//FileStat
	FileStat Event = "/file/stat"
	//KVCreate
	KVCreate Event = "/kv/new"
	//KVList
	KVList Event = "/kv/ls"
	//KVOpen
	KVOpen Event = "/kv/open"
	//KVDelete
	KVDelete Event = "/kv/delete"
	//KVCount
	KVCount Event = "/kv/count"
	//KVEntryPresent
	KVEntryPresent Event = "/kv/entry/present"
	//KVEntryPut
	KVEntryPut Event = "/kv/entry/put"
	//KVEntryGet
	KVEntryGet Event = "/kv/entry/get"
	//KVEntryDelete
	KVEntryDelete Event = "/kv/entry/del"
	//KVLoadCSV
	KVLoadCSV Event = "/kv/loadcsv"
	//KVLoadCSVStream
	KVLoadCSVStream Event = "/kv/loadcsv/stream"
	//KVSeek
	KVSeek Event = "/kv/seek"
	//KVSeekNext
	KVSeekNext Event = "/kv/seek/next"
	//DocCreate
	DocCreate Event = "/doc/new"
	//DocList
	DocList Event = "/doc/ls"
	//DocOpen
	DocOpen Event = "/doc/open"
	//DocCount
	DocCount Event = "/doc/count"
	//DocDelete
	DocDelete Event = "/doc/delete"
	//DocFind
	DocFind Event = "/doc/find"
	//DocEntryPut
	DocEntryPut Event = "/doc/entry/put"
	//DocEntryGet
	DocEntryGet Event = "/doc/entry/get"
	//DocEntryDel
	DocEntryDel Event = "/doc/entry/del"
	//DocLoadJson
	DocLoadJson Event = "/doc/loadjson"
	//DocLoadJsonStream
	DocLoadJsonStream Event = "/doc/loadjson/stream"
	//DocIndexJson
	DocIndexJson Event = "/doc/indexjson"
)

// WebsocketRequest
type WebsocketRequest struct {
	Id     string      `json:"_id"`
	Event  Event       `json:"event"`
	Params interface{} `json:"params,omitempty"`
}

// FileRequest
type FileRequest struct {
	PodName       string `json:"podName,omitempty"`
	TableName     string `json:"tableName,omitempty"`
	DirPath       string `json:"dirPath,omitempty"`
	BlockSize     string `json:"blockSize,omitempty"`
	FileName      string `json:"fileName,omitempty"`
	ContentLength string `json:"contentLength,omitempty"`
	Compression   string `json:"compression,omitempty"`
	Overwrite     bool   `json:"overwrite,omitempty"`
}

// FileDownloadRequest
type FileDownloadRequest struct {
	PodName  string `json:"podName,omitempty"`
	Filepath string `json:"filePath,omitempty"`
}

// WebsocketResponse
type WebsocketResponse struct {
	Id          string      `json:"_id"`
	Event       Event       `json:"event"`
	Params      interface{} `json:"params,omitempty"`
	StatusCode  int         `json:"code,omitempty"`
	buf         bytes.Buffer
	contentType string
}

// NewWebsocketResponse
func NewWebsocketResponse() *WebsocketResponse {
	return &WebsocketResponse{}
}

func (w *WebsocketResponse) Write(bytes []byte) (int, error) {
	return w.buf.Write(bytes)
}

// WriteJson
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

// Marshal
func (w *WebsocketResponse) Marshal() []byte {
	if w.contentType == "json" {
		data, _ := json.Marshal(w)
		return data
	}
	return w.buf.Bytes()
}
