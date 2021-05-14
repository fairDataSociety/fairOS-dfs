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
	"path/filepath"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Mkdir is a controller function which validates if the user is logged in,
// pod is open and calls the make directory function in the dir object.
func (d *DfsAPI) Mkdir(dirToCreateWithPath, sessionId string) error {
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
	err = directory.MkDir(dirToCreateWithPath)
	if err != nil {
		return err
	}
	return nil
}

// IsDirPresent is acontroller function which validates if the user is logged in,
// pod is open and calls the dir object to check if the directory is present.
func (d *DfsAPI) IsDirPresent(directoryNameWithPath, sessionId string) (bool, error) {
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

	dirPresent := directory.IsDirectoryPresent(directoryNameWithPath)
	return dirPresent, nil
}

// RmDir is a controller function which validates if the user is logged in,
// pod is open and calls the dir object to remove the supplied directory.
func (d *DfsAPI) RmDir(directoryNameWithPath, sessionId string) error {
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
	err = directory.RmDir(directoryNameWithPath)
	if err != nil {
		return err
	}
	return nil
}

// ListDir is a controller function which validates if the user is logged in,
// pod is open and calls the dir object to list the contents of the supplied directory.
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
	dEntries, fileList, err := directory.ListDir(currentDir)
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

// DirectoryStat is a controller function which validates if the user is logged in,
// pod is open and calls the dir object to get the information about the given directory.
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

// DeleteFile is a controller function which validates if the user is logged in,
// pod is open and delete the file. It also remove the file entry from the parent
// directory.
func (d *DfsAPI) DeleteFile(podFileWithPath, sessionId string) error {
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
	err = file.RmFile(podFileWithPath)
	if err != nil {
		return err
	}

	// update the directory by removing the file from it
	fileDir := filepath.Dir(podFileWithPath)
	fileName := filepath.Base(podFileWithPath)
	return directory.RemoveEntryFromDir(fileDir, fileName, true)
}

// FileStat is a controller function which validates if the user is logged in,
// pod is open and gets the information about the file.
func (d *DfsAPI) FileStat(podFileWithPath, sessionId string) (*f.Stats, error) {
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
	ds, err := file.GetStats(ui.GetPodName(), podFileWithPath)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// UploadFile is a controller function which validates if the user is logged in,
//  pod is open and calls the upload function.
func (d *DfsAPI) UploadFile(podFileName, sessionId string, fileSize int64, fd io.Reader, podPath, compression string, blockSize uint32) error {
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
	err = file.Upload(fd, podFileName, fileSize, blockSize, podPath, compression)
	if err != nil {
		return err
	}

	// add the file to the directory metadata
	return directory.AddEntryToDir(podPath, podFileName, true)
}

// DownloadFile is a controller function which validates if the user is logged in,
// pod is open and calls the download function.
func (d *DfsAPI) DownloadFile(podFileWithPath, sessionId string) (io.ReadCloser, uint64, error) {
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

	// download the file by creating the reader
	file := podInfo.GetFile()
	reader, size, err := file.Download(podFileWithPath)
	if err != nil {
		return nil, 0, err
	}
	return reader, size, nil
}

// ShareFile is a controller function which validates if the user is logged in,
// pod is open and calls the sharefile function.
func (d *DfsAPI) ShareFile(podFileWithPath, destinationUser, sessionId string) (string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return "", ErrPodNotOpen
	}

	// get podInfo and construct the path
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return "", err
	}

	sharingRef, err := d.users.ShareFileWithUser(podFileWithPath, destinationUser, ui, ui.GetPod(), podInfo.GetAccountInfo().GetAddress())
	if err != nil {
		return "", err
	}
	return sharingRef, nil
}

// ReceiveFile is a controller function which validates if the user is logged in,
// pod is open and calls the ReceiveFile function to get the shared file in to the
// given pod.
func (d *DfsAPI) ReceiveFile(sessionId string, sharingRef utils.SharingReference, dir string) (string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return "", ErrPodNotOpen
	}

	return d.users.ReceiveFileFromUser(ui.GetPodName(), sharingRef, ui, ui.GetPod(), dir)
}

// ReceiveInfo is a controller function which validates if the user is logged in,
// pod is open and calls the ReceiveInfo function to display the shared files
// information.
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

	return d.users.ReceiveFileInfo(sharingRef)
}
