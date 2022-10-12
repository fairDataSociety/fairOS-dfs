package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
	dataDir, err := os.MkdirTemp("", "new")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dataDir)
	users := user.NewUsers(dataDir, mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger, dataDir)
	handler = api.NewMockHandler(dfsApi, logger, []string{})
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

	t.Run("signup-login-pod-dir-file-rename", func(t *testing.T) {
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

		// pod new
		podRequest := &common.PodRequest{
			PodName:  randStringRunes(16),
			Password: userRequest.Password,
		}
		podBytes, err := json.Marshal(podRequest)
		if err != nil {
			t.Fatal(err)
		}
		podNewHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.PodNew)), bytes.NewBuffer(podBytes))
		if err != nil {
			t.Fatal(err)

		}
		podNewHttpReq.Header.Set("Cookie", cookie[0])
		podNewHttpReq.Header.Add("Content-Type", "application/json")
		podNewHttpReq.Header.Add("Content-Length", strconv.Itoa(len(podBytes)))
		podNewResp, err := c.Do(podNewHttpReq)
		if err != nil {
			t.Fatal(err)
		}

		err = podNewResp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if podNewResp.StatusCode != 201 {
			t.Fatal("pod creation failed")
		}
		entries := []struct {
			path    string
			isDir   bool
			size    int64
			content []byte
		}{
			{
				path:  "/dir1",
				isDir: true,
			},
			{
				path:  "/dir2",
				isDir: true,
			},
			{
				path:  "/dir3",
				isDir: true,
			},
			{
				path: "/file1",
				size: 1024 * 1024,
			},
			{
				path: "/dir1/file11",
				size: 1024 * 512,
			},
			{
				path: "/dir1/file12",
				size: 1024 * 1024,
			},
			{
				path: "/dir3/file31",
				size: 1024 * 1024,
			},
			{
				path: "/dir3/file32",
				size: 1024 * 1024,
			},
			{
				path: "/dir3/file33",
				size: 1024,
			},
			{
				path:  "/dir2/dir4",
				isDir: true,
			},
			{
				path:  "/dir2/dir4/dir5",
				isDir: true,
			},
			{
				path: "/dir2/dir4/file241",
				size: 5 * 1024 * 1024,
			},
			{
				path: "/dir2/dir4/dir5/file2451",
				size: 10 * 1024 * 1024,
			},
		}

		for _, v := range entries {
			if v.isDir {
				mkdirRqst := common.FileSystemRequest{
					PodName:       podRequest.PodName,
					DirectoryPath: v.path,
				}
				mkDirBytes, err := json.Marshal(mkdirRqst)
				if err != nil {
					t.Fatal(err)
				}
				mkDirHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.DirMkdir)), bytes.NewBuffer(mkDirBytes))
				if err != nil {
					t.Fatal(err)

				}
				mkDirHttpReq.Header.Set("Cookie", cookie[0])
				mkDirHttpReq.Header.Add("Content-Type", "application/json")
				mkDirHttpReq.Header.Add("Content-Length", strconv.Itoa(len(mkDirBytes)))
				mkDirResp, err := c.Do(mkDirHttpReq)
				if err != nil {
					t.Fatal(err)
				}
				err = mkDirResp.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if mkDirResp.StatusCode != 201 {
					t.Fatal("mkdir failed")
				}
			} else {
				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)
				contentLength := fmt.Sprintf("%d", v.size)

				err = writer.WriteField("pod_name", podRequest.PodName)
				if err != nil {
					t.Fatal(err)
				}
				err = writer.WriteField("content_length", contentLength)
				if err != nil {
					t.Fatal(err)
				}
				err = writer.WriteField("dir_path", filepath.Dir(v.path))
				if err != nil {
					t.Fatal(err)
				}
				err = writer.WriteField("block_size", "4kb")
				if err != nil {
					t.Fatal(err)
				}
				part, err := writer.CreateFormFile("files", filepath.Base(v.path))
				if err != nil {
					t.Fatal(err)
				}
				reader := &io.LimitedReader{R: rand.Reader, N: v.size}
				_, err = io.Copy(part, reader)
				if err != nil {
					t.Fatal(err)
				}
				err = writer.Close()
				if err != nil {
					t.Fatal(err)
				}

				uploadReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(common.FileUpload)), body)
				if err != nil {
					t.Fatal(err)

				}
				uploadReq.Header.Set("Cookie", cookie[0])
				contentType := fmt.Sprintf("multipart/form-data;boundary=%v", writer.Boundary())
				uploadReq.Header.Add("Content-Type", contentType)
				uploadResp, err := c.Do(uploadReq)
				if err != nil {
					t.Fatal(err)
				}
				err = uploadResp.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if uploadResp.StatusCode != 200 {
					t.Fatal("upload failed")
				}
			}
		}

		for _, v := range entries {
			if v.isDir {
				statReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?pod_name=%s&dir_path=%s", basev1, string(common.DirStat), podRequest.PodName, v.path), http.NoBody)
				if err != nil {
					t.Fatal(err)

				}
				statReq.Header.Set("Cookie", cookie[0])
				statResp, err := c.Do(statReq)
				if err != nil {
					t.Fatal(err)
				}
				err = statResp.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if statResp.StatusCode != 200 {
					t.Fatal("dir stat failed")
				}
			} else {
				if v.isDir {
					statReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?pod_name=%s&dir_path=%s", basev1, string(common.FileStat), podRequest.PodName, v.path), http.NoBody)
					if err != nil {
						t.Fatal(err)

					}
					statReq.Header.Set("Cookie", cookie[0])
					statResp, err := c.Do(statReq)
					if err != nil {
						t.Fatal(err)
					}
					err = statResp.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
					if statResp.StatusCode != 200 {
						t.Fatal("file stat failed")
					}
				}
			}
		}
		// rename  file "/dir2/dir4/dir5/file2451" => "/dir2/dir4/dir5/file24511"
		renames := []struct {
			oldPath string
			newPath string
			isDir   bool
		}{
			{
				oldPath: "/dir2/dir4/dir5/file2451",
				newPath: "/dir2/dir4/dir5/file24511",
				isDir:   false,
			},
			{
				oldPath: "/dir2/dir4/dir5/file24511",
				newPath: "/file24511",
				isDir:   false,
			},
			{
				oldPath: "/dir2",
				newPath: "/dir2020",
				isDir:   true,
			},
			{
				oldPath: "/dir2020/dir4",
				newPath: "/dir2020/dir4040",
				isDir:   true,
			}, {
				oldPath: "/dir3/file33",
				newPath: "/dir2020/file33",
				isDir:   false,
			}, {
				oldPath: "/dir1/file12",
				newPath: "/dir2020/dir4040/file12",
				isDir:   false,
			},
		}
		for _, v := range renames {
			renameReq := common.RenameRequest{
				PodName: podRequest.PodName,
				OldPath: v.oldPath,
				NewPath: v.newPath,
			}

			renameBytes, err := json.Marshal(renameReq)
			if err != nil {
				t.Fatal(err)
			}
			url := common.FileRename
			if v.isDir {
				url = common.DirRename
			}
			renameHttpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", basev1, string(url)), bytes.NewBuffer(renameBytes))
			if err != nil {
				t.Fatal(err)

			}
			renameHttpReq.Header.Set("Cookie", cookie[0])
			renameHttpReq.Header.Add("Content-Type", "application/json")
			renameHttpReq.Header.Add("Content-Length", strconv.Itoa(len(renameBytes)))
			renameResp, err := c.Do(renameHttpReq)
			if err != nil {
				t.Fatal(err)
			}
			err = renameResp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if renameResp.StatusCode != 200 {
				t.Fatal("rename failed", url)
			}
		}

		newEntries := []struct {
			path    string
			isDir   bool
			size    int64
			content []byte
		}{
			{
				path:  "/dir1",
				isDir: true,
			},
			{
				path:  "/dir2020",
				isDir: true,
			},
			{
				path:  "/dir3",
				isDir: true,
			},
			{
				path: "/file1",
				size: 1024 * 1024,
			},
			{
				path: "/dir1/file11",
				size: 1024 * 512,
			},
			{
				path: "/dir2020/dir4040/file12",
				size: 1024 * 1024,
			},
			{
				path: "/dir3/file31",
				size: 1024 * 1024,
			},
			{
				path: "/dir3/file32",
				size: 1024 * 1024,
			},
			{
				path: "/dir2020/file33",
				size: 1024,
			},
			{
				path:  "/dir2020/dir4040",
				isDir: true,
			},
			{
				path:  "/dir2020/dir4040/dir5",
				isDir: true,
			},
			{
				path: "/dir2020/dir4040/file241",
				size: 5 * 1024 * 1024,
			},
			{
				path: "/file24511",
				size: 10 * 1024 * 1024,
			},
		}
		for _, v := range newEntries {
			if v.isDir {
				statReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?pod_name=%s&dir_path=%s", basev1, string(common.DirStat), podRequest.PodName, v.path), http.NoBody)
				if err != nil {
					t.Fatal(err)

				}
				statReq.Header.Set("Cookie", cookie[0])
				statResp, err := c.Do(statReq)
				if err != nil {
					t.Fatal(err)
				}
				err = statResp.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if statResp.StatusCode != 200 {
					t.Fatal("dir stat failed")
				}
			} else {
				if v.isDir {
					statReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?pod_name=%s&dir_path=%s", basev1, string(common.FileStat), podRequest.PodName, v.path), http.NoBody)
					if err != nil {
						t.Fatal(err)

					}
					statReq.Header.Set("Cookie", cookie[0])
					statResp, err := c.Do(statReq)
					if err != nil {
						t.Fatal(err)
					}
					err = statResp.Body.Close()
					if err != nil {
						t.Fatal(err)
					}
					if statResp.StatusCode != 200 {
						t.Fatal("file stat failed")
					}
				}
			}
		}
	})
}
