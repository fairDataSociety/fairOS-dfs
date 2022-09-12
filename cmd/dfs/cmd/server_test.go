package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
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
	"github.com/sirupsen/logrus"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			return string(b)
		}
		b[i] = letterRunes[num.Int64()]
	}
	return string(b)
}

func TestApis(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()
	logger := logging.New(io.Discard, logrus.ErrorLevel)
	dataDir, err := ioutil.TempDir("", "new")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dataDir)
	users := user.NewUsers(dataDir, mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger, dataDir)
	handler = api.NewMockHandler(dfsApi, logger)
	httpPort = ":9090"
	pprofPort = ":9091"
	basev1 := "http://localhost:9090/v1"
	basev2 := "http://localhost:9090/v2"
	go startHttpService(logger)

	// wait 10 seconds for the server to start
	<-time.After(time.Second * 10)
	t.Run("login-fail-test", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}
		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}
		userBytes, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}
		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogin)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userLoginHttpReq.Header.Add("Content-Type", "application/json")
		userLoginHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userLoginResp, err := c.Do(userLoginHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		err = userLoginResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if userLoginResp.StatusCode == http.StatusOK {
			t.Fatal("user should not be able to login")
		}
	})

	t.Run("signup-login", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}
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
		signupRequestResp, err := c.Do(signupRequestDataHttpReq)
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

		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogin)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userLoginHttpReq.Header.Add("Content-Type", "application/json")
		userLoginHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userLoginResp, err := c.Do(userLoginHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		err = userLoginResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if userLoginResp.StatusCode != http.StatusOK {
			t.Fatal("user should be able to login")
		}
	})

	t.Run("signup-login-logout-loggedin", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}
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
		signupRequestResp, err := c.Do(signupRequestDataHttpReq)
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

		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogin)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userLoginHttpReq.Header.Add("Content-Type", "application/json")
		userLoginHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userLoginResp, err := c.Do(userLoginHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		err = userLoginResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if userLoginResp.StatusCode != http.StatusOK {
			t.Fatal("user should be able to login")
		}

		cookie := userLoginResp.Header["Set-Cookie"]
		userLogoutHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogout)), http.NoBody)
		if err != nil {
			t.Fatal(err)

		}
		userLogoutHttpReq.Header.Set("Cookie", cookie[0])
		userLogoutResp, err := c.Do(userLogoutHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = userLogoutResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		userIsLoggedInHttpReq, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s%s?username%s", basev1, string(common.UserIsLoggedin), userRequest.UserName), http.NoBody)
		if err != nil {
			t.Fatal(err)
		}
		userIsLoggedInResp, err := c.Do(userIsLoggedInHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		useIsLoggedBodyBytes, err := io.ReadAll(userIsLoggedInResp.Body)
		if err != nil {
			t.Fatal(err)
		}

		err = userIsLoggedInResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		loggedInStatus := &api.LoginStatus{}
		err = json.Unmarshal(useIsLoggedBodyBytes, loggedInStatus)
		if err != nil {
			t.Fatal(err)
		}

		if loggedInStatus.LoggedIn {
			t.Fatal("user should be logged out")
		}
	})
}
