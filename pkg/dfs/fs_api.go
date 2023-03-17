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
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Mkdir is a controller function which validates if the user is logged-in,
// pod is open and calls the make directory function in the dir object.
func (a *API) Mkdir(podName, dirToCreateWithPath, sessionId string, mode uint32) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// get the dir object and make directory
	podInfo, podPassword, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.MkDir(dirToCreateWithPath, podPassword, mode)
}

// RenameDir is a controller function which validates if the user is logged-in,
// pod is open and calls the rename directory function in the dir object.
func (a *API) RenameDir(podName, dirToRenameWithPath, newName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// get the dir object and rename directory
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.RenameDir(dirToRenameWithPath, newName, podInfo.GetPodPassword())
}

// IsDirPresent is a controller function which validates if the user is logged-in,
// pod is open and calls the dir object to check if the directory is present.
func (a *API) IsDirPresent(podName, directoryNameWithPath, sessionId string) (bool, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false, ErrUserNotLoggedIn
	}

	// get pod Info
	podInfo, podPassword, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return false, err
	}
	directory := podInfo.GetDirectory()
	directoryNameWithPath = filepath.ToSlash(directoryNameWithPath)
	dirPresent := directory.IsDirectoryPresent(directoryNameWithPath, podPassword)
	return dirPresent, nil
}

// RmDir is a controller function which validates if the user is logged-in,
// pod is open and calls the dir object to remove the supplied directory.
func (a *API) RmDir(podName, directoryNameWithPath, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// get the dir object and remove directory
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	return directory.RmDir(directoryNameWithPath, podInfo.GetPodPassword())
}

// ListDir is a controller function which validates if the user is logged-in,
// pod is open and calls the dir object to list the contents of the supplied directory.
func (a *API) ListDir(podName, currentDir, sessionId string) ([]dir.Entry, []f.Entry, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// get the dir object and list directory
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, nil, err
	}
	directory := podInfo.GetDirectory()

	// check if directory present
	totalPath := utils.CombinePathAndFile(currentDir, "")
	if directory.GetInode(podInfo.GetPodPassword(), totalPath) == nil {
		return nil, nil, dir.ErrDirectoryNotPresent
	}
	dEntries, fileList, err := directory.ListDir(currentDir, podInfo.GetPodPassword())
	if err != nil {
		return nil, nil, err
	}
	file := podInfo.GetFile()
	fEntries, err := file.ListFiles(fileList, podInfo.GetPodPassword())
	if err != nil {
		return nil, nil, err
	}
	return dEntries, fEntries, nil
}

// DirectoryStat is a controller function which validates if the user is logged-in,
// pod is open and calls the dir object to get the information about the given directory.
func (a *API) DirectoryStat(podName, directoryPath, sessionId string) (*dir.Stats, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// get the dir object and stat directory
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}
	directory := podInfo.GetDirectory()
	directoryPath = filepath.ToSlash(directoryPath)
	if directoryPath != utils.PathSeparator {
		parent := filepath.ToSlash(filepath.Dir(directoryPath))
		inode := directory.GetInode(podInfo.GetPodPassword(), parent)
		if inode == nil {
			return nil, dir.ErrDirectoryNotPresent
		}
		found := false
		directoryName := filepath.Base(directoryPath)
		for _, name := range inode.FileOrDirNames {
			if strings.TrimPrefix(name, "_D_") == directoryName {
				found = true
				break
			}
		}
		if !found {
			return nil, dir.ErrDirectoryNotPresent
		}
	}

	return directory.DirStat(podName, podInfo.GetPodPassword(), directoryPath)
}

// DirectoryInode is a controller function which validates if the user is logged-in,
// pod is open and calls the dir object to get the inode info about the given directory.
func (a *API) DirectoryInode(podName, directoryName, sessionId string) (*dir.Inode, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// get the dir object and stat directory
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}
	directory := podInfo.GetDirectory()
	inode := directory.GetInode(podInfo.GetPodPassword(), directoryName)
	if inode == nil {
		a.logger.Errorf("dir not found: %s", directoryName)
		return nil, fmt.Errorf("dir not found")
	}
	return inode, nil
}

// ChmodDir is a controller function which validates if the user is logged-in,
// pod is open and calls changes mode of a directory
func (a *API) ChmodDir(podName, directoryNameWithPath, sessionId string, mode uint32) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	// get status of the file
	directory := podInfo.GetDirectory()
	return directory.Chmod(directoryNameWithPath, podInfo.GetPodPassword(), mode)
}

// DeleteFile is a controller function which validates if the user is logged-in,
// pod is open and delete the file. It also removes the file entry from the parent
// directory.
func (a *API) DeleteFile(podName, podFileWithPath, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	// check if the pod is readonly before deleting a file
	if podInfo.GetAccountInfo().IsReadOnlyPod() {
		return errReadOnlyPod
	}
	directory := podInfo.GetDirectory()

	file := podInfo.GetFile()
	err = file.RmFile(podFileWithPath, podInfo.GetPodPassword())
	if err != nil {
		if err == f.ErrDeletedFeed {
			return pod.ErrInvalidFile
		}
		return err
	}

	// update the directory by removing the file from it
	fileDir := filepath.Dir(podFileWithPath)
	fileName := filepath.Base(podFileWithPath)
	return directory.RemoveEntryFromDir(fileDir, podInfo.GetPodPassword(), fileName, true)
}

// FileStat is a controller function which validates if the user is logged-in,
// pod is open and gets the information about the file.
func (a *API) FileStat(podName, podFileWithPath, sessionId string) (*f.Stats, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}
	podFileWithPath = filepath.ToSlash(podFileWithPath)
	directory := podInfo.GetDirectory()
	inode := directory.GetInode(podInfo.GetPodPassword(), filepath.ToSlash(filepath.Dir(podFileWithPath)))
	if inode != nil {
		found := false
		fileName := filepath.Base(podFileWithPath)
		for _, name := range inode.FileOrDirNames {
			if strings.TrimPrefix(name, "_F_") == fileName {
				found = true
				break
			}
		}
		if !found {
			return nil, f.ErrFileNotFound
		}
	}

	file := podInfo.GetFile()
	return file.GetStats(podName, podFileWithPath, podInfo.GetPodPassword())
}

// UploadFile is a controller function which validates if the user is logged-in,
//
//	pod is open and calls the upload function.
func (a *API) UploadFile(podName, podFileName, sessionId string, fileSize int64, fd io.Reader, podPath, compression string, blockSize, mode uint32, overwrite bool) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	directory := podInfo.GetDirectory()
	podPath = filepath.ToSlash(podPath)
	// check if file exists, then backup the file
	totalPath := utils.CombinePathAndFile(podPath, podFileName)
	alreadyPresent := file.IsFileAlreadyPresent(podInfo.GetPodPassword(), totalPath)
	if alreadyPresent {
		if !overwrite {
			m, err := file.BackupFromFileName(totalPath, podInfo.GetPodPassword())
			if err != nil {
				return err
			}
			err = directory.AddEntryToDir(podPath, podInfo.GetPodPassword(), m.Name, true)
			if err != nil {
				return err
			}
		}
		err = directory.RemoveEntryFromDir(podPath, podInfo.GetPodPassword(), podFileName, true)
		if err != nil {
			return err
		}
	}
	err = file.Upload(fd, podFileName, fileSize, blockSize, mode, podPath, compression, podInfo.GetPodPassword())
	if err != nil {
		return err
	}

	// add the file to the directory metadata
	return directory.AddEntryToDir(podPath, podInfo.GetPodPassword(), podFileName, true)
}

// RenameFile is a controller function which validates if the user is logged-in,
//
//	pod is open and calls renaming of a file
func (a *API) RenameFile(podName, fileNameWithPath, newFileNameWithPath, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	directory := podInfo.GetDirectory()

	fileNameWithPath = filepath.ToSlash(fileNameWithPath)
	newFileNameWithPath = filepath.ToSlash(newFileNameWithPath)

	// check if file exists
	if !file.IsFileAlreadyPresent(podInfo.GetPodPassword(), fileNameWithPath) {
		return ErrFileNotPresent
	}
	if file.IsFileAlreadyPresent(podInfo.GetPodPassword(), newFileNameWithPath) {
		return ErrFileAlreadyPresent
	}

	m, err := file.RenameFromFileName(fileNameWithPath, newFileNameWithPath, podInfo.GetPodPassword())
	if err != nil {
		return err
	}
	oldPrnt := filepath.ToSlash(filepath.Dir(fileNameWithPath))
	newPrnt := filepath.ToSlash(filepath.Dir(newFileNameWithPath))

	// add the file to the directory metadata
	err = directory.AddEntryToDir(newPrnt, podInfo.GetPodPassword(), m.Name, true)
	if err != nil {
		return err
	}

	return directory.RemoveEntryFromDir(oldPrnt, podInfo.GetPodPassword(), filepath.Base(fileNameWithPath), true)
}

// DownloadFile is a controller function which validates if the user is logged-in,
// pod is open and calls the download function.
func (a *API) DownloadFile(podName, podFileWithPath, sessionId string) (io.ReadCloser, uint64, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, 0, ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, 0, err
	}

	// download the file by creating the reader
	file := podInfo.GetFile()
	reader, size, err := file.Download(podFileWithPath, podInfo.GetPodPassword())
	if err != nil {
		return nil, 0, err
	}
	return reader, size, nil
}

// WriteAtFile is a controller function which writes a file from a given offset
//
//	pod is open and calls writeAt of a file
func (a *API) WriteAtFile(podName, fileNameWithPath, sessionId string, update io.Reader, offset uint64, truncate bool) (int, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return 0, err
	}
	file := podInfo.GetFile()
	fileNameWithPath = filepath.ToSlash(fileNameWithPath)
	// check if file exists
	if !file.IsFileAlreadyPresent(podInfo.GetPodPassword(), fileNameWithPath) {
		return 0, ErrFileNotPresent
	}

	return file.WriteAt(fileNameWithPath, podInfo.GetPodPassword(), update, offset, truncate)
}

// ReadSeekCloser is a controller function which validates if the user is logged-in,
// pod is open and calls the download function.
func (a *API) ReadSeekCloser(podName, podFileWithPath, sessionId string) (io.ReadSeekCloser, uint64, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, 0, ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, 0, err
	}

	// download the file by creating the reader
	file := podInfo.GetFile()
	reader, size, err := file.ReadSeeker(podFileWithPath, podInfo.GetPodPassword())
	if err != nil {
		return nil, 0, err
	}
	return reader, size, nil
}

// ShareFile is a controller function which validates if the user is logged-in,
// pod is open and calls the shareFile function.
func (a *API) ShareFile(podName, podFileWithPath, destinationUser, sessionId string) (string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return "", err
	}

	sharingRef, err := a.users.ShareFileWithUser(podName, podInfo.GetPodPassword(), podFileWithPath, destinationUser, ui, ui.GetPod(), podInfo.GetAccountInfo().GetAddress())
	if err != nil {
		return "", err
	}
	return sharingRef, nil
}

// ReceiveFile is a controller function which validates if the user is logged-in,
// pod is open and calls the ReceiveFile function to get the shared file in to the
// given pod.
func (a *API) ReceiveFile(podName, sessionId, ref string, dir string) (string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	return a.users.ReceiveFileFromUser(podName, ref, ui, ui.GetPod(), dir)
}

// ReceiveInfo is a controller function which validates if the user is logged-in,
// calls the ReceiveInfo function to display the shared file's information.
func (a *API) ReceiveInfo(sessionId, ref string) (*user.ReceiveFileInfo, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return a.users.ReceiveFileInfo(ref)
}

// StatusFile is a controller function which validates if the user is logged-in,
// pod is open and calls the sync status of file function.
func (a *API) StatusFile(podName, podFileWithPath, sessionId string) (int64, int64, int64, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, 0, 0, ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return 0, 0, 0, err
	}

	// get status of the file
	file := podInfo.GetFile()
	return file.Status(podFileWithPath, podInfo.GetPodPassword())
}

// ChmodFile is a controller function which validates if the user is logged-in,
// pod is open and calls changes mode of a file
func (a *API) ChmodFile(podName, podFileWithPath, sessionId string, mode uint32) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// get podInfo and construct the path
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	// get status of the file
	file := podInfo.GetFile()
	return file.Chmod(podFileWithPath, podInfo.GetPodPassword(), mode)
}
