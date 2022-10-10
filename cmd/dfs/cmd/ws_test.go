package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func TestWsConnection(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()
	logger := logging.New(io.Discard, logrus.DebugLevel)
	dataDir, err := os.MkdirTemp("", "new")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dataDir)
	users := user.NewUsers(dataDir, mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger, dataDir)
	handler = api.NewMockHandler(dfsApi, logger)

	httpPort = ":9090"
	base := "localhost:9090"
	basev2 := "http://localhost:9090/v2"
	go startHttpService(logger)

	// wait 10 seconds for the server to start
	<-time.After(time.Second * 10)

	u := url.URL{Scheme: "ws", Host: base, Path: "/ws/v1/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	defer c.Close()

	downloadFn := func(cl string) {
		mt2, reader, err := c.NextReader()
		if mt2 != websocket.BinaryMessage {
			t.Fatal("non binary message while download")
		}
		if err != nil {
			t.Fatal("download failed", err)
		}
		fo, err := os.Create(fmt.Sprintf("./%d", time.Now().Unix()))
		if err != nil {
			t.Fatal("download failed", err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				t.Fatal("download failed", err)
			}
		}()
		n, err := io.Copy(fo, reader)
		if err != nil {
			t.Fatal("download failed", err)
		}
		if fmt.Sprintf("%d", n) == cl {
			t.Log("download finished ")
			return
		}
	}

	go func() {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			switch mt {
			case 1:
				res := &common.WebsocketResponse{}
				if err := json.Unmarshal(message, res); err != nil {
					t.Error("got error ", err)
					continue
				}
				if res.Event == common.FileDownload {
					params := res.Params.(map[string]interface{})
					cl := fmt.Sprintf("%v", params["content_length"])
					downloadFn(cl)
					continue
				}
				if res.StatusCode != 200 && res.StatusCode != 201 {
					t.Errorf("%s failed: %s\n", res.Event, res.Params)
					continue
				}
				t.Logf("%s ran successfully : %s\n", res.Event, res.Params)
			}
		}
	}()
	t.Run("ws test", func(t *testing.T) {
		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}

		userBytes, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}

		signupRequestDataHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserSignup)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)
		}
		signupRequestDataHttpReq.Header.Add("Content-Type", "application/json")
		signupRequestDataHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))

		httpClient := http.Client{Timeout: time.Duration(1) * time.Minute}
		signupRequestResp, err := httpClient.Do(signupRequestDataHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = signupRequestResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if signupRequestResp.StatusCode != http.StatusCreated {
			t.Fatal("Signup failed", signupRequestResp.StatusCode)
		}

		// userLogin
		podName := fmt.Sprintf("%d", time.Now().UnixNano())

		login := &common.WebsocketRequest{
			Event:  common.UserLoginV2,
			Params: userRequest,
		}

		data, err := json.Marshal(login)
		if err != nil {
			t.Fatal("failed to marshal login request: ", err)
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			t.Fatal("write:", err)
		}

		// userPresent
		uPresent := &common.WebsocketRequest{
			Event: common.UserPresentV2,
			Params: common.UserRequest{
				UserName: userRequest.UserName,
			},
		}
		data, err = json.Marshal(uPresent)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// userLoggedIN
		uLoggedIn := &common.WebsocketRequest{
			Event: common.UserIsLoggedin,
			Params: common.UserRequest{
				UserName: userRequest.UserName,
			},
		}
		data, err = json.Marshal(uLoggedIn)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// userStat
		userStat := &common.WebsocketRequest{
			Event: common.UserStat,
		}
		data, err = json.Marshal(userStat)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// podNew
		podNew := &common.WebsocketRequest{
			Event: common.PodNew,
			Params: common.PodRequest{
				PodName:  podName,
				Password: userRequest.Password,
			},
		}
		data, err = json.Marshal(podNew)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// podLs
		podLs := &common.WebsocketRequest{
			Event: common.PodLs,
		}
		data, err = json.Marshal(podLs)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// mkdir
		mkDir := &common.WebsocketRequest{
			Event: common.DirMkdir,
			Params: common.FileRequest{
				PodName: podName,
				DirPath: "/d",
			},
		}
		data, err = json.Marshal(mkDir)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// rmDir
		rmDir := &common.WebsocketRequest{
			Event: common.DirRmdir,
			Params: common.FileRequest{
				PodName: podName,
				DirPath: "/d",
			},
		}
		data, err = json.Marshal(rmDir)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// dirLs
		dirLs := &common.WebsocketRequest{
			Event: common.DirLs,
			Params: common.FileRequest{
				PodName: podName,
				DirPath: "/",
			},
		}
		data, err = json.Marshal(dirLs)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// dirStat
		dirStat := &common.WebsocketRequest{
			Event: common.DirStat,
			Params: common.FileRequest{
				PodName: podName,
				DirPath: "/",
			},
		}
		data, err = json.Marshal(dirStat)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// dirPresent
		dirPresent := &common.WebsocketRequest{
			Event: common.DirIsPresent,
			Params: common.FileRequest{
				PodName: podName,
				DirPath: "/d",
			},
		}
		data, err = json.Marshal(dirPresent)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		//Upload
		upload := &common.WebsocketRequest{
			Event: common.FileUpload,
			Params: common.FileRequest{
				PodName:   podName,
				DirPath:   "/",
				BlockSize: "1Mb",
				FileName:  "README.md",
			},
		}
		data, err = json.Marshal(upload)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}
		file, err := os.Open("../../../README.md")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		body := &bytes.Buffer{}
		_, err = io.Copy(body, file)
		if err != nil {
			panic(err)
		}
		err = c.WriteMessage(websocket.BinaryMessage, body.Bytes())
		if err != nil {
			log.Println("write:", err)
			return
		}

		//Download
		download := &common.WebsocketRequest{
			Event: common.FileDownload,
			Params: common.FileDownloadRequest{
				PodName:  podName,
				Filepath: "/README.md",
			},
		}
		data, err = json.Marshal(download)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// stat
		stat := &common.WebsocketRequest{
			Event: common.FileStat,
			Params: common.FileSystemRequest{
				PodName:       podName,
				DirectoryPath: "/README.md",
			},
		}
		data, err = json.Marshal(stat)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		table := "kv_1"
		// kvCreate
		kvCreate := &common.WebsocketRequest{
			Event: common.KVCreate,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
				IndexType: "string",
			},
		}
		data, err = json.Marshal(kvCreate)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvList
		kvList := &common.WebsocketRequest{
			Event: common.KVList,
			Params: common.KVRequest{
				PodName: podName,
			},
		}
		data, err = json.Marshal(kvList)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvOpen
		kvOpen := &common.WebsocketRequest{
			Event: common.KVOpen,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
			},
		}
		data, err = json.Marshal(kvOpen)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		//kvEntryPut
		kvEntryPut := &common.WebsocketRequest{
			Event: common.KVEntryPut,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
				Key:       "key1",
				Value:     "value1",
			},
		}
		data, err = json.Marshal(kvEntryPut)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvCount
		kvCount := &common.WebsocketRequest{
			Event: common.KVCount,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
			},
		}
		data, err = json.Marshal(kvCount)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvGet
		kvGet := &common.WebsocketRequest{
			Event: common.KVEntryGet,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
				Key:       "key1",
			},
		}
		data, err = json.Marshal(kvGet)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvSeek
		kvSeek := &common.WebsocketRequest{
			Event: common.KVSeek,
			Params: common.KVRequest{
				PodName:     podName,
				TableName:   table,
				StartPrefix: "key",
			},
		}
		data, err = json.Marshal(kvSeek)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvSeek
		kvSeekNext := &common.WebsocketRequest{
			Event: common.KVSeekNext,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
			},
		}
		data, err = json.Marshal(kvSeekNext)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// kvEntryDel
		kvEntryDel := &common.WebsocketRequest{
			Event: common.KVEntryDelete,
			Params: common.KVRequest{
				PodName:   podName,
				TableName: table,
				Key:       "key1",
			},
		}
		data, err = json.Marshal(kvEntryDel)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		docTable := "doc_1"
		// docCreate
		docCreate := &common.WebsocketRequest{
			Event: common.DocCreate,
			Params: common.DocRequest{
				PodName:     podName,
				TableName:   docTable,
				SimpleIndex: "first_name=string,age=number",
				Mutable:     true,
			},
		}
		data, err = json.Marshal(docCreate)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docLs
		docLs := &common.WebsocketRequest{
			Event: common.DocList,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
			},
		}
		data, err = json.Marshal(docLs)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docOpen
		docOpen := &common.WebsocketRequest{
			Event: common.DocOpen,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
			},
		}
		data, err = json.Marshal(docOpen)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docEntryPut
		docEntryPut := &common.WebsocketRequest{
			Event: common.DocEntryPut,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
				Document:  `{"id":"1", "first_name": "Hello1", "age": 11}`,
			},
		}
		data, err = json.Marshal(docEntryPut)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docEntryGet
		docEntryGet := &common.WebsocketRequest{
			Event: common.DocEntryGet,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
				ID:        "1",
			},
		}
		data, err = json.Marshal(docEntryGet)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docFind
		docFind := &common.WebsocketRequest{
			Event: common.DocFind,
			Params: common.DocRequest{
				PodName:    podName,
				TableName:  docTable,
				Expression: `age>10`,
			},
		}
		data, err = json.Marshal(docFind)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docCount
		docCount := &common.WebsocketRequest{
			Event: common.DocCount,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
			},
		}
		data, err = json.Marshal(docCount)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docEntryGet
		docEntryDel := &common.WebsocketRequest{
			Event: common.DocEntryDel,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
				ID:        "1",
			},
		}
		data, err = json.Marshal(docEntryDel)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// docDel
		docDel := &common.WebsocketRequest{
			Event: common.DocDelete,
			Params: common.DocRequest{
				PodName:   podName,
				TableName: docTable,
			},
		}
		data, err = json.Marshal(docDel)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}
		// user Logout
		uLogout := &common.WebsocketRequest{
			Event: common.UserLogout,
		}
		data, err = json.Marshal(uLogout)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// userLoggedIN
		uLoggedIn = &common.WebsocketRequest{
			Event: common.UserIsLoggedin,
			Params: common.UserRequest{
				UserName: userRequest.UserName,
			},
		}
		data, err = json.Marshal(uLoggedIn)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}

		err = c.WriteMessage(websocket.CloseMessage, []byte{})
		if err != nil {
			t.Fatal("write:", err)
		}
		// wait
		<-time.After(time.Second)
	})
}
