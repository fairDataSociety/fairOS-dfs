package common

import (
	"bytes"
	"encoding/json"
)

// Event is a string that represents a websocket event
type Event string

var (
	// UserSignup is the event for user signup
	UserSignup Event = "/user/signup"
	// UserLogin is the event for user login
	UserLogin Event = "/user/login"
	// UserIsLoggedin is the event for checking if a user is logged in
	UserIsLoggedin Event = "/user/isloggedin"
	// UserLogout is the event for user logout
	UserLogout Event = "/user/logout"
	// UserDelete is the event for user delete
	UserDelete Event = "/user/delete"
	// UserPresent is the event for checking if a user is present
	UserPresent Event = "/user/present"
	// UserStat is the event for user stat
	UserStat Event = "/user/stat"
	// PodNew is the event for pod new
	PodNew Event = "/pod/new"
	// PodOpen is the event for pod open
	PodOpen Event = "/pod/open"
	// PodClose is the event for pod close
	PodClose Event = "/pod/close"
	// PodSync is the event for pod sync
	PodSync Event = "/pod/sync"
	// PodDelete is the event for pod delete
	PodDelete Event = "/pod/delete"
	// PodLs is the event for listing all the pods
	PodLs Event = "/pod/ls"
	// PodStat is the event for pod stat
	PodStat Event = "/pod/stat"
	// PodShare is the event for pod share
	PodShare Event = "/pod/share"
	// PodReceive is the event for pod receive with sharingReference
	PodReceive Event = "/pod/receive"
	// PodReceiveInfo is the event for receive info of a pod from sharingReference
	PodReceiveInfo Event = "/pod/receiveinfo"
	// DirIsPresent is the event for checking if a directory is present
	DirIsPresent Event = "/dir/present"
	// DirMkdir is the event for making a directory
	DirMkdir Event = "/dir/mkdir"
	// DirRename is the event for renaming a directory
	DirRename Event = "/dir/rename"
	// DirRmdir is the event for removing a directory
	DirRmdir Event = "/dir/rmdir"
	// DirLs is the event for listing content in the directory
	DirLs Event = "/dir/ls"
	// DirStat is the event for directory stat
	DirStat Event = "/dir/stat"
	// FileDownload is the event for downloading a file
	FileDownload Event = "/file/download"
	// FileDownloadStream is the event for downloading a file stream
	FileDownloadStream Event = "/file/download/stream"
	// FileUpload is the event for uploading a file
	FileUpload Event = "/file/upload"
	// FileUploadStream is the event for uploading a file stream
	FileUploadStream Event = "/file/upload/stream"
	// FileShare is the event for sharing a file
	FileShare Event = "/file/share"
	// FileReceive is the event for receiving a file from sharingReference
	FileReceive Event = "/file/receive"
	// FileRename is the event for renaming a file
	FileRename Event = "/file/rename"
	// FileReceiveInfo is the event for receive info of a file from sharingReference
	FileReceiveInfo Event = "/file/receiveinfo"
	// FileDelete is the event for deleting a file
	FileDelete Event = "/file/delete"
	// FileStat is the event for file stat
	FileStat Event = "/file/stat"
	// KVCreate is the event for creating a KV store
	KVCreate Event = "/kv/new"
	// KVList is the event for listing all the KV stores
	KVList Event = "/kv/ls"
	// KVOpen is the event for opening a KV store
	KVOpen Event = "/kv/open"
	// KVDelete is the event for deleting a KV store
	KVDelete Event = "/kv/delete"
	// KVCount is the event for counting the number of entries in a KV store
	KVCount Event = "/kv/count"
	// KVEntryPresent is the event for checking if an entry is present in a KV store
	KVEntryPresent Event = "/kv/entry/present"
	// KVEntryPut is the event for putting an entry in a KV store
	KVEntryPut Event = "/kv/entry/put"
	// KVEntryGet is the event for getting an entry from a KV store
	KVEntryGet Event = "/kv/entry/get"
	// KVEntryDelete is the event for deleting an entry from a KV store
	KVEntryDelete Event = "/kv/entry/del"
	// KVLoadCSV is the event for loading a CSV file into a KV store
	KVLoadCSV Event = "/kv/loadcsv"
	// KVLoadCSVStream is the event for loading a CSV file into a KV store
	KVLoadCSVStream Event = "/kv/loadcsv/stream"
	// KVSeek is the event for seeking to a key in a KV store
	KVSeek Event = "/kv/seek"
	// KVSeekNext is the event for seeking to the next key in a KV store
	KVSeekNext Event = "/kv/seek/next"
	// DocCreate is the event for creating a document store
	DocCreate Event = "/doc/new"
	// DocList is the event for listing all the document stores
	DocList Event = "/doc/ls"
	// DocOpen is the event for opening a document store
	DocOpen Event = "/doc/open"
	// DocCount is the event for counting the number of documents in a document store
	DocCount Event = "/doc/count"
	// DocDelete is the event for deleting a document store
	DocDelete Event = "/doc/delete"
	// DocFind is the event for finding documents in a document store
	DocFind Event = "/doc/find"
	// DocEntryPut is the event for putting a document in a document store
	DocEntryPut Event = "/doc/entry/put"
	// DocEntryGet is the event for getting a document from a document store
	DocEntryGet Event = "/doc/entry/get"
	// DocEntryDel is the event for deleting a document from a document store
	DocEntryDel Event = "/doc/entry/del"
	// DocLoadJson is the event for loading a JSON file into a document store
	DocLoadJson Event = "/doc/loadjson"
	// DocLoadJsonStream is the event for loading a JSON file into a document store
	DocLoadJsonStream Event = "/doc/loadjson/stream"
	// DocIndexJson is the event for indexing a JSON file already in the pod into a document store
	DocIndexJson Event = "/doc/indexjson"
)

// WebsocketRequest is the request sent to the websocket
type WebsocketRequest struct {
	Id     string      `json:"_id"`
	Event  Event       `json:"event"`
	Params interface{} `json:"params,omitempty"`
}

// FileRequest is the request for file operations
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

// FileDownloadRequest is the request for file download
type FileDownloadRequest struct {
	PodName  string `json:"podName,omitempty"`
	Filepath string `json:"filePath,omitempty"`
}

// WebsocketResponse is the response sent from the websocket
type WebsocketResponse struct {
	Id          string      `json:"_id"`
	Event       Event       `json:"event"`
	Params      interface{} `json:"params,omitempty"`
	StatusCode  int         `json:"code,omitempty"`
	buf         bytes.Buffer
	contentType string
}

// NewWebsocketResponse creates a new WebsocketResponse
func NewWebsocketResponse() *WebsocketResponse {
	return &WebsocketResponse{}
}

func (w *WebsocketResponse) Write(bytes []byte) (int, error) {
	return w.buf.Write(bytes)
}

// WriteJson writes the json bytes to the response
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

// Marshal marshals the response
func (w *WebsocketResponse) Marshal() []byte {
	if w.contentType == "json" {
		data, _ := json.Marshal(w)
		return data
	}
	return w.buf.Bytes()
}
