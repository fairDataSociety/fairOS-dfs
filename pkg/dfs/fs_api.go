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
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *DfsAPI) Mkdir(directoryName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	// make dir
	err := ui.GetPod().MakeDir(ui.GetPodName(), directoryName)
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
	_, _, err = directory.GetDirNode(podDir, ui.GetFeed(), ui.GetAccount().GetUserAccountInfo())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *DfsAPI) RmDir(directoryName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	err := ui.GetPod().RemoveDir(ui.GetPodName(), directoryName)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) ListDir(currentDir, sessionId string) ([]dir.DirOrFileEntry, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	entries, err := ui.GetPod().ListEntiesInDir(ui.GetPodName(), currentDir)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (d *DfsAPI) DirectoryStat(directoryName, sessionId string, printNames bool) (*dir.DirStats, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	ds, err := ui.GetPod().DirectoryStat(ui.GetPodName(), directoryName, printNames)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (d *DfsAPI) ChangeDirectory(directoryName, sessionId string) (*pod.Info, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().ChangeDir(ui.GetPodName(), directoryName)
	if err != nil {
		return nil, err
	}
	return podInfo, nil
}

//
// File related API's
//
func (d *DfsAPI) CopyToLocal(localDir, podFile, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	err := ui.GetPod().CopyToLocal(ui.GetPodName(), localDir, podFile)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) Cat(fileName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	err := ui.GetPod().Cat(ui.GetPodName(), fileName)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) DeleteFile(podFile, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	err := ui.GetPod().RemoveFile(ui.GetPodName(), podFile)
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) FileStat(fileName, sessionId string) (*file.FileStats, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	ds, err := ui.GetPod().FileStat(ui.GetPodName(), fileName)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (d *DfsAPI) UploadFile(fileName, sessionId string, fileSize int64, fd io.Reader, podDir, blockSize, compression string) (string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return "", ErrPodNotOpen
	}

	ref, err := ui.GetPod().UploadFile(ui.GetPodName(), fileName, fileSize, fd, podDir, blockSize, compression)
	if err != nil {
		return "", err
	}
	return ref, nil
}

func (d *DfsAPI) DownloadFile(podFile, sessionId string) (io.ReadCloser, string, string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, "", "", ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, "", "", ErrPodNotOpen
	}

	reader, ref, size, err := ui.GetPod().DownloadFile(ui.GetPodName(), podFile)
	if err != nil {
		return nil, "", "", err
	}
	return reader, ref, size, nil
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
