package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/gorilla/websocket"
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
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeDeadline))
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					h.logger.Debugf("ws event handler: upload: failed to send ping: %v", err)
					h.logger.Error("ws event handler: upload: failed to send ping")
					return
				}
			}
		}
	}()

	// add read deadline in pong
	conn.SetPongHandler(func(message string) error {
		err := conn.SetReadDeadline(time.Now().Add(readDeadline))
		if err != nil {
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
		conn.SetReadDeadline(time.Now().Add(readDeadline))

		message := map[string]interface{}{}
		message["message"] = err.Error()
		response.Body = &message
		response.StatusCode = http.StatusInternalServerError

		conn.SetWriteDeadline(time.Now().Add(writeDeadline))
		err = conn.WriteMessage(websocket.TextMessage, response.Marshal())
		if err != nil {
			h.logger.Debugf("ws event handler: upload: failed to write error response: %v", err)
			h.logger.Error("ws event handler: upload: failed to write error response")
			return
		}
	}

	makeQueryParams := func(base string, params map[string]interface{}) string {
		url := string(base) + "?"
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
			conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
		}
		res.Event = req.Event
		conn.SetReadDeadline(time.Time{})
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
			h.PodOpenHandler(res, httpReq)
			logEventDescription(string(common.PodNew), to, res.StatusCode, h.logger)
		case common.PodOpen:
			conn.SetReadDeadline(time.Time{})
			jsonBytes, _ := json.Marshal(req.Params)
			httpReq, err := newRequest(http.MethodPost, string(common.PodOpen), jsonBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			conn.SetReadDeadline(time.Now().Add(readDeadline))
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
		case common.FileUpload:
			jsonBytes, _ := json.Marshal(req.Params)
			args := make(map[string]string)
			err = json.Unmarshal(jsonBytes, &args)
			if err != nil {
				h.logger.Debugf("ws event handler: upload: failed to read params: %v", err)
				h.logger.Error("ws event handler: upload: failed to read params")
				respondWithError(res, err)
				continue
			}
			mt, reader, err := conn.NextReader()
			if mt != websocket.BinaryMessage {
				h.logger.Warning("non binary message in file upload")
				respondWithError(res, errors.New("non binary message in file upload"))
				continue
			}
			compression := ""
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			fileName := ""
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
					h.logger.Debugf("ws event handler: upload: failed to write fields in form: %v", err)
					h.logger.Error("ws event handler: upload: failed to write fields in form")
					respondWithError(res, err)
					continue
				}
			}

			part, err := writer.CreateFormFile("files", fileName)
			if err != nil {
				h.logger.Debugf("ws event handler: upload: failed to create files field in form: %v", err)
				h.logger.Error("ws event handler: upload: failed to create files field in form")
				respondWithError(res, err)
				continue
			}
			_, err = io.Copy(part, reader)
			if err != nil {
				h.logger.Debugf("ws event handler: upload: failed to read file: %v", err)
				h.logger.Error("ws event handler: upload: failed to read file")
				respondWithError(res, err)
				continue
			}
			err = writer.Close()
			if err != nil {
				h.logger.Debugf("ws event handler: upload: failed to close writer: %v", err)
				h.logger.Error("ws event handler: upload: failed to close writer")
				respondWithError(res, err)
				continue
			}

			httpReq, err := newMultipartRequest(http.MethodPost, string(common.FileUpload), writer.Boundary(), body)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			if compression != "" {
				httpReq.Header.Set(CompressionHeader, compression)
			}
			h.FileUploadHandler(res, httpReq)
			logEventDescription(string(common.FileUpload), to, res.StatusCode, h.logger)
		}
		conn.SetReadDeadline(time.Now().Add(readDeadline))
		err = conn.WriteMessage(messageType, res.Marshal())
		if err != nil {
			h.logger.Debugf("ws event handler: upload: failed to write in connection: %v", err)
			h.logger.Error("ws event handler: upload: failed to write in connection")
			return err
		}
	}
}
