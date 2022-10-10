package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/sirupsen/logrus"
)

func TestWsConnection(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()
	logger := logging.New(os.Stdout, logrus.DebugLevel)
	dataDir, err := ioutil.TempDir("", "new")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dataDir)
	users := user.NewUsers(dataDir, mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger, dataDir)
	handler = api.NewMockHandler(dfsApi, logger)

	httpPort = ":9090"
	base := "localhost:9090"
	go startHttpService(logger)

	// wait 10 seconds for the server to start
	<-time.After(time.Second * 10)
	t.Run("login-fail-test", func(t *testing.T) {
		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}

		u := url.URL{Scheme: "ws", Host: base, Path: "/ws/v1/"}
		t.Log(u.String())
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Fatal("dial:", err)
		}
		defer c.Close()

		go func() {
			for {
				//cl := "0"
				mt, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", mt, err)
					return
				}
				switch mt {
				case 1:
					res := &common.WebsocketResponse{}
					if err := json.Unmarshal(message, res); err != nil {
						t.Log("got error ", err)
						continue
					}
					//if res.Event == common.FileDownload {
					//	params := res.Params.(map[string]interface{})
					//	t.Log("DOWNLOAD", params)
					//	cl = fmt.Sprintf("%v", params["content_length"])
					//	downloadFn(cl)
					//	continue
					//}
					if res.StatusCode != 200 && res.StatusCode != 201 {
						fmt.Printf("%s failed: %s\n", res.Event, res.Params)
						continue
					}
					fmt.Printf("%s ran successfully : %s\n", res.Event, res.Params)
				}
			}
		}()

		// userSignup
		sighup := &common.WebsocketRequest{
			Event:  common.UserSignup,
			Params: userRequest,
		}

		data, err := json.Marshal(sighup)
		if err != nil {
			log.Println("Marshal:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write:", err)
			return
		}
	})
}
