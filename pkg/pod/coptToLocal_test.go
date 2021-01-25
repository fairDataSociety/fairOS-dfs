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
	"strings"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestPod_CopyToLocal(t *testing.T) {
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
	t.Run("copy-file-from-root", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		podFile := createRandomFileInPod(t, 540, pod1, podName1, info.GetCurrentPodPathAndName())

		err = pod1.CopyToLocal(podName1, podFile, os.TempDir())
		if err != nil {
			t.Fatalf("error copying file to local dir %s", err.Error())
		}

		fileInfo, err := os.Stat(os.TempDir() + utils.PathSeperator + filepath.Base(podFile))
		if err != nil {
			t.Fatalf("file not copied to local")
		}

		if fileInfo.Size() != 540 {
			t.Fatalf("invalid file size")
		}

		os.Remove(fileInfo.Name())
		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("copy-file-from-firstdir", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		err = pod1.MakeDir(podName1, firstDir)
		if err != nil {
			t.Fatalf("error creating directory %s", firstDir)
		}
		dirPath := utils.PathSeperator + podName1 + utils.PathSeperator + firstDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}
		podFile := createRandomFileInPod(t, 540, pod1, podName1, dirPath)

		err = pod1.CopyToLocal(podName1, podFile, os.TempDir())
		if err != nil {
			t.Fatalf("error copying file to local dir %s", err.Error())
		}

		fileInfo, err := os.Stat(os.TempDir() + utils.PathSeperator + filepath.Base(podFile))
		if err != nil {
			t.Fatalf("file not copied to local")
		}

		if fileInfo.Size() != 540 {
			t.Fatalf("invalid file size")
		}

		os.Remove(fileInfo.Name())
		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("copy-file-to-dot", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		podFile := createRandomFileInPod(t, 540, pod1, podName1, info.GetCurrentPodPathAndName())
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = pod1.CopyToLocal(podName1, podFile, pwd)
		if err != nil {
			t.Fatalf("error copying file to local dir %s", err.Error())
		}

		fileInfo, err := os.Stat(pwd + utils.PathSeperator + filepath.Base(podFile))
		if err != nil {
			t.Fatalf("file not copied to local")
		}

		if fileInfo.Size() != 540 {
			t.Fatalf("invalid file size")
		}

		err = os.Remove(fileInfo.Name())
		if err != nil {
			t.Fatal(err)
		}
		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

}

func createRandomFileInPod(t *testing.T, size int, pod1 *Pod, podName, podDir string) string {
	file, err := ioutil.TempFile("/tmp", "fairOS-dfs")
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

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	fn := file.Name()
	podDir = strings.TrimPrefix(podDir, utils.PathSeperator+podName)
	if podDir == "" {
		podDir = "."
	}

	fd, err := os.Open(fn)
	if err != nil {
		t.Fatal(err)
	}
	fName := filepath.Base(file.Name())
	_, err = pod1.UploadFile(podName, fName, int64(size), fd, podDir, "100", "false")
	if err != nil {
		t.Fatalf("createRandomFileInPod failed: %s", err.Error())
	}
	err = fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if podDir == "." {
		podDir = ""
	}
	fileName := podDir + utils.PathSeperator + filepath.Base(file.Name())
	fileName = strings.TrimPrefix(fileName, utils.PathSeperator+podName)
	return fileName
}
