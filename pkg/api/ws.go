package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	readDeadline = 4 * time.Second

	writeDeadline = 4 * time.Second
)

func (h *Handler) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{} // use default options
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	err = h.handleEvents(conn)
	if err != nil {
		h.logger.Errorf("Error during handling event:", err)
		return
	}
}

func (h *Handler) handleEvents(conn *websocket.Conn) error {
	defer conn.Close()

	err := conn.SetReadDeadline(time.Now().Add(readDeadline))
	if err != nil {
		h.logger.Debugf("ws event handler: set read deadline failed on connection : %v", err)
		h.logger.Error("ws event handler: set read deadline failed on connection")
		return err
	}

	// keep pinging to check pong
	go func() {
		pingPeriod := (readDeadline * 9) / 10
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				h.logger.Debugf("ws event handler: upload: failed to send ping: %v", err)
				h.logger.Error("ws event handler: upload: failed to send ping")
				return
			}
		}
	}()

	// add read deadline in pong
	conn.SetPongHandler(func(message string) error {
		if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			h.logger.Debugf("ws event handler: set read deadline failed on connection : %v", err)
			h.logger.Error("ws event handler: set read deadline failed on connection")
			return err
		}
		return nil
	})

	var cookie []string

	newRequest := func(method, url string, buf []byte) (*http.Request, error) {
		httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Add("Content-Type", "application/json")
		httpReq.Header.Add("Content-Length", strconv.Itoa(len(buf)))
		if cookie != nil {
			httpReq.Header.Set("Cookie", cookie[0])
		}
		return httpReq, nil
	}

	newMultipartRequestWithBinaryMessage := func(params map[string]interface{}, formField, method, url string) (*http.Request, error) {
		jsonBytes, _ := json.Marshal(params)
		args := make(map[string]string)
		if err := json.Unmarshal(jsonBytes, &args); err != nil {
			h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to read params: %v", err)
			h.logger.Error("ws event handler: multipart rqst w/ body: failed to read params")
			return nil, err
		}
		mt, reader, err := conn.NextReader()
		if mt != websocket.BinaryMessage {
			h.logger.Warning("non binary message in loadcsv")
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		fileName := ""
		compression := ""
		// Add parameters
		for k, v := range args {
			if k == "file_name" {
				fileName = v
			}
			if k == "compression" {
				compression = strings.ToLower(compression)
			}
			err := writer.WriteField(k, v)
			if err != nil {
				h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to write fields in form: %v", err)
				h.logger.Error("ws event handler: multipart rqst w/ body: failed to write fields in form")
				return nil, err
			}
		}

		part, err := writer.CreateFormFile(formField, fileName)
		if err != nil {
			h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to create files field in form: %v", err)
			h.logger.Error("ws event handler: multipart rqst w/ body: failed to create files field in form")
			return nil, err
		}
		_, err = io.Copy(part, reader)
		if err != nil {
			h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to read file: %v", err)
			h.logger.Error("ws event handler: multipart rqst w/ body: failed to read file")
			return nil, err
		}
		err = writer.Close()
		if err != nil {
			h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to close writer: %v", err)
			h.logger.Error("ws event handler: multipart rqst w/ body: failed to close writer")
			return nil, err
		}

		httpReq, err := http.NewRequest(method, url, body)
		if err != nil {
			h.logger.Debugf("ws event handler: multipart rqst w/ body: failed to create http request: %v", err)
			h.logger.Error("ws event handler: multipart rqst w/ body: failed to create http request")
			return nil, err
		}
		contentType := fmt.Sprintf("multipart/form-data;boundary=%v", writer.Boundary())
		httpReq.Header.Set("Content-Type", contentType)
		if cookie != nil {
			httpReq.Header.Set("Cookie", cookie[0])
		}
		if compression != "" {
			httpReq.Header.Set(CompressionHeader, compression)
		}
		return httpReq, nil
	}

	newMultipartRequest := func(method, url, boundary string, r io.Reader) (*http.Request, error) {
		httpReq, err := http.NewRequest(method, url, r)
		if err != nil {
			return nil, err
		}
		contentType := fmt.Sprintf("multipart/form-data;boundary=%v", boundary)
		httpReq.Header.Set("Content-Type", contentType)
		if cookie != nil {
			httpReq.Header.Set("Cookie", cookie[0])
		}
		return httpReq, nil
	}

	respondWithError := func(response *common.WebsocketResponse, err error) {
		if err == nil {
			return
		}
		if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			return
		}

		message := map[string]interface{}{}
		message["message"] = err.Error()
		response.Body = &message
		response.StatusCode = http.StatusInternalServerError

		if err := conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, response.Marshal())
		if err != nil {
			h.logger.Debugf("ws event handler: upload: failed to write error response: %v", err)
			h.logger.Error("ws event handler: upload: failed to write error response")
			return
		}
	}

	makeQueryParams := func(base string, params map[string]interface{}) string {
		url := base + "?"
		for i, v := range params {
			url = fmt.Sprintf("%s%s=%s&", url, i, v)
		}
		return url
	}

	logEventDescription := func(url string, startTime time.Time, status int, logger logging.Logger) {
		fields := logrus.Fields{
			"uri":      url,
			"duration": time.Since(startTime).String(),
			"status":   status,
		}
		logger.WithFields(fields).Log(logrus.DebugLevel, "ws event response: ")
	}

	for {
		res := common.NewWebsocketResponse()
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Debugf("ws event handler: read message error: %v", err)
			h.logger.Error("ws event handler: read message error")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return err
			}
			return nil
		}
		to := time.Now()
		req := &common.WebsocketRequest{}
		err = json.Unmarshal(message, req)
		if err != nil {
			h.logger.Debugf("ws event handler: failed to read request: %v", err)
			h.logger.Error("ws event handler: failed to read request")
			if err := conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
				return err
			}
			return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
		}
		res.Event = req.Event
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			continue
		}
		switch req.Event {
		// user related events
		case common.UserSignup:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.UserSignup), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserSignupHandler(res, httpReq)
			cookie = res.Header()["Set-Cookie"]
			logEventDescription(string(common.UserSignup), to, res.StatusCode, h.logger)
		case common.UserLogin:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.UserLogin), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserLoginHandler(res, httpReq)
			cookie = res.Header()["Set-Cookie"]
			logEventDescription(string(common.UserLogin), to, res.StatusCode, h.logger)
		case common.UserImport:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.UserImport), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.ImportUserHandler(res, httpReq)
			logEventDescription(string(common.UserImport), to, res.StatusCode, h.logger)
		case common.UserPresent:
			url := makeQueryParams(string(common.UserPresent), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserPresentHandler(res, httpReq)
			logEventDescription(string(common.UserPresent), to, res.StatusCode, h.logger)
		case common.UserIsLoggedin:
			url := makeQueryParams(string(common.UserIsLoggedin), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.IsUserLoggedInHandler(res, httpReq)
			logEventDescription(string(common.UserIsLoggedin), to, res.StatusCode, h.logger)
		case common.UserLogout:
			httpReq, err := newRequest(http.MethodPost, string(common.UserLogout), nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserLogoutHandler(res, httpReq)
			logEventDescription(string(common.UserLogout), to, res.StatusCode, h.logger)
		case common.UserExport:
			httpReq, err := newRequest(http.MethodPost, string(common.UserExport), nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.ExportUserHandler(res, httpReq)
			logEventDescription(string(common.UserExport), to, res.StatusCode, h.logger)
		case common.UserDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.UserDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserDeleteHandler(res, httpReq)
			logEventDescription(string(common.UserDelete), to, res.StatusCode, h.logger)
		case common.UserStat:
			httpReq, err := newRequest(http.MethodGet, string(common.UserStat), nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.UserStatHandler(res, httpReq)
			logEventDescription(string(common.UserStat), to, res.StatusCode, h.logger)
		// pod related events
		case common.PodReceive:
			url := makeQueryParams(string(common.PodReceive), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodReceiveHandler(res, httpReq)
			logEventDescription(string(common.PodReceive), to, res.StatusCode, h.logger)
		case common.PodReceiveInfo:
			url := makeQueryParams(string(common.PodReceiveInfo), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodReceiveInfoHandler(res, httpReq)
			logEventDescription(string(common.PodReceiveInfo), to, res.StatusCode, h.logger)
		case common.PodNew:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodNew), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodCreateHandler(res, httpReq)
			logEventDescription(string(common.PodNew), to, res.StatusCode, h.logger)
		case common.PodOpen:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodOpen), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodOpenHandler(res, httpReq)
			logEventDescription(string(common.PodOpen), to, res.StatusCode, h.logger)
		case common.PodClose:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodClose), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodCloseHandler(res, httpReq)
			logEventDescription(string(common.PodClose), to, res.StatusCode, h.logger)
		case common.PodSync:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodSync), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodSyncHandler(res, httpReq)
			logEventDescription(string(common.PodSync), to, res.StatusCode, h.logger)
		case common.PodShare:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodShare), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodShareHandler(res, httpReq)
			logEventDescription(string(common.PodShare), to, res.StatusCode, h.logger)
		case common.PodDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.PodDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodDeleteHandler(res, httpReq)
			logEventDescription(string(common.PodDelete), to, res.StatusCode, h.logger)
		case common.PodLs:
			httpReq, err := newRequest(http.MethodGet, string(common.PodLs), nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodListHandler(res, httpReq)
			logEventDescription(string(common.PodLs), to, res.StatusCode, h.logger)
		case common.PodStat:
			url := makeQueryParams(string(common.UserPresent), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.PodStatHandler(res, httpReq)
			logEventDescription(string(common.PodStat), to, res.StatusCode, h.logger)

		// file related events
		case common.DirMkdir:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DirMkdir), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DirectoryMkdirHandler(res, httpReq)
			logEventDescription(string(common.DirMkdir), to, res.StatusCode, h.logger)
		case common.DirRmdir:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.DirRmdir), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DirectoryRmdirHandler(res, httpReq)
			logEventDescription(string(common.DirRmdir), to, res.StatusCode, h.logger)
		case common.DirLs:
			url := makeQueryParams(string(common.DirLs), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DirectoryLsHandler(res, httpReq)
			logEventDescription(string(common.DirLs), to, res.StatusCode, h.logger)
		case common.DirStat:
			url := makeQueryParams(string(common.DirStat), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DirectoryStatHandler(res, httpReq)
			logEventDescription(string(common.DirStat), to, res.StatusCode, h.logger)
		case common.DirIsPresent:
			url := makeQueryParams(string(common.DirIsPresent), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DirectoryPresentHandler(res, httpReq)
			logEventDescription(string(common.DirIsPresent), to, res.StatusCode, h.logger)
		case common.FileDownload:
			jsonBytes, _ := json.Marshal(req.Params)
			args := make(map[string]string)
			if err := json.Unmarshal(jsonBytes, &args); err != nil {
				h.logger.Debugf("ws event handler: download: failed to read params: %v", err)
				h.logger.Error("ws event handler: download: failed to read params")
				respondWithError(res, err)
				continue
			}
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			for k, v := range args {
				err := writer.WriteField(k, v)
				if err != nil {
					h.logger.Debugf("ws event handler: download: failed to write fields in form: %v", err)
					h.logger.Error("ws event handler: download: failed to write fields in form")
					respondWithError(res, err)
					continue
				}
			}
			err = writer.Close()
			if err != nil {
				h.logger.Debugf("ws event handler: download: failed to close writer: %v", err)
				h.logger.Error("ws event handler: download: failed to close writer")
				respondWithError(res, err)
				continue
			}
			httpReq, err := newMultipartRequest(http.MethodPost, string(common.FileUpload), writer.Boundary(), body)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileDownloadHandler(res, httpReq)
			downloadConfirmResponse := common.NewWebsocketResponse()
			downloadConfirmResponse.Event = common.FileDownload
			downloadConfirmResponse.Header().Set("Content-Type", "application/json; charset=utf-8")
			if res.Header().Get("Content-Length") != "" {
				dlMessage := map[string]string{}
				dlMessage["content_length"] = res.Header().Get("Content-Length")
				data, _ := json.Marshal(dlMessage)
				_, err = downloadConfirmResponse.Write(data)
				if err != nil {
					h.logger.Debugf("ws event handler: download: failed to send download confirm: %v", err)
					h.logger.Error("ws event handler: download: failed to send download confirm")
					continue
				}
			}
			downloadConfirmResponse.WriteHeader(http.StatusOK)
			if err := conn.WriteMessage(messageType, downloadConfirmResponse.Marshal()); err != nil {
				h.logger.Debugf("ws event handler: download: failed to write in connection: %v", err)
				h.logger.Error("ws event handler: download: failed to write in connection")
				continue
			}
			messageType = websocket.BinaryMessage
			logEventDescription(string(common.FileDownload), to, res.StatusCode, h.logger)
		case common.FileUpload:
			httpReq, err := newMultipartRequestWithBinaryMessage(req.Params, "files", http.MethodPost, string(common.FileUpload))
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileUploadHandler(res, httpReq)
			logEventDescription(string(common.FileUpload), to, res.StatusCode, h.logger)
		case common.FileShare:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.FileShare), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileShareHandler(res, httpReq)
			logEventDescription(string(common.FileShare), to, res.StatusCode, h.logger)
		case common.FileReceive:
			url := makeQueryParams(string(common.FileReceive), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileReceiveHandler(res, httpReq)
			logEventDescription(string(common.FileReceive), to, res.StatusCode, h.logger)
		case common.FileReceiveInfo:
			url := makeQueryParams(string(common.FileReceiveInfo), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileReceiveInfoHandler(res, httpReq)
			logEventDescription(string(common.FileReceiveInfo), to, res.StatusCode, h.logger)
		case common.FileDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.FileDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileDeleteHandler(res, httpReq)
			logEventDescription(string(common.FileDelete), to, res.StatusCode, h.logger)
		case common.FileStat:
			url := makeQueryParams(string(common.FileStat), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.FileStatHandler(res, httpReq)
			logEventDescription(string(common.FileStat), to, res.StatusCode, h.logger)

		// kv related events
		case common.KVCreate:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.KVCreate), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVCreateHandler(res, httpReq)
			logEventDescription(string(common.KVCreate), to, res.StatusCode, h.logger)
		case common.KVList:
			url := makeQueryParams(string(common.KVList), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVListHandler(res, httpReq)
			logEventDescription(string(common.KVList), to, res.StatusCode, h.logger)
		case common.KVOpen:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.KVOpen), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVOpenHandler(res, httpReq)
			logEventDescription(string(common.KVOpen), to, res.StatusCode, h.logger)
		case common.KVCount:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.KVCount), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVCountHandler(res, httpReq)
			logEventDescription(string(common.KVCount), to, res.StatusCode, h.logger)
		case common.KVDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.KVDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVDeleteHandler(res, httpReq)
			logEventDescription(string(common.KVDelete), to, res.StatusCode, h.logger)
		case common.KVEntryPut:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.KVEntryPut), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVPutHandler(res, httpReq)
			logEventDescription(string(common.KVEntryPut), to, res.StatusCode, h.logger)
		case common.KVEntryGet:
			url := makeQueryParams(string(common.KVEntryGet), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVGetHandler(res, httpReq)
			logEventDescription(string(common.KVEntryGet), to, res.StatusCode, h.logger)
		case common.KVEntryDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.KVEntryDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVDelHandler(res, httpReq)
			logEventDescription(string(common.KVEntryDelete), to, res.StatusCode, h.logger)
		case common.KVLoadCSV:
			httpReq, err := newMultipartRequestWithBinaryMessage(req.Params, "csv", http.MethodPost, string(common.KVLoadCSV))
			if err != nil {
				respondWithError(res, err)
				continue
			}

			h.KVLoadCSVHandler(res, httpReq)
			logEventDescription(string(common.KVLoadCSV), to, res.StatusCode, h.logger)
		case common.KVSeek:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.KVSeek), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVSeekHandler(res, httpReq)
			logEventDescription(string(common.KVSeek), to, res.StatusCode, h.logger)
		case common.KVSeekNext:
			url := makeQueryParams(string(common.KVSeekNext), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.KVGetNextHandler(res, httpReq)
			logEventDescription(string(common.KVSeekNext), to, res.StatusCode, h.logger)

		// doc related events
		case common.DocCreate:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DocCreate), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocCreateHandler(res, httpReq)
			logEventDescription(string(common.DocCreate), to, res.StatusCode, h.logger)
		case common.DocList:
			url := makeQueryParams(string(common.DocList), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocListHandler(res, httpReq)
			logEventDescription(string(common.DocList), to, res.StatusCode, h.logger)
		case common.DocOpen:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DocOpen), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocOpenHandler(res, httpReq)
			logEventDescription(string(common.DocOpen), to, res.StatusCode, h.logger)
		case common.DocCount:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DocCount), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocCountHandler(res, httpReq)
			logEventDescription(string(common.DocCount), to, res.StatusCode, h.logger)
		case common.DocDelete:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.DocDelete), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocDeleteHandler(res, httpReq)
			logEventDescription(string(common.DocDelete), to, res.StatusCode, h.logger)
		case common.DocFind:
			url := makeQueryParams(string(common.DocFind), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocFindHandler(res, httpReq)
			logEventDescription(string(common.DocFind), to, res.StatusCode, h.logger)
		case common.DocEntryPut:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DocEntryPut), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocPutHandler(res, httpReq)
			logEventDescription(string(common.DocEntryPut), to, res.StatusCode, h.logger)
		case common.DocEntryGet:
			url := makeQueryParams(string(common.DocEntryGet), req.Params)
			httpReq, err := newRequest(http.MethodGet, url, nil)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocGetHandler(res, httpReq)
			logEventDescription(string(common.DocEntryGet), to, res.StatusCode, h.logger)
		case common.DocEntryDel:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodDelete, string(common.DocEntryDel), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocDelHandler(res, httpReq)
			logEventDescription(string(common.DocEntryDel), to, res.StatusCode, h.logger)
		case common.DocLoadJson:
			httpReq, err := newMultipartRequestWithBinaryMessage(req.Params, "json", http.MethodPost, string(common.DocLoadJson))
			if err != nil {
				respondWithError(res, err)
				continue
			}

			h.DocLoadJsonHandler(res, httpReq)
			logEventDescription(string(common.DocLoadJson), to, res.StatusCode, h.logger)
		case common.DocIndexJson:
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.DocIndexJson), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			h.DocIndexJsonHandler(res, httpReq)
			logEventDescription(string(common.DocIndexJson), to, res.StatusCode, h.logger)
		}
		if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			return err
		}
		if err := conn.WriteMessage(messageType, res.Marshal()); err != nil {
			h.logger.Debugf("ws event handler: response: failed to write in connection: %v", err)
			h.logger.Error("ws event handler: response: failed to write in connection")
			return err
		}
	}
}
