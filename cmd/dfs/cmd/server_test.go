package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"mime/multipart"
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

	t.Run("signup-login-migrate", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}

		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}
		userBytes, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}
		signupRequestDataHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserSignup)), bytes.NewBuffer(userBytes))
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
			t.Fatal(err)
		}

		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserLogin)), bytes.NewBuffer(userBytes))
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

		migrateRequest := &common.UserRequest{Password: userRequest.Password}
		migrateRequestData, err := json.Marshal(migrateRequest)
		if err != nil {
			t.Fatal(err)
		}
		userMigrateHttpReq, err := http.NewRequest(http.MethodPost, basev2+"/user/migrate", bytes.NewBuffer(migrateRequestData))
		if err != nil {
			t.Fatal(err)

		}
		userMigrateHttpReq.Header.Add("Content-Type", "application/json")
		userMigrateHttpReq.Header.Add("Content-Length", strconv.Itoa(len(migrateRequestData)))
		userMigrateHttpReq.Header.Set("Cookie", cookie[0])
		useMigrateResp, err := c.Do(userMigrateHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = useMigrateResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if useMigrateResp.StatusCode != http.StatusOK {
			t.Fatal("user should be migrated")
		}
	})

	t.Run("signup-login-migrate-new-username", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}

		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}

		userBytes, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}

		//create v2 user
		signupRequestDataHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserSignup)), bytes.NewBuffer(userBytes))
		if err != nil {
			return
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

		// create user with same name in v1 to test migration with different name
		signupRequestDataHttpReq, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserSignup)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)
		}
		signupRequestDataHttpReq.Header.Add("Content-Type", "application/json")
		signupRequestDataHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		signupRequestResp, err = c.Do(signupRequestDataHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = signupRequestResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if signupRequestResp.StatusCode != http.StatusCreated {
			t.Fatal("Signup failed", signupRequestResp.StatusCode)
			return
		}

		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserLogin)), bytes.NewBuffer(userBytes))
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
		userStatHttpReq, err := http.NewRequest(http.MethodGet, basev1+"/user/stat", http.NoBody)
		if err != nil {
			t.Fatal(err)

		}
		userStatHttpReq.Header.Set("Cookie", cookie[0])
		userStatResp, err := c.Do(userStatHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		useStateBodyBytes, err := io.ReadAll(userStatResp.Body)
		if err != nil {
			t.Fatal(err)
		}

		err = userStatResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		userStat := &user.Stat{}
		err = json.Unmarshal(useStateBodyBytes, userStat)
		if err != nil {
			t.Fatal(err)
		}
		userMigrateHttpReq, err := http.NewRequest(http.MethodPost, basev2+"/user/migrate", bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userMigrateHttpReq.Header.Add("Content-Type", "application/json")
		userMigrateHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userMigrateHttpReq.Header.Set("Cookie", cookie[0])
		useMigrateResp, err := c.Do(userMigrateHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = useMigrateResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if useMigrateResp.StatusCode == http.StatusOK {
			t.Fatal("migration should fail as username is already taken")
		}
		newUsername := randStringRunes(16)
		migrateRequest := &common.UserRequest{Password: userRequest.Password, UserName: newUsername}
		migrateRequestData, err := json.Marshal(migrateRequest)
		if err != nil {
			t.Fatal(err)
		}
		userMigrateHttpReq, err = http.NewRequest(http.MethodPost, basev2+"/user/migrate", bytes.NewBuffer(migrateRequestData))
		if err != nil {
			t.Fatal(err)

		}
		userMigrateHttpReq.Header.Add("Content-Type", "application/json")
		userMigrateHttpReq.Header.Add("Content-Length", strconv.Itoa(len(migrateRequestData)))
		userMigrateHttpReq.Header.Set("Cookie", cookie[0])
		useMigrateResp, err = c.Do(userMigrateHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = useMigrateResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if useMigrateResp.StatusCode != http.StatusOK {
			t.Fatal("user should be migrated")
		}

		userLoginHttpReq, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogin)), bytes.NewBuffer(migrateRequestData))
		if err != nil {
			t.Fatal(err)

		}
		userLoginHttpReq.Header.Add("Content-Type", "application/json")
		userLoginHttpReq.Header.Add("Content-Length", strconv.Itoa(len(migrateRequestData)))
		userLoginResp, err = c.Do(userLoginHttpReq)
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
		cookie = userLoginResp.Header["Set-Cookie"]

		userStatHttpReq, err = http.NewRequest(http.MethodGet, "http://localhost:9090/v1/user/stat", http.NoBody)
		if err != nil {
			t.Fatal(err)

		}
		userStatHttpReq.Header.Set("Cookie", cookie[0])
		userStatResp, err = c.Do(userStatHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		useStateBodyBytes, err = io.ReadAll(userStatResp.Body)
		if err != nil {
			t.Fatal(err)
		}
		err = userStatResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		userStat2 := &user.Stat{}
		err = json.Unmarshal(useStateBodyBytes, userStat2)
		if err != nil {
			t.Fatal(err)
		}
		if userStat.Name == userStat2.Name {
			t.Fatal("username should not be same after migration with different username")
		}
		if userStat.Reference != userStat2.Reference {
			t.Fatal("completely different user stats")
		}
	})

	t.Run("signup-login-migrate-with-pod-and-file-upload-download", func(t *testing.T) {
		c := http.Client{Timeout: time.Duration(1) * time.Minute}

		userRequest := &common.UserRequest{
			UserName: randStringRunes(16),
			Password: randStringRunes(8),
		}

		userBytes, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}

		// create user v1
		signupRequestDataHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserSignup)), bytes.NewBuffer(userBytes))
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
			return
		}

		userLoginHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.UserLogin)), bytes.NewBuffer(userBytes))
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

		podReq := &common.PodRequest{
			PodName:  randStringRunes(16),
			Password: userRequest.Password,
		}
		podReqData, err := json.Marshal(podReq)
		if err != nil {
			t.Fatal(err)
		}

		podHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.PodNew)), bytes.NewBuffer(podReqData))
		if err != nil {
			t.Fatal(err)
		}
		podHttpReq.Header.Add("Content-Type", "application/json")
		podHttpReq.Header.Add("Content-Length", strconv.Itoa(len(podReqData)))
		podHttpReq.Header.Set("Cookie", cookie[0])
		podNewResp, err := c.Do(podHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		err = podNewResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if podNewResp.StatusCode != http.StatusCreated {
			t.Fatal("pod new failed")
		}

		uploadBuf := new(bytes.Buffer)
		fileName := fmt.Sprintf("file_%d", time.Now().Unix())
		uploadWriter := multipart.NewWriter(uploadBuf)
		dataBytes := []byte(fmt.Sprintf("Latest updates %d", time.Now().Unix()))
		err = uploadWriter.WriteField("pod_name", podReq.PodName)
		if err != nil {
			t.Fatal("pod new failed")
		}
		err = uploadWriter.WriteField("dir_path", "/")
		if err != nil {
			t.Fatal("pod new failed")
		}
		err = uploadWriter.WriteField("block_size", "1Mb")
		if err != nil {
			t.Fatal("pod new failed")
		}
		err = uploadWriter.WriteField("content_length", fmt.Sprintf("%d", len(dataBytes)))
		if err != nil {
			t.Fatal("pod new failed")
		}
		uploadPart, err := uploadWriter.CreateFormFile("files", fileName)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(uploadPart, bytes.NewReader(dataBytes))
		if err != nil {
			t.Fatal(err)
		}
		err = uploadWriter.Close()
		if err != nil {
			t.Fatal(err)
		}
		contentType := fmt.Sprintf("multipart/form-data;boundary=%v", uploadWriter.Boundary())
		uploadHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.FileUpload)), uploadBuf)
		if err != nil {
			t.Fatal(err)
		}
		uploadHttpReq.Header.Set("Content-Type", contentType)
		if cookie != nil {
			uploadHttpReq.Header.Set("Cookie", cookie[0])
		}
		uploadResp, err := c.Do(uploadHttpReq)
		if err != nil {
			t.Fatal(err)
		}
		err = uploadResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if uploadResp.StatusCode != http.StatusOK {
			t.Fatal("upload failed")
		}

		userMigrateHttpReq, err := http.NewRequest(http.MethodPost, "http://localhost:9090/v2/user/migrate", bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userMigrateHttpReq.Header.Add("Content-Type", "application/json")
		userMigrateHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userMigrateHttpReq.Header.Set("Cookie", cookie[0])
		useMigrateResp, err := c.Do(userMigrateHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = useMigrateResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if useMigrateResp.StatusCode != http.StatusOK {
			t.Fatal("user should be migrated")
		}

		userLoginHttpReq, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev2, string(common.UserLogin)), bytes.NewBuffer(userBytes))
		if err != nil {
			t.Fatal(err)

		}
		userLoginHttpReq.Header.Add("Content-Type", "application/json")
		userLoginHttpReq.Header.Add("Content-Length", strconv.Itoa(len(userBytes)))
		userLoginResp, err = c.Do(userLoginHttpReq)
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
		cookie = userLoginResp.Header["Set-Cookie"]
		podOpenHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.PodOpen)), bytes.NewBuffer(podReqData))
		if err != nil {
			t.Fatal(err)

		}
		podOpenHttpReq.Header.Add("Content-Type", "application/json")
		podOpenHttpReq.Header.Add("Content-Length", strconv.Itoa(len(podReqData)))
		podOpenHttpReq.Header.Set("Cookie", cookie[0])
		podOpenResp, err := c.Do(podOpenHttpReq)
		if err != nil {
			t.Fatal(err)

		}
		err = podOpenResp.Body.Close()
		if err != nil {
			t.Fatal(err)

		}
		if podOpenResp.StatusCode != http.StatusOK {
			t.Fatal("pod open failed")

		}

		downloadBuf := new(bytes.Buffer)
		downloadWriter := multipart.NewWriter(downloadBuf)
		err = downloadWriter.WriteField("pod_name", podReq.PodName)
		if err != nil {
			t.Fatal("pod new failed")
		}
		err = downloadWriter.WriteField("file_path", fmt.Sprintf("/%s", fileName))
		if err != nil {
			t.Fatal("pod new failed")
		}

		err = downloadWriter.Close()
		if err != nil {
			t.Fatal(err)

		}
		contentType = fmt.Sprintf("multipart/form-data;boundary=%v", downloadWriter.Boundary())
		downloadHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.FileDownload)), downloadBuf)
		if err != nil {
			t.Fatal(err)

		}
		downloadHttpReq.Header.Set("Content-Type", contentType)
		if cookie != nil {
			downloadHttpReq.Header.Set("Cookie", cookie[0])
		}
		downloadResp, err := c.Do(downloadHttpReq)
		if err != nil {
			t.Fatal(err)

		}
		err = downloadResp.Body.Close()
		if err != nil {
			t.Fatal(err)

		}
		if downloadResp.StatusCode != http.StatusOK {
			t.Fatal("download failed")
		}
	})
}
