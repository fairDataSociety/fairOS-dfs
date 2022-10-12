package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

//const (
//	wsChunkLimit = 1000000
//)

var (
	readDeadline = 4 * time.Second

	writeDeadline = 4 * time.Second
)

func (h *Handler) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{} // use default options
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
		//origin := r.Header.Get("Origin")
		//for _, v := range h.whitelistedOrigins {
		//	if origin == v {
		//		return true
		//	}
		//}
		//return false
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("Error during connection upgrade:", err)
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
				h.logger.Debugf("ws event handler: failed to send ping: %v", err)
				h.logger.Error("ws event handler: failed to send ping")
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

	var sessionID string
	logEventDescription := func(url string, startTime time.Time, status int, logger logging.Logger) {
		fields := logrus.Fields{
			"uri":      url,
			"duration": time.Since(startTime).String(),
			"status":   status,
		}
		logger.WithFields(fields).Log(logrus.DebugLevel, "ws event response: ")
	}
	respondWithError := func(response *common.WebsocketResponse, originalErr error) {
		response.StatusCode = http.StatusInternalServerError
		if originalErr == nil {
			return
		}
		if err := conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			return
		}

		message := map[string]interface{}{}
		message["message"] = originalErr.Error()

		messageBytes, err := json.Marshal(message)
		if err != nil {
			return
		}
		_, err = response.WriteJson(messageBytes)
		if err != nil {
			return
		}
		if err := conn.SetWriteDeadline(time.Now().Add(writeDeadline)); err != nil {
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, response.Marshal()); err != nil {
			h.logger.Debugf("ws event handler: failed to write error response: %v", err)
			h.logger.Error("ws event handler: failed to write error response")
			return
		}
		logEventDescription(string(response.Event), time.Now(), response.StatusCode, h.logger)
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
		res.Id = req.Id
		res.Event = req.Event
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			continue
		}
		switch req.Event {
		// user related events
		case common.UserLoginV2:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			loginRequest := &common.UserRequest{}
			err = json.Unmarshal(jsonBytes, loginRequest)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			ui, nameHash, publicKey, err := h.dfsAPI.LoginUserV2(loginRequest.UserName, loginRequest.Password, "")
			if err != nil {
				respondWithError(res, err)
				continue
			}
			sessionID = ui.GetSessionId()
			loginResponse := &UserSignupResponse{
				NameHash:  nameHash,
				PublicKey: publicKey,
			}
			resBytes, err := json.Marshal(loginResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserLogin), to, http.StatusOK, h.logger)
		case common.UserPresentV2:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			request := &common.UserRequest{}
			err = json.Unmarshal(jsonBytes, request)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			presentResponse := &PresentResponse{
				Present: h.dfsAPI.IsUserNameAvailableV2(request.UserName),
			}
			resBytes, err := json.Marshal(presentResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserPresentV2), to, res.StatusCode, h.logger)
		case common.UserIsLoggedin:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			request := &common.UserRequest{}
			err = json.Unmarshal(jsonBytes, request)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			loggedInResponse := &LoginStatus{
				LoggedIn: h.dfsAPI.IsUserLoggedIn(request.UserName),
			}
			resBytes, err := json.Marshal(loggedInResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserIsLoggedin), to, res.StatusCode, h.logger)
		case common.UserLogout:
			err := h.dfsAPI.LogoutUser(sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "user logged out successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserLogout), to, res.StatusCode, h.logger)
		case common.UserDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			request := &common.UserRequest{}
			err = json.Unmarshal(jsonBytes, request)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DeleteUserV2(request.Password, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "user deleted successfully"

			resBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserDelete), to, res.StatusCode, h.logger)
		case common.UserStat:
			userStat, err := h.dfsAPI.GetUserStat(sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			resBytes, err := json.Marshal(userStat)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.UserStat), to, res.StatusCode, h.logger)
		// pod related events
		case common.PodReceive:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			request := &common.PodReceiveRequest{}
			err = json.Unmarshal(jsonBytes, request)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			ref, err := utils.ParseHexReference(request.Reference)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			pi, err := h.dfsAPI.PodReceive(sessionID, request.PodName, ref)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = fmt.Sprintf("public pod %q, added as shared pod", pi.GetPodName())

			resBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodReceive), to, res.StatusCode, h.logger)
		case common.PodReceiveInfo:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			request := &common.PodReceiveRequest{}
			err = json.Unmarshal(jsonBytes, request)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			ref, err := utils.ParseHexReference(request.Reference)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			shareInfo, err := h.dfsAPI.PodReceiveInfo(sessionID, ref)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			resBytes, err := json.Marshal(shareInfo)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodReceiveInfo), to, res.StatusCode, h.logger)
		case common.PodNew:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			_, err = h.dfsAPI.CreatePod(podReq.PodName, podReq.Password, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "pod created successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodNew), to, res.StatusCode, h.logger)
		case common.PodOpen:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			_, err = h.dfsAPI.OpenPod(podReq.PodName, podReq.Password, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "pod opened successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodOpen), to, res.StatusCode, h.logger)
		case common.PodClose:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			err = h.dfsAPI.ClosePod(podReq.PodName, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "pod closed successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodClose), to, res.StatusCode, h.logger)
		case common.PodSync:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			err = h.dfsAPI.SyncPod(podReq.PodName, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "pod synced successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodSync), to, res.StatusCode, h.logger)
		case common.PodShare:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			sharedPodName := podReq.SharedPodName
			if sharedPodName == "" {
				sharedPodName = podReq.PodName
			}
			sharingRef, err := h.dfsAPI.PodShare(podReq.PodName, sharedPodName, podReq.Password, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			response := &PodSharingReference{
				Reference: sharingRef,
			}
			resBytes, err := json.Marshal(response)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodShare), to, res.StatusCode, h.logger)
		case common.PodDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			err = h.dfsAPI.DeletePod(podReq.PodName, podReq.Password, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "pod deleted successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodDelete), to, res.StatusCode, h.logger)
		case common.PodLs:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			pods, sharedPods, err := h.dfsAPI.ListPods(sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			if pods == nil {
				pods = make([]string, 0)
			}
			if sharedPods == nil {
				sharedPods = make([]string, 0)
			}
			listResponse := &PodListResponse{
				Pods:       pods,
				SharedPods: sharedPods,
			}
			resBytes, err := json.Marshal(listResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(resBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodLs), to, res.StatusCode, h.logger)
		case common.PodStat:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podReq := &common.PodRequest{}
			err = json.Unmarshal(jsonBytes, podReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			stat, err := h.dfsAPI.PodStat(podReq.PodName, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			podStatRenponse := &PodStatResponse{
				PodName:    stat.PodName,
				PodAddress: stat.PodAddress,
			}

			messageBytes, err := json.Marshal(podStatRenponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.PodStat), to, res.StatusCode, h.logger)

		// file related events
		case common.DirMkdir:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.Mkdir(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "directory created successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DirMkdir), to, res.StatusCode, h.logger)
		case common.DirRmdir:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.RmDir(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "directory removed successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DirRmdir), to, res.StatusCode, h.logger)
		case common.DirLs:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			dEntries, fEntries, err := h.dfsAPI.ListDir(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			if dEntries == nil {
				dEntries = make([]dir.Entry, 0)
			}
			if fEntries == nil {
				fEntries = make([]file.Entry, 0)
			}
			listResponse := &ListFileResponse{
				Directories: dEntries,
				Files:       fEntries,
			}
			messageBytes, err := json.Marshal(listResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DirLs), to, res.StatusCode, h.logger)
		case common.DirStat:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			ds, err := h.dfsAPI.DirectoryStat(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			messageBytes, err := json.Marshal(ds)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DirStat), to, res.StatusCode, h.logger)
		case common.DirIsPresent:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			present, err := h.dfsAPI.IsDirPresent(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			presentResponse := &DirPresentResponse{
				Present: present,
			}
			messageBytes, err := json.Marshal(presentResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DirIsPresent), to, res.StatusCode, h.logger)
		//case common.FileDownloadStream:
		//	jsonBytes, _ := json.Marshal(req.Params)
		//	args := make(map[string]string)
		//	if err := json.Unmarshal(jsonBytes, &args); err != nil {
		//		h.logger.Debugf("ws event handler: download: failed to read params: %v", err)
		//		h.logger.Error("ws event handler: download: failed to read params")
		//		respondWithError(res, err)
		//		continue
		//	}
		//	body := new(bytes.Buffer)
		//	writer := multipart.NewWriter(body)
		//	for k, v := range args {
		//		err := writer.WriteField(k, v)
		//		if err != nil {
		//			h.logger.Debugf("ws event handler: download: failed to write fields in form: %v", err)
		//			h.logger.Error("ws event handler: download: failed to write fields in form")
		//			respondWithError(res, err)
		//			continue
		//		}
		//	}
		//	err = writer.Close()
		//	if err != nil {
		//		h.logger.Debugf("ws event handler: download: failed to close writer: %v", err)
		//		h.logger.Error("ws event handler: download: failed to close writer")
		//		respondWithError(res, err)
		//		continue
		//	}
		//	httpReq, err := newMultipartRequest(http.MethodPost, string(common.FileDownload), writer.Boundary(), body)
		//	if err != nil {
		//		respondWithError(res, err)
		//		continue
		//	}
		//	h.FileDownloadHandler(res, httpReq)
		//	if res.StatusCode != 0 {
		//		errMessage := res.Params.(map[string]interface{})
		//		respondWithError(res, fmt.Errorf("%s", errMessage["message"]))
		//		continue
		//	}
		//	downloadConfirmResponse := common.NewWebsocketResponse()
		//	downloadConfirmResponse.Event = common.FileDownloadStream
		//	downloadConfirmResponse.Header().Set("Content-Type", "application/json; charset=utf-8")
		//	if res.Header().Get("Content-Length") != "" {
		//		dlMessage := map[string]string{}
		//		dlMessage["content_length"] = res.Header().Get("Content-Length")
		//		dlMessage["file_name"] = filepath.Base(args["file_path"])
		//		data, _ := json.Marshal(dlMessage)
		//		_, err = downloadConfirmResponse.Write(data)
		//		if err != nil {
		//			h.logger.Debugf("ws event handler: download: failed to send download confirm: %v", err)
		//			h.logger.Error("ws event handler: download: failed to send download confirm")
		//			continue
		//		}
		//	}
		//	downloadConfirmResponse.WriteHeader(http.StatusOK)
		//	if err := conn.WriteMessage(messageType, downloadConfirmResponse.Marshal()); err != nil {
		//		h.logger.Debugf("ws event handler: download: failed to write in connection: %v", err)
		//		h.logger.Error("ws event handler: download: failed to write in connection")
		//		continue
		//	}
		//	if res.StatusCode == 0 {
		//		messageType = websocket.BinaryMessage
		//		data := res.Marshal()
		//		head := 0
		//		tail := len(data)
		//		for head+wsChunkLimit < tail {
		//			if err := conn.WriteMessage(messageType, data[head:(head+wsChunkLimit)]); err != nil {
		//				h.logger.Debugf("ws event handler: response: failed to write in connection: %v", err)
		//				h.logger.Error("ws event handler: response: failed to write in connection")
		//				return err
		//			}
		//			head += wsChunkLimit
		//		}
		//		if err := conn.WriteMessage(messageType, data[head:tail]); err != nil {
		//			h.logger.Debugf("ws event handler: response: failed to write in connection: %v", err)
		//			h.logger.Error("ws event handler: response: failed to write in connection")
		//			return err
		//		}
		//	}
		//	messageType = websocket.TextMessage
		//	res.Header().Set("Content-Type", "application/json; charset=utf-8")
		//	if res.Header().Get("Content-Length") != "" {
		//		dlFinishedMessage := map[string]string{}
		//		dlFinishedMessage["message"] = "download finished"
		//		data, _ := json.Marshal(dlFinishedMessage)
		//		_, err = res.Write(data)
		//		if err != nil {
		//			h.logger.Debugf("ws event handler: download: failed to send download confirm: %v", err)
		//			h.logger.Error("ws event handler: download: failed to send download confirm")
		//			continue
		//		}
		//		res.WriteHeader(http.StatusOK)
		//	}
		//	logEventDescription(string(common.FileDownloadStream), to, res.StatusCode, h.logger)
		case common.FileDownload:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileDownloadRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			data, n, err := h.dfsAPI.DownloadFile(fsReq.PodName, fsReq.Filepath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(data)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			data.Close()
			downloadConfirmResponse := common.NewWebsocketResponse()
			downloadConfirmResponse.Event = common.FileDownload
			downloadConfirmResponse.Id = res.Id
			dlMessage := map[string]string{}
			dlMessage["content_length"] = fmt.Sprintf("%d", n)
			dlMessage["file_name"] = filepath.Base(fsReq.Filepath)
			dsRes, _ := json.Marshal(dlMessage)
			_, err = downloadConfirmResponse.WriteJson(dsRes)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			downloadConfirmResponse.StatusCode = http.StatusOK
			if err := conn.WriteMessage(messageType, downloadConfirmResponse.Marshal()); err != nil {
				respondWithError(res, err)
				continue
			}

			res.StatusCode = http.StatusOK
			_, err = res.Write(buf.Bytes())
			if err != nil {
				respondWithError(res, err)
				continue
			}
			messageType = websocket.BinaryMessage
			//if err := conn.WriteMessage(messageType, res.Marshal()); err != nil {
			//	respondWithError(res, err)
			//	return err
			//}
			//fmt.Println(5)
			//
			//messageType = websocket.TextMessage
			//dlFinishedMessage := map[string]string{}
			//dlFinishedMessage["message"] = "download finished"
			//finishedRes, _ := json.Marshal(dlFinishedMessage)
			//res.StatusCode = http.StatusOK
			//_, err = res.WriteJson(finishedRes)
			//if err := conn.WriteMessage(messageType, res.Marshal()); err != nil {
			//	respondWithError(res, err)
			//	return err
			//}
			//fmt.Println(6)

			logEventDescription(string(common.FileDownload), to, res.StatusCode, h.logger)
		case common.FileUpload, common.FileUploadStream:
			streaming := false
			if req.Event == common.FileUploadStream {
				streaming = true
			}
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileRequest{}
			if err := json.Unmarshal(jsonBytes, fsReq); err != nil {
				respondWithError(res, err)
				continue
			}

			fileName := fsReq.FileName
			compression := strings.ToLower(fsReq.Compression)
			contentLength := fsReq.ContentLength

			data := &bytes.Buffer{}
			if streaming {
				if contentLength == "" || contentLength == "0" {
					respondWithError(res, fmt.Errorf("streaming needs \"content_length\""))
					continue
				}
				var totalRead int64 = 0
				for {
					mt, reader, err := conn.NextReader()
					if err != nil {
						respondWithError(res, err)
						continue
					}
					if mt != websocket.BinaryMessage {
						respondWithError(res, fmt.Errorf("file content should be as binary message"))
						continue
					}
					n, err := io.Copy(data, reader)
					if err != nil {
						respondWithError(res, err)
						continue
					}
					totalRead += n
					if fmt.Sprintf("%d", totalRead) == contentLength {
						h.logger.Debug("streamed full content")
						break
					}
				}
			} else {
				mt, reader, err := conn.NextReader()
				if err != nil {
					respondWithError(res, err)
					continue
				}
				if mt != websocket.BinaryMessage {
					respondWithError(res, fmt.Errorf("file content should be as binary message"))
					continue
				}
				_, err = io.Copy(data, reader)
				if err != nil {
					respondWithError(res, err)
					continue
				}
			}
			bs, err := humanize.ParseBytes(fsReq.BlockSize)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.UploadFile(fsReq.PodName, fileName, sessionID, int64(len(data.Bytes())), data, fsReq.DirPath, compression, uint32(bs))
			if err != nil {
				respondWithError(res, err)
				continue
			}
			responses := &UploadResponse{FileName: fileName, Message: "uploaded successfully"}
			messageBytes, err := json.Marshal(responses)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileUpload), to, res.StatusCode, h.logger)
		case common.FileShare:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			sharingRef, err := h.dfsAPI.ShareFile(fsReq.PodName, fsReq.DirectoryPath, fsReq.Destination, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsShareResponse := &FileSharingReference{
				Reference: sharingRef,
			}
			messageBytes, err := json.Marshal(fsShareResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileShare), to, res.StatusCode, h.logger)
		case common.FileReceive:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileReceiveRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			sharingRef, err := utils.ParseSharingReference(fsReq.SharingReference)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			filePath, err := h.dfsAPI.ReceiveFile(fsReq.PodName, fsReq.DirectoryPath, sharingRef, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReceiveResponse := &ReceiveFileResponse{
				FileName: filePath,
			}
			messageBytes, err := json.Marshal(fsReceiveResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileReceive), to, res.StatusCode, h.logger)
		case common.FileReceiveInfo:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileReceiveRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			sharingRef, err := utils.ParseSharingReference(fsReq.SharingReference)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			receiveInfo, err := h.dfsAPI.ReceiveInfo(fsReq.PodName, sessionID, sharingRef)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			messageBytes, err := json.Marshal(receiveInfo)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileReceiveInfo), to, res.StatusCode, h.logger)
		case common.FileDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DeleteFile(fsReq.PodName, fsReq.FilePath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "file deleted successfully"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileDelete), to, res.StatusCode, h.logger)
		case common.FileStat:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			fsReq := &common.FileSystemRequest{}
			err = json.Unmarshal(jsonBytes, fsReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			stat, err := h.dfsAPI.FileStat(fsReq.PodName, fsReq.DirectoryPath, sessionID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			messageBytes, err := json.Marshal(stat)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.FileStat), to, res.StatusCode, h.logger)

		// kv related events
		case common.KVCreate:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			idxType := kvReq.IndexType
			if idxType == "" {
				idxType = "string"
			}

			var indexType collection.IndexType
			switch idxType {
			case "string":
				indexType = collection.StringIndex
			case "number":
				indexType = collection.NumberIndex
			case "bytes":
			default:
				respondWithError(res, fmt.Errorf("kv create: invalid \"indexType\" "))
				continue
			}
			err = h.dfsAPI.KVCreate(sessionID, kvReq.PodName, kvReq.TableName, indexType)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "kv store created"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVCreate), to, res.StatusCode, h.logger)
		case common.KVList:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			collections, err := h.dfsAPI.KVList(sessionID, kvReq.PodName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var col Collections
			for k, v := range collections {
				m := Collection{
					Name:           k,
					IndexedColumns: v,
					CollectionType: "KV Store",
				}
				col.Tables = append(col.Tables, m)
			}

			messageBytes, err := json.Marshal(col)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVList), to, res.StatusCode, h.logger)
		case common.KVOpen:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			err = h.dfsAPI.KVOpen(sessionID, kvReq.PodName, kvReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "kv store created"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVOpen), to, res.StatusCode, h.logger)
		case common.KVCount:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			count, err := h.dfsAPI.KVCount(sessionID, kvReq.PodName, kvReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			messageBytes, err := json.Marshal(count)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVCount), to, res.StatusCode, h.logger)
		case common.KVDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			err = h.dfsAPI.KVDelete(sessionID, kvReq.PodName, kvReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "kv store deleted"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVDelete), to, res.StatusCode, h.logger)
		case common.KVEntryPresent:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			presentResponse := &PresentResponse{
				Present: true,
			}
			_, _, err = h.dfsAPI.KVGet(sessionID, kvReq.PodName, kvReq.TableName, kvReq.Key)
			if err != nil {
				presentResponse.Present = false
			}
			messageBytes, err := json.Marshal(presentResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVEntryPresent), to, res.StatusCode, h.logger)
		case common.KVEntryPut:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.KVPut(sessionID, kvReq.PodName, kvReq.TableName, kvReq.Key, []byte(kvReq.Value))
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "key added"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVEntryPut), to, res.StatusCode, h.logger)
		case common.KVEntryGet:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			columns, data, err := h.dfsAPI.KVGet(sessionID, kvReq.PodName, kvReq.TableName, kvReq.Key)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var resp KVResponse
			if columns != nil {
				resp.Keys = columns
			} else {
				resp.Keys = []string{kvReq.Key}
			}
			resp.Values = data
			messageBytes, err := json.Marshal(resp)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVEntryGet), to, res.StatusCode, h.logger)
		case common.KVEntryDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			_, err = h.dfsAPI.KVDel(sessionID, kvReq.PodName, kvReq.TableName, kvReq.Key)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			message := map[string]interface{}{}
			message["message"] = "key deleted"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVEntryDelete), to, res.StatusCode, h.logger)
		//case common.KVLoadCSV, common.KVLoadCSVStream:
		//	streaming := false
		//	if req.Event == common.KVLoadCSVStream {
		//		streaming = true
		//	}
		//	httpReq, err := newMultipartRequestWithBinaryMessage(req.Params, "csv", http.MethodPost, string(req.Event), streaming)
		//	if err != nil {
		//		respondWithError(res, err)
		//		continue
		//	}
		//
		//	h.KVLoadCSVHandler(res, httpReq)
		//	logEventDescription(string(common.KVLoadCSV), to, res.StatusCode, h.logger)
		case common.KVSeek:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			if kvReq.Limit == "" {
				kvReq.Limit = DefaultSeekLimit
			}
			noOfRows, err := strconv.ParseInt(kvReq.Limit, 10, 64)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			_, err = h.dfsAPI.KVSeek(sessionID, kvReq.PodName, kvReq.TableName,
				kvReq.StartPrefix, kvReq.EndPrefix, noOfRows)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "seeked closest to the start key"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.KVSeek), to, res.StatusCode, h.logger)
		case common.KVSeekNext:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			kvReq := &common.KVRequest{}
			err = json.Unmarshal(jsonBytes, kvReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			columns, key, data, err := h.dfsAPI.KVGetNext(sessionID, kvReq.PodName, kvReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			resp := &KVResponse{}
			if columns != nil {
				resp.Keys = columns
			} else {
				resp.Keys = []string{key}
			}
			resp.Values = data

			messageBytes, err := json.Marshal(resp)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}

			logEventDescription(string(common.KVSeekNext), to, res.StatusCode, h.logger)

		// doc related events
		case common.DocCreate:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			indexes := make(map[string]collection.IndexType)
			si := docReq.SimpleIndex
			if si != "" {
				idxs := strings.Split(si, ",")
				for _, idx := range idxs {
					nt := strings.Split(idx, "=")
					if len(nt) != 2 {
						respondWithError(res, fmt.Errorf("doc  create: \"si\" invalid argument"))
						continue
					}
					switch nt[1] {
					case "string":
						indexes[nt[0]] = collection.StringIndex
					case "number":
						indexes[nt[0]] = collection.NumberIndex
					case "map":
						indexes[nt[0]] = collection.MapIndex
					case "list":
						indexes[nt[0]] = collection.ListIndex
					case "bytes":
					default:
						respondWithError(res, fmt.Errorf("doc create: invalid \"indexType\" "))
						continue
					}
				}
			}

			err = h.dfsAPI.DocCreate(sessionID, docReq.PodName, docReq.TableName,
				indexes, docReq.Mutable)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "document db created"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocCreate), to, res.StatusCode, h.logger)
		case common.DocList:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			collections, err := h.dfsAPI.DocList(sessionID, docReq.PodName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var col DocumentDBs
			for name, dbSchema := range collections {
				var indexes []collection.SIndex
				indexes = append(indexes, dbSchema.SimpleIndexes...)
				indexes = append(indexes, dbSchema.MapIndexes...)
				indexes = append(indexes, dbSchema.ListIndexes...)
				m := documentDB{
					Name:           name,
					IndexedColumns: indexes,
					CollectionType: "Document Store",
				}
				col.Tables = append(col.Tables, m)
			}
			messageBytes, err := json.Marshal(col)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocList), to, res.StatusCode, h.logger)
		case common.DocOpen:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DocOpen(sessionID, docReq.PodName, docReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "document store opened"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocOpen), to, res.StatusCode, h.logger)
		case common.DocCount:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			count, err := h.dfsAPI.DocCount(sessionID, docReq.PodName, docReq.TableName, docReq.Expression)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			messageBytes, err := json.Marshal(count)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocCount), to, res.StatusCode, h.logger)
		case common.DocDelete:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DocDelete(sessionID, docReq.PodName, docReq.TableName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "document store deleted"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocDelete), to, res.StatusCode, h.logger)
		case common.DocFind:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var limitInt int
			if docReq.Limit == "" {
				limitInt = 10
			} else {
				lmt, err := strconv.Atoi(docReq.Limit)
				if err != nil {
					respondWithError(res, fmt.Errorf("doc find: invalid value for argument \"limit\""))
					continue
				}
				limitInt = lmt
			}
			data, err := h.dfsAPI.DocFind(sessionID, docReq.PodName, docReq.TableName, docReq.Expression, limitInt)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var docs DocFindResponse
			docs.Docs = data
			messageBytes, err := json.Marshal(docs)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocFind), to, res.StatusCode, h.logger)
		case common.DocEntryPut:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DocPut(sessionID, docReq.PodName, docReq.TableName, []byte(docReq.Document))
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "added document to db"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocEntryPut), to, res.StatusCode, h.logger)
		case common.DocEntryGet:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			data, err := h.dfsAPI.DocGet(sessionID, docReq.PodName, docReq.TableName, docReq.ID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			var getResponse DocGetResponse
			getResponse.Doc = data

			messageBytes, err := json.Marshal(getResponse)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocEntryGet), to, res.StatusCode, h.logger)
		case common.DocEntryDel:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DocDel(sessionID, docReq.PodName, docReq.TableName, docReq.ID)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "deleted document from db"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocEntryDel), to, res.StatusCode, h.logger)
		//case common.DocLoadJson, common.DocLoadJsonStream:
		//	streaming := false
		//	if req.Event == common.DocLoadJsonStream {
		//		streaming = true
		//	}
		//	httpReq, err := newMultipartRequestWithBinaryMessage(req.Params, "json", http.MethodPost, string(req.Event), streaming)
		//	if err != nil {
		//		respondWithError(res, err)
		//		continue
		//	}
		//
		//	h.DocLoadJsonHandler(res, httpReq)
		//	logEventDescription(string(common.DocLoadJson), to, res.StatusCode, h.logger)
		case common.DocIndexJson:
			jsonBytes, err := json.Marshal(req.Params)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			docReq := &common.DocRequest{}
			err = json.Unmarshal(jsonBytes, docReq)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			err = h.dfsAPI.DocIndexJson(sessionID, docReq.PodName, docReq.TableName, docReq.FileName)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			message := map[string]interface{}{}
			message["message"] = "indexing started"

			messageBytes, err := json.Marshal(message)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			res.StatusCode = http.StatusOK
			_, err = res.WriteJson(messageBytes)
			if err != nil {
				respondWithError(res, err)
				continue
			}
			logEventDescription(string(common.DocIndexJson), to, res.StatusCode, h.logger)
		default:
			respondWithError(res, fmt.Errorf("unknown event"))
			continue
		}
		if err := conn.SetWriteDeadline(time.Now().Add(readDeadline)); err != nil {
			return err
		}
		if err := conn.WriteMessage(messageType, res.Marshal()); err != nil {
			h.logger.Debugf("ws event handler: response: failed to write in connection: %v", err)
			h.logger.Error("ws event handler: response: failed to write in connection")
			return err
		}
	}
}
