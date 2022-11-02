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

package pod_test

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func TestOpen(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	pod1 := pod.NewPod(mockClient, fd, acc, tm, logger)
	podName1 := "test1"
	podName2 := "test2"

	t.Run("open-pod", func(t *testing.T) {
		// open non-existent the pod
		_, err := pod1.OpenPod(podName1, "password")
		if !errors.Is(err, pod.ErrInvalidPodName) {
			t.Fatal("pod should not be present")
		}

		// create a pod
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		info, err := pod1.CreatePod(podName1, "password", "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod1, podName1)

		// open the pod
		podInfo, err := pod1.OpenPod(podName1, "password")
		if err != nil {
			t.Fatal(err)
		}

		// validate if properly opened
		if podInfo == nil {
			t.Fatalf("pod not opened")
		}
		gotPodInfo, err := pod1.GetPodInfoFromPodMap(podName1)
		if err != nil {
			t.Fatalf("pod not opened")
		}
		if gotPodInfo == nil {
			t.Fatalf("pod not opened")
		}
		if gotPodInfo.GetPodName() != podName1 {
			t.Fatalf("invalid pod name")
		}
	})

	t.Run("open-pod-async", func(t *testing.T) {
		// open non-existent the pod
		_, err := pod1.OpenPod(podName2, "password")
		if !errors.Is(err, pod.ErrInvalidPodName) {
			t.Fatal("pod should not be present")
		}

		// create a pod
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		info, err := pod1.CreatePod(podName2, "password", "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod1, podName2)

		// open the pod
		podInfo, err := pod1.OpenPodAsync(context.Background(), podName2, "password")
		if err != nil {
			t.Fatal(err)
		}

		// validate if properly opened
		if podInfo == nil {
			t.Fatalf("pod not opened")
		}
		gotPodInfo, err := pod1.GetPodInfoFromPodMap(podName2)
		if err != nil {
			t.Fatalf("pod not opened")
		}
		if gotPodInfo == nil {
			t.Fatalf("pod not opened")
		}
		if gotPodInfo.GetPodName() != podName2 {
			t.Fatalf("invalid pod name")
		}
	})
}

func uploadFile(t *testing.T, fileObject *file.File, filePath, fileName, compression string, fileSize int64, blockSize uint32) ([]byte, error) {
	// create a temp file
	fd, err := os.CreateTemp("", fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fd.Name())

	// write contents to file
	content := make([]byte, fileSize)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
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

func addFilesAndDirectories(t *testing.T, info *pod.Info, pod1 *pod.Pod, podName1 string) {
	t.Helper()
	dirObject := info.GetDirectory()
	err := dirObject.MkDir("/parentDir")
	if err != nil {
		t.Fatal(err)
	}

	node := dirObject.GetDirFromDirectoryMap("/parentDir")
	if pod1.GetName(node) != "parentDir" {
		t.Fatal("dir name mismatch in pod")
	}
	if pod1.GetPath(node) != "/" {
		t.Fatal("dir path mismatch in pod")
	}

	// populate the directory with few directory and files
	err = dirObject.MkDir("/parentDir/subDir1")
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.MkDir("/parentDir/subDir2")
	if err != nil {
		t.Fatal(err)
	}
	fileObject := info.GetFile()
	_, err = uploadFile(t, fileObject, "/parentDir", "file1", "", 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.AddEntryToDir("/parentDir", "file1", true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = uploadFile(t, fileObject, "/parentDir", "file2", "", 200, 20)
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.AddEntryToDir("/parentDir", "file2", true)
	if err != nil {
		t.Fatal(err)
	}

	// close the pod
	err = pod1.ClosePod(podName1)
	if err != nil {
		t.Fatal(err)
	}

	// close the pod
	err = pod1.ClosePod(podName1)
	if !errors.Is(err, pod.ErrPodNotOpened) {
		t.Fatal("pod should not be open")
	}
}
