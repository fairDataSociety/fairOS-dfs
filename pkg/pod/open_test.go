/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestPod_LoginPod(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	pod1 := NewPod(mockClient, fd, acc, logger)

	podName1 := "test1"
	firstDir := "dir1"
	t.Run("simple-login-to-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		err = pod1.ClosePod(podName1)
		if err != nil {
			t.Fatalf("could not logout")
		}

		infoLogin, err := pod1.OpenPod(podName1, "password")
		if err != nil {
			t.Fatalf("login failed")
		}
		if info.podName != infoLogin.podName {
			t.Fatalf("invalid podname")
		}
		if info.GetCurrentPodPathAndName() != infoLogin.GetCurrentPodPathAndName() {
			t.Fatalf("invalid podname path and name")
		}

		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("login-with-sync-contents", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		//Make a dir
		err = pod1.MakeDir(podName1, firstDir)
		if err != nil {
			t.Fatalf("error creating directory %s", firstDir)
		}

		dirPath := utils.PathSeperator + podName1 + utils.PathSeperator + firstDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		// create a file
		localFile, clean := createRandomFile(t, 540)
		defer clean()
		podDir := utils.PathSeperator + firstDir
		fileName := filepath.Base(localFile)
		fd, err := os.Open(localFile)
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()
		_, err = pod1.UploadFile(podName1, fileName, 540, fd, podDir, "100", "false")
		if err != nil {
			t.Fatalf("upload failed: %s", err.Error())
		}
		if !info.getFile().IsFileAlreadyPResent(dirPath + utils.PathSeperator + filepath.Base(localFile)) {
			t.Fatalf("file not copied in pod")
		}

		err = pod1.ClosePod(podName1)
		if err != nil {
			t.Fatalf("could not logout")
		}

		// Now login and check if the dir and file exists
		infoLogin, err := pod1.OpenPod(podName1, "password")
		if err != nil {
			t.Fatalf("login failed")
		}
		if info.podName != infoLogin.podName {
			t.Fatalf("invalid podname")
		}
		if info.GetCurrentPodPathAndName() != infoLogin.GetCurrentPodPathAndName() {
			t.Fatalf("invalid podname path and name")
		}
		dirInodeLogin := infoLogin.dir.GetDirFromDirectoryMap(dirPath)
		if dirInodeLogin == nil {
			t.Fatalf("dir not synced")
		}
		if dirInodeLogin.Meta.Path != info.GetCurrentPodPathAndName() {
			t.Fatalf("dir not synced")
		}
		if dirInodeLogin.Meta.Name != firstDir {
			t.Fatalf("dir not synced")
		}
		fileMeta := infoLogin.getFile().GetFromFileMap(dirPath + utils.PathSeperator + filepath.Base(localFile))
		if fileMeta == nil {
			t.Fatalf("file not synced")
		}

		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

}

func createRandomFile(t *testing.T, size int) (string, func()) {
	file, err := ioutil.TempFile("/tmp", "FairOS")
	if err != nil {
		t.Fatal(err)
	}
	bytes := make([]byte, size)
	_, err = rand.Read(bytes)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		t.Fatal(err)
	}
	clean := func() { os.Remove(file.Name()) }
	fileName := file.Name()
	return fileName, clean
}
