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
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Mkdir is a controller function which validates if the user is logged in,
// pod is open and calls the make directory function in the dir object.
func (a *API) Mkdir(podName, dirToCreateWithPath, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// get the dir object and make directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.MkDir(dirToCreateWithPath)
}

// RenameDir is a controller function which validates if the user is logged in,
// pod is open and calls the rename directory function in the dir object.
func (a *API) RenameDir(podName, dirToRenameWithPath, newName, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// get the dir object and rename directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.RenameDir(dirToRenameWithPath, newName)
}

// IsDirPresent is acontroller function which validates if the user is logged in,
// pod is open and calls the dir object to check if the directory is present.
func (a *API) IsDirPresent(podName, directoryNameWithPath, sessionId string) (bool, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return false, ErrPodNotOpen
	}

	// get pod Info
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return false, err
	}
	directory := podInfo.GetDirectory()
	directoryNameWithPath = filepath.ToSlash(directoryNameWithPath)
	dirPresent := directory.IsDirectoryPresent(directoryNameWithPath)
	return dirPresent, nil
}

// RmDir is a controller function which validates if the user is logged in,
// pod is open and calls the dir object to remove the supplied directory.
func (a *API) RmDir(podName, directoryNameWithPath, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// get the dir object and remove directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.RmDir(directoryNameWithPath)
}

// ListDir is a controller function which validates if the user is logged in,
// pod is open and calls the dir object to list the contents of the supplied directory.
func (a *API) ListDir(podName, currentDir, sessionId string) ([]dir.Entry, []f.Entry, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, nil, ErrPodNotOpen
	}

	// get the dir object and list directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, nil, err
	}
	directory := podInfo.GetDirectory()

	// check if directory present
	totalPath := utils.CombinePathAndFile(currentDir, "")
	if directory.GetDirFromDirectoryMap(totalPath) == nil {
		return nil, nil, dir.ErrDirectoryNotPresent
	}
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
func (a *API) DirectoryStat(podName, directoryName, sessionId string) (*dir.Stats, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	// get the dir object and stat directory
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}
	directory := podInfo.GetDirectory()
	ds, err := directory.DirStat(podName, directoryName)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// DeleteFile is a controller function which validates if the user is logged in,
// pod is open and delete the file. It also remove the file entry from the parent
// directory.
func (a *API) DeleteFile(podName, podFileWithPath, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	// check if the pod is readonly before deleting a file
	if podInfo.GetAccountInfo().IsReadOnlyPod() {
		return errReadOnlyPod
	}
	directory := podInfo.GetDirectory()

	file := podInfo.GetFile()
	err = file.RmFile(podFileWithPath)
	if err != nil {
		if err == f.ErrDeletedFeed {
			return pod.ErrInvalidFile
		}
		return err
	}

	// update the directory by removing the file from it
	fileDir := filepath.Dir(podFileWithPath)
	fileName := filepath.Base(podFileWithPath)
	return directory.RemoveEntryFromDir(fileDir, fileName, true)
}

// FileStat is a controller function which validates if the user is logged in,
// pod is open and gets the information about the file.
func (a *API) FileStat(podName, podFileWithPath, sessionId string) (*f.Stats, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}
	file := podInfo.GetFile()
	ds, err := file.GetStats(podName, podFileWithPath)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// UploadFile is a controller function which validates if the user is logged in,
//  pod is open and calls the upload function.
func (a *API) UploadFile(podName, podFileName, sessionId string, fileSize int64, fd io.Reader, podPath, compression string, blockSize uint32) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	directory := podInfo.GetDirectory()
	podPath = filepath.ToSlash(podPath)

	// check if file exists, then backup the file
	totalPath := utils.CombinePathAndFile(podPath, podFileName)
	if file.IsFileAlreadyPresent(totalPath) {
		m, err := file.BackupFromFileName(totalPath)
		if err != nil {
			return err
		}
		err = directory.AddEntryToDir(podPath, m.Name, true)
		if err != nil {
			return err
		}
		err = directory.RemoveEntryFromDir(podPath, podFileName, true)
		if err != nil {
			return err
		}
	}

	err = file.Upload(fd, podFileName, fileSize, blockSize, podPath, compression)
	if err != nil {
		return err
	}

	// add the file to the directory metadata
	return directory.AddEntryToDir(podPath, podFileName, true)
}

// RenameFile is a controller function which validates if the user is logged in,
//  pod is open and calls renaming of a file
func (a *API) RenameFile(podName, fileNameWithPath, newFileNameWithPath, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	directory := podInfo.GetDirectory()

	fileNameWithPath = filepath.ToSlash(fileNameWithPath)
	newFileNameWithPath = filepath.ToSlash(newFileNameWithPath)

	// check if file exists
	if !file.IsFileAlreadyPresent(fileNameWithPath) {
		return ErrFileNotPresent
	}
	if file.IsFileAlreadyPresent(newFileNameWithPath) {
		return ErrFileAlreadyPresent
	}

	m, err := file.RenameFromFileName(fileNameWithPath, newFileNameWithPath)
	if err != nil {
		return err
	}

	oldPrnt := filepath.ToSlash(filepath.Dir(fileNameWithPath))
	newPrnt := filepath.ToSlash(filepath.Dir(newFileNameWithPath))

	// add the file to the directory metadata
	err = directory.AddEntryToDir(newPrnt, m.Name, true)
	if err != nil {
		return err
	}

	return directory.RemoveEntryFromDir(oldPrnt, filepath.Base(fileNameWithPath), true)
}

// DownloadFile is a controller function which validates if the user is logged in,
// pod is open and calls the download function.
func (a *API) DownloadFile(podName, podFileWithPath, sessionId string) (io.ReadCloser, uint64, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, 0, ErrPodNotOpen
	}

	// check if logged in to pod
	if !ui.GetPod().IsPodOpened(podName) {
		return nil, 0, fmt.Errorf("login to pod to do this operation")
	}

	// get podInfo and construct the path
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
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

// ReadSeekCloser is a controller function which validates if the user is logged in,
// pod is open and calls the download function.
func (a *API) ReadSeekCloser(podName, podFileWithPath, sessionId string) (io.ReadSeekCloser, uint64, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, 0, ErrPodNotOpen
	}

	// check if logged in to pod
	if !ui.GetPod().IsPodOpened(podName) {
		return nil, 0, fmt.Errorf("login to pod to do this operation")
	}

	// get podInfo and construct the path
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, 0, err
	}

	// download the file by creating the reader
	file := podInfo.GetFile()
	reader, size, err := file.ReadSeeker(podFileWithPath)
	if err != nil {
		return nil, 0, err
	}
	return reader, size, nil
}

// ShareFile is a controller function which validates if the user is logged in,
// pod is open and calls the sharefile function.
func (a *API) ShareFile(podName, podFileWithPath, destinationUser, sessionId string) (string, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return "", ErrPodNotOpen
	}

	// get podInfo and construct the path
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return "", err
	}

	sharingRef, err := a.users.ShareFileWithUser(podName, podFileWithPath, destinationUser, ui, ui.GetPod(), podInfo.GetAccountInfo().GetAddress())
	if err != nil {
		return "", err
	}
	return sharingRef, nil
}

// ReceiveFile is a controller function which validates if the user is logged in,
// pod is open and calls the ReceiveFile function to get the shared file in to the
// given pod.
func (a *API) ReceiveFile(podName, sessionId string, sharingRef utils.SharingReference, dir string) (string, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return "", ErrPodNotOpen
	}

	return a.users.ReceiveFileFromUser(podName, sharingRef, ui, ui.GetPod(), dir)
}

// ReceiveInfo is a controller function which validates if the user is logged in,
// pod is open and calls the ReceiveInfo function to display the shared files
// information.
func (a *API) ReceiveInfo(podName, sessionId string, sharingRef utils.SharingReference) (*user.ReceiveFileInfo, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	return a.users.ReceiveFileInfo(sharingRef)
}
