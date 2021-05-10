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

package user_test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestSharing(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)

	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = acc1.CreatePodAccount(1, "password", false)
	if err != nil {
		t.Fatal(err)
	}
	fd1 := feed.New(acc1.GetUserAccountInfo(), mockClient, logger)
	pod1 := pod.NewPod(mockClient, fd1, acc1, logger)
	podName1 := "test1"

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = acc2.CreatePodAccount(1, "password", false)
	if err != nil {
		t.Fatal(err)
	}
	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, logger)
	podName2 := "test2"

	t.Run("sharing-user", func(t *testing.T) {
		dataDir1, err := ioutil.TempDir("", "sharing")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir1)

		//create source user
		userObject1 := user.NewUsers(dataDir1, mockClient, "", logger)
		_, _, ui, err := userObject1.CreateNewUser("user1", "password1", "", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		// create source pod
		info1, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		ui.SetPodName(podName1)

		// create dir and file
		dirObject1 := info1.GetDirectory()
		err = dirObject1.MkDir("/", "parentDir1")
		if err != nil {
			t.Fatal(err)
		}
		fileObject1 := info1.GetFile()
		_, err = uploadFile(t, fileObject1, "/parentDir1", "file1", "", 100, 10)
		if err != nil {
			t.Fatal(err)
		}

		// share file with another user
		sharingRefString, err := userObject1.ShareFileWithUser(podName1, "/parentDir1/file1", "user2", ui, pod1, info1.GetPodAddress())
		if err != nil {
			t.Fatal(err)
		}

		// create destination user
		dataDir2, err := ioutil.TempDir("", "sharing")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir2)

		//create destination user
		userObject2 := user.NewUsers(dataDir2, mockClient, "", logger)
		_, _, ui, err = userObject2.CreateNewUser("user2", "password2", "", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		// create destination pod
		info2, err := pod2.CreatePod(podName2, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		// create dir and file
		dirObject2 := info2.GetDirectory()
		err = dirObject2.MkDir("/", "parentDir2")
		if err != nil {
			t.Fatal(err)
		}

		// receive file info
		sharingRef, err := utils.ParseSharingReference(sharingRefString)
		if err != nil {
			t.Fatal(err)
		}
		receiveFileInfo, err := userObject2.ReceiveFileInfo(sharingRef)
		if err != nil {
			t.Fatal(err)
		}

		// validate receive file info
		if receiveFileInfo == nil {
			t.Fatalf("invalid receive file info")
		}
		if receiveFileInfo.FileName != "file1" {
			t.Fatalf("invalid filename received")
		}
		if receiveFileInfo.PodName != podName1 {
			t.Fatalf("invalid podName received")
		}
		if receiveFileInfo.Size != strconv.FormatUint(100, 10) {
			t.Fatalf("invalid file size received")
		}
		if receiveFileInfo.BlockSize != strconv.FormatUint(10, 10) {
			t.Fatalf("invalid block size received")
		}

		// receive file
		destinationFilePath, err := userObject2.ReceiveFileFromUser(podName2, sharingRef, ui, pod2, "/parentDir2")
		if err != nil {
			t.Fatal(err)
		}

		// varify receive
		if destinationFilePath != "/parentDir2/file1" {
			t.Fatalf("invalid destination file name")
		}
		_, files, err := dirObject2.ListDir("/parentDir2")
		if err != nil {
			t.Fatal(err)
		}
		if files == nil {
			t.Fatalf("file not imported")
		}
		if len(files) != 1 {
			t.Fatalf("file not imported")
		}
		if files[0] != "/parentDir2/file1" {
			t.Fatalf("file not imported")
		}

	})
}

func uploadFile(t *testing.T, fileObject *file.File, filePath, fileName, compression string, fileSize int64, blockSize uint32) ([]byte, error) {
	// create a temp file
	fd, err := ioutil.TempFile("", fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fd.Name())

	// write contents to file
	content := make([]byte, fileSize)
	rand.Read(content)
	if _, err = fd.Write(content); err != nil {
		t.Fatal(err)
	}

	// close file
	uploadFileName := fd.Name()
	err = fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	// open file to upload
	f1, err := os.Open(uploadFileName)
	if err != nil {
		t.Fatal(err)
	}

	// upload  the temp file
	return content, fileObject.Upload(f1, fileName, fileSize, blockSize, filePath, compression)
}
