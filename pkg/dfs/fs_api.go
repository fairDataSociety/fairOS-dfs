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

package dfs

import (
	"fmt"
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *DfsAPI) Mkdir(path, directoryName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	// get the dir object and make directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	err = directory.MkDir(ui.GetPodName(), path, directoryName)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) IsDirPresent(directoryName, sessionId string) (bool, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return false, ErrPodNotOpen
	}

	// get pod Info
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return false, err
	}
	directory := podInfo.GetDirectory()
	podDir := podInfo.GetCurrentPodPathAndName() + directoryName
	_, _, err = directory.GetDirNode(podDir, podInfo.GetFeed(), podInfo.GetUserAddress())
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DfsAPI) RmDir(path, directoryName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	// get the dir object and remove directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	err = directory.RmDir(ui.GetPodName(), path, directoryName)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) ListDir(currentDir, sessionId string) ([]dir.Entry, []f.Entry, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, nil, ErrPodNotOpen
	}

	// get the dir object and list directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, nil, err
	}
	directory := podInfo.GetDirectory()
	dEntries, fileList, err := directory.ListDir(ui.GetPodName(), currentDir)
	if err != nil {
		return nil, nil, err
	}
	file := podInfo.GetFile()
	fEntries, err := file.ListFiles(fileList)
	if err != nil {
		return nil, nil, err
	}
	return dEntries, fEntries, nil
}

func (d *DfsAPI) DirectoryStat(directoryName, sessionId string) (*dir.DirStats, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	// get the dir object and stat directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}
	directory := podInfo.GetDirectory()
	ds, err := directory.DirStat(ui.GetPodName(), directoryName)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

//
// File related API's
//
func (d *DfsAPI) DeleteFile(path, podFile, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	// check if the pod is readonly before deleting a file
	if podInfo.GetAccountInfo().IsReadOnlyPod() {
		return ErrReadOnlyPod
	}
	directory := podInfo.GetDirectory()

	file := podInfo.GetFile()
	err = file.RmFile(ui.GetPodName(), path, podFile)
	if err != nil {
		return err
	}

	// update the directory by removing the file from it
	return directory.RemoveFileFromDirectory(path, podFile)
}

func (d *DfsAPI) FileStat(fileName, sessionId string) (*f.Stats, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}
	file := podInfo.GetFile()
	ds, err := file.GetStats(ui.GetPodName(), fileName)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (d *DfsAPI) UploadFile(fileName, sessionId string, fileSize int64, fd io.Reader, podDir, compression string, blockSize uint32) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	directory := podInfo.GetDirectory()
	err = file.Upload(fd, fileName, fileSize, blockSize, podDir, compression)
	if err != nil {
		return err
	}

	// add the file to the directory metadata
	return directory.AddFileToDirectory(podDir, fileName)
}

func (d *DfsAPI) DownloadFile(podFile, sessionId string) (io.ReadCloser, uint64, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, 0, ErrPodNotOpen
	}

	// check if logged in to pod
	if !ui.GetPod().IsPodOpened(ui.GetPodName()) {
		return nil, 0, fmt.Errorf("login to pod to do this operation")
	}

	// get podInfo and construct the path
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, 0, err
	}
	var path string
	if podInfo.IsCurrentDirRoot() {
		path = podInfo.GetCurrentPodPathAndName() + podFile
	} else {
		path = podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + podFile
	}

	// check if file already present
	if !podInfo.GetFile().IsFileAlreadyPresent(path) {
		return nil, 0, fmt.Errorf("file not present in pod")
	}

	// download the file by creating the reader
	file := podInfo.GetFile()
	reader, size, err := file.Download(podFile)
	if err != nil {
		return nil, 0, err
	}
	return reader, size, nil
}

func (d *DfsAPI) ShareFile(podFile, destinationUser, sessionId string) (string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return "", ErrPodNotOpen
	}

	sharingRef, err := d.users.ShareFileWithUser(ui.GetPodName(), podFile, destinationUser, ui, ui.GetPod())
	if err != nil {
		return "", err
	}
	return sharingRef, nil
}

func (d *DfsAPI) ReceiveFile(sessionId string, sharingRef utils.SharingReference, dir string) (string, string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return "", "", ErrPodNotOpen
	}

	return d.users.ReceiveFileFromUser(ui.GetPodName(), sharingRef, ui, ui.GetPod(), dir)
}

func (d *DfsAPI) ReceiveInfo(sessionId string, sharingRef utils.SharingReference) (*user.ReceiveFileInfo, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	return d.users.ReceiveFileInfo(ui.GetPodName(), sharingRef, ui, ui.GetPod())
}
