package dir_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestRenameDirectory(t *testing.T) {
	mockClient := bm.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("rename-dir-same-prnt", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, fileObject, tm, logger)
		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}
		err := dirObject.MkDir("/", podPassword)
		if !errors.Is(err, dir.ErrInvalidDirectoryName) {
			t.Fatal("invalid dir name")
		}
		longDirName, err := utils.GetRandString(101)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/"+longDirName, podPassword)
		if !errors.Is(err, dir.ErrTooLongDirectoryName) {
			t.Fatal("dir name too long")
		}

		// create some dir and files
		err = dirObject.MkDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir", podPassword)
		if !errors.Is(err, dir.ErrDirectoryAlreadyPresent) {
			t.Fatal("dir already present")
		}
		// populate the directory with few directory and files
		err = dirObject.MkDir("/parentDir/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		r := new(bytes.Buffer)
		err = fileObject.Upload(r, "file1", 0, 100, "/parentDir", "", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = fileObject.Upload(r, "file2", 0, 100, "/parentDir", "", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = fileObject.Upload(r, "file2", 0, 100, "/parentDir/subDir2", "", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file enty as file listing is not tested here
		err = dirObject.AddEntryToDir("/parentDir", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/parentDir", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/parentDir/subDir2", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}
		// rename
		err = dirObject.RenameDir("/parentDir", "/parentNew", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirEntries, _, err := dirObject.ListDir("/", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if dirEntries[0].Name != "parentNew" {
			t.Fatal("rename failed for parentDir")
		}

		err = dirObject.MkDir("/parent", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.RenameDir("/parentNew", "/parent", podPassword)
		if !errors.Is(err, dir.ErrDirectoryAlreadyPresent) {
			t.Fatal("directory name should already be present")
		}

		// validate dir listing
		dirEntries, files, err := dirObject.ListDir("/parentNew", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirs := []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 2 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 2 {
			t.Fatalf("invalid files entry count")
		}

		sort.Strings(dirs)
		sort.Strings(files)
		// validate entry names
		if dirs[0] != "subDir1" {
			t.Fatalf("invalid directory name")
		}
		if dirs[1] != "subDir2" {
			t.Fatalf("invalid directory name")
		}
		if files[0] != "/parentNew/file1" {
			t.Fatalf("invalid file name")
		}
		if files[1] != "/parentNew/file2" {
			t.Fatalf("invalid file name")
		}

		_, files, err = dirObject.ListDir("/parentNew/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Fatal("file count mismatch /parentNew/subDir2")
		}
		if files[0] != "/parentNew/subDir2/file2" {
			t.Fatal("file name mismatch /parentNew/subDir2")
		}

		_, n, err := fileObject.Download("/parentNew/subDir2/file2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatal("file size mismatch")
		}
	})

	t.Run("rename-dir-diff-prnt", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, fileObject, tm, logger)
		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}
		err := dirObject.MkDir("/", podPassword)
		if !errors.Is(err, dir.ErrInvalidDirectoryName) {
			t.Fatal("invalid dir name")
		}
		longDirName, err := utils.GetRandString(101)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/"+longDirName, podPassword)
		if !errors.Is(err, dir.ErrTooLongDirectoryName) {
			t.Fatal("dir name too long")
		}

		// create some dir and files
		err = dirObject.MkDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err = dirObject.MkDir("/parentDir/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir/subDir1/subDir11", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir/subDir1/subDir11/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		r := new(bytes.Buffer)
		err = fileObject.Upload(r, "file1", 0, 100, "/parentDir/subDir1/subDir11/sub111", "", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// just add dummy file enty as file listing is not tested here
		err = dirObject.AddEntryToDir("/parentDir/subDir1/subDir11/sub111", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}

		// rename
		err = dirObject.RenameDir("/parentDir/subDir1/subDir11/sub111", "/parentDir/subDir2/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = dirObject.ListDir("/parentDir/subDir1/subDir11/sub111", podPassword)
		if err == nil {
			t.Fatal("should fail")
		}

		dirEntries, files, err := dirObject.ListDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirs := []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 2 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		sort.Strings(dirs)
		sort.Strings(files)

		if dirs[0] != "subDir1" && dirs[1] != "subDir2" {
			t.Fatal("wrong list of directories")
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		dirs = []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 1 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		if dirs[0] != "subDir11" {
			t.Fatal("wrong list of directories")
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirs = []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 1 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		if dirs[0] != "sub111" {
			t.Fatal("wrong list of directories")
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir/subDir2/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if len(dirEntries) != 0 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 1 {
			t.Fatalf("invalid files entry count")
		}

		if files[0] != "/parentDir/subDir2/sub111/file1" {
			t.Fatal("wrong list of files")
		}

		err = dirObject.RenameDir("/parentDir/subDir2/sub111", "/parentDir/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirs = []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 3 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		sort.Strings(dirs)
		sort.Strings(files)
		if dirs[0] != "sub111" && dirs[1] != "subDir1" && dirs[2] != "subDir2" {
			t.Fatal("wrong list of directories")
		}

		// validate dir listing
		dirEntries, files, err = dirObject.ListDir("/parentDir/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		dirs = []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 1 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		if dirs[0] != "subDir11" {
			t.Fatal("wrong list of directories")
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if len(dirEntries) != 0 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 0 {
			t.Fatalf("invalid files entry count")
		}

		_, _, err = dirObject.ListDir("/parentDir/subDir2/sub111", podPassword)
		if err == nil {
			t.Fatal("should be err")
		}

		dirEntries, files, err = dirObject.ListDir("/parentDir/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		dirs = []string{}

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 0 {
			t.Fatalf("invalid directory entry count")
		}
		if len(files) != 1 {
			t.Fatalf("invalid files entry count")
		}

		if files[0] != "/parentDir/sub111/file1" {
			t.Fatal("wrong list of files")
		}
		err = dirObject.RenameDir("/parentDir/sub111", "/parentDir/subDir2/sub111", podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})
}
