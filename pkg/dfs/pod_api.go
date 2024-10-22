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
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/datahub"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// CreatePod creates a new pod
func (a *API) CreatePod(podName, sessionId string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// open the pod
	pi, err := a.prepareOwnPod(ui, podName)
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(podName, pi)
	return pi, nil
}

// DeletePod deletes a pod
func (a *API) DeletePod(podName, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	// delete all the directory, files, and database tables under this pod from
	// the Swarm network.
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()
	// check if this is a shared pod
	if podInfo.GetFeed().IsReadOnlyFeed() {
		// delete the pod and close if it is opened
		err = ui.GetPod().DeleteSharedPod(podName)
		if err != nil {
			return err
		}

		// remove from the login session
		ui.RemovePodName(podName)
		return nil
	}
	err = directory.RmRootDir(podInfo.GetPodPassword())
	if err != nil && !errors.Is(err, file.ErrFileNotFound) {
		return err
	}
	// delete the pod and close if it is opened
	err = ui.GetPod().DeleteOwnPod(podName)
	if err != nil {
		return err
	}
	ui.RemovePodName(podName)
	return nil
}

// OpenPod opens a pod
func (a *API) OpenPod(podName, sessionId string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(podName, podInfo)
	return podInfo, nil
}

// ClosePod closes a pod
func (a *API) ClosePod(podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// close the pod
	err := ui.GetPod().ClosePod(podName)
	if err != nil {
		return err
	}

	// delete podName in the login user session
	ui.RemovePodName(podName)
	return nil
}

// CommitPodFeeds commits feed for a pod on swarm
func (a *API) CommitPodFeeds(podName, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return ui.GetPod().CommitFeeds(podName)
}

// PodStat returns the pod stat
func (a *API) PodStat(podName, sessionId string) (*pod.Stat, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// get the pod stat
	podStat, err := ui.GetPod().PodStat(podName)
	if err != nil {
		return nil, err
	}
	return podStat, nil
}

// SyncPod syncs a pod
func (a *API) SyncPod(podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// sync the pod
	err := ui.GetPod().SyncPod(podName)
	if err != nil {
		return err
	}
	return nil
}

// SyncPodAsync syncs a pod asynchronously
func (a *API) SyncPodAsync(ctx context.Context, podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// sync the pod
	err := ui.GetPod().SyncPodAsync(ctx, podName)
	if err != nil {
		return err
	}
	return nil
}

// ListPods lists all the pods of a user
func (a *API) ListPods(sessionId string) ([]string, []string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// list pods of a user
	pods, sharedPods, err := ui.GetPod().ListPods()
	if err != nil {
		return nil, nil, err
	}
	return pods, sharedPods, nil
}

// PodList lists all available pods in json format
func (a *API) PodList(sessionId string) (*pod.List, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// list pods of a user
	return ui.GetPod().PodList()
}

// PodShare shares a pod
func (a *API) PodShare(podName, sharedPodName, sessionId string) (string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// get the pod stat
	address, err := ui.GetPod().PodShare(podName, sharedPodName)
	if err != nil {
		return "", err
	}
	return address, nil
}

// PodReceiveInfo receives the pod information
func (a *API) PodReceiveInfo(sessionId string, ref utils.Reference) (*pod.ShareInfo, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePodInfo(ref)
}

// PublicPodReceiveInfo receives the pod information for a public pod
func (a *API) PublicPodReceiveInfo(ref utils.Reference) (*pod.ShareInfo, error) {
	data, resp, err := a.client.DownloadBlob(ref.Bytes())
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if resp != http.StatusOK { // skipcq: TCV-001
		return nil, fmt.Errorf("ReceivePodInfo: could not download blob")
	}

	var shareInfo *pod.ShareInfo
	err = json.Unmarshal(data, &shareInfo)
	if err != nil {
		return nil, err
	}

	return shareInfo, nil
}

// PublicPodFileDownload downloads a file from a public pod
func (a *API) PublicPodFileDownload(pod *pod.ShareInfo, filePath string) (io.ReadCloser, uint64, error) {

	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.feedCacheSize, a.feedCacheTTL, a.logger)
	topic := utils.HashString(filePath)
	_, metaBytes, err := fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password), false)
	if err != nil {
		return nil, 0, err
	}

	if string(metaBytes) == utils.DeletedFeedMagicWord {
		a.logger.Errorf("found deleted feed for %s\n", filePath)
		return nil, 0, file.ErrDeletedFeed
	}

	var meta *file.MetaData
	err = json.Unmarshal(metaBytes, &meta)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	var fileInode file.INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	reader := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
	return reader, meta.Size, nil
}

// PublicPodFileDownloadFromMetadata downloads a file from a public pod
func (a *API) PublicPodFileDownloadFromMetadata(meta *file.MetaData) (io.ReadCloser, uint64, error) {

	fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	var fileInode file.INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	reader := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
	return reader, meta.Size, nil
}

// PublicPodKVEntryGet gets a kv entry from a public pod
func (a *API) PublicPodKVEntryGet(pod *pod.ShareInfo, name, key string) ([]string, []byte, error) {

	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.feedCacheSize, a.feedCacheTTL, a.logger)
	kvStore := c.NewKeyValueStore(pod.PodName, fd, accountInfo, address, a.client, a.logger)

	err := kvStore.OpenKVTable(name, pod.Password)
	if err != nil {
		return nil, nil, err
	}

	return kvStore.KVGet(name, key)
}

// PublicPodKVGetter gets a kv store getter interface
func (a *API) PublicPodKVGetter(pod *pod.ShareInfo) KVGetter {
	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.feedCacheSize, a.feedCacheTTL, a.logger)
	return c.NewKeyValueStore(pod.PodName, fd, accountInfo, address, a.client, a.logger)
}

// PublicPodDisLs lists a directory from a public pod
func (a *API) PublicPodDisLs(pod *pod.ShareInfo, dirPathToLs string) ([]dir.Entry, []file.Entry, error) {
	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.feedCacheSize, a.feedCacheTTL, a.logger)

	dirNameWithPath := filepath.ToSlash(dirPathToLs)
	var (
		inode dir.Inode
		data  []byte
	)

	topic := utils.HashString(utils.CombinePathAndFile(dirNameWithPath, dir.IndexFileName))
	_, metaBytes, err := fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password), false)
	if err != nil { // skipcq: TCV-001
		topic = utils.HashString(dirNameWithPath)
		_, data, err = fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password), false)
		if err != nil {
			return nil, nil, fmt.Errorf("list dir : %v", err) // skipcq: TCV-001
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}
	} else {
		if string(metaBytes) == utils.DeletedFeedMagicWord {
			a.logger.Errorf("found deleted feed for %s\n", dirNameWithPath)
			return nil, nil, file.ErrDeletedFeed
		}

		var meta *file.MetaData
		err = json.Unmarshal(metaBytes, &meta)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}
		fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}

		var fileInode file.INode
		err = json.Unmarshal(fileInodeBytes, &fileInode)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}
		r := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
		data, err = io.ReadAll(r)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, nil, err
		}
	}

	var wg sync.WaitGroup
	dirChan := make(chan dir.Entry, len(inode.FileOrDirNames))
	fileChan := make(chan file.Entry, len(inode.FileOrDirNames))
	errChan := make(chan error, len(inode.FileOrDirNames))
	semaphore := make(chan struct{}, 4)
	missingCount := 0
	for _, fileOrDirName := range inode.FileOrDirNames {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore slot

		go func(fileOrDirName string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore slot
			if strings.HasPrefix(fileOrDirName, "_D_") {

				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				dirPath := utils.CombinePathAndFile(dirNameWithPath, dirName)
				var (
					inode dir.Inode
					data  []byte
				)

				dirTopic := utils.HashString(utils.CombinePathAndFile(dirPath, dir.IndexFileName))
				_, indexBytes, err := fd.GetFeedData(dirTopic, accountInfo.GetAddress(), []byte(pod.Password), false)
				if err != nil { // skipcq: TCV-001
					topic = utils.HashString(dirName)
					_, data, err = fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password), false)
					if err != nil {
						errChan <- fmt.Errorf("list dir : %v", err)
						return
					}
					err = inode.Unmarshal(data)
					if err != nil {
						errChan <- err
						return
					}
				} else {
					if string(indexBytes) == utils.DeletedFeedMagicWord {
						a.logger.Errorf("found deleted feed for %s\n", dirNameWithPath)
						errChan <- file.ErrDeletedFeed
						return
					}

					var meta *file.MetaData
					err = json.Unmarshal(indexBytes, &meta)
					if err != nil {
						errChan <- err
						return
					}
					fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
					if err != nil {
						errChan <- err
						return
					}

					var fileInode file.INode
					err = json.Unmarshal(fileInodeBytes, &fileInode)
					if err != nil {
						errChan <- err
						return
					}
					r := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
					data, err = io.ReadAll(r)
					if err != nil {
						errChan <- err
						return
					}
					err = inode.Unmarshal(data)
					if err != nil {
						errChan <- err
						return
					}
				}
				entry := dir.Entry{
					Name:             inode.Meta.Name,
					ContentType:      dir.MimeTypeDirectory, // per RFC2425
					CreationTime:     strconv.FormatInt(inode.Meta.CreationTime, 10),
					AccessTime:       strconv.FormatInt(inode.Meta.AccessTime, 10),
					ModificationTime: strconv.FormatInt(inode.Meta.ModificationTime, 10),
					Mode:             inode.Meta.Mode,
				}
				dirChan <- entry
			} else if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimPrefix(fileOrDirName, "_F_")
				filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)

				fileTopic := utils.HashString(utils.CombinePathAndFile(filePath, ""))

				_, data, err := fd.GetFeedData(fileTopic, accountInfo.GetAddress(), []byte(pod.Password), false)
				if err != nil { // skipcq: TCV-001
					errChan <- fmt.Errorf("file mtdt : %s : %v", filePath, err)
					return
				}
				if string(data) == utils.DeletedFeedMagicWord { // skipcq: TCV-001
					return
				}
				var meta *file.MetaData
				err = json.Unmarshal(data, &meta)
				if err != nil { // skipcq: TCV-001
					errChan <- fmt.Errorf("file mtdt : %v", err)
					return
				}
				entry := file.Entry{
					Name:             meta.Name,
					ContentType:      meta.ContentType,
					Size:             strconv.FormatUint(meta.Size, 10),
					BlockSize:        strconv.FormatInt(int64(meta.BlockSize), 10),
					CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
					AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
					ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
					Mode:             meta.Mode,
				}

				fileChan <- entry
			}
		}(fileOrDirName)
	}

	// Close channels in a goroutine after all goroutines are done
	go func() {
		wg.Wait()
		close(dirChan)
		close(fileChan)
		close(errChan)
	}()

	var listEntries []dir.Entry
	var fileEntries []file.Entry

	for {
		select {
		case dirEntry, ok := <-dirChan:
			if ok {
				listEntries = append(listEntries, dirEntry)
			}
		case fileEntry, ok := <-fileChan:
			if ok {
				fileEntries = append(fileEntries, fileEntry)
			}
		case err := <-errChan:
			if err != nil {
				return nil, nil, err
			}
		case <-time.After(time.Hour):
			return nil, nil, fmt.Errorf("timeout while listing directory")
		}
		if len(listEntries)+len(fileEntries)+missingCount == len(inode.FileOrDirNames) {
			break
		}
	}

	return listEntries, fileEntries, nil
}

// PublicPodSnapshot Gets the current snapshot from a public pod
func (a *API) PublicPodSnapshot(p *pod.ShareInfo, dirPathToLs string) (*pod.DirSnapShot, error) {
	accountInfo := &account.Info{}
	address := utils.HexToAddress(p.Address)
	accountInfo.SetAddress(address)
	dirSnapShot := &pod.DirSnapShot{
		FileList: make([]file.MetaData, 0),
		DirList:  make([]*pod.DirSnapShot, 0),
	}
	fd := feed.New(accountInfo, a.client, a.feedCacheSize, a.feedCacheTTL, a.logger)

	dirNameWithPath := filepath.ToSlash(dirPathToLs)
	var (
		inode dir.Inode
		data  []byte
	)
	fmt.Println("dirNameWithPath", dirNameWithPath)

	topic := utils.HashString(utils.CombinePathAndFile(dirNameWithPath, dir.IndexFileName))
	_, metaBytes, err := fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(p.Password), false)
	if err != nil { // skipcq: TCV-001
		fmt.Println("err", err)
		topic = utils.HashString(dirNameWithPath)
		_, data, err = fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(p.Password), false)
		if err != nil {
			return nil, fmt.Errorf("list dir : %v for %s", err, dirNameWithPath) // skipcq: TCV-001
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	} else {
		if string(metaBytes) == utils.DeletedFeedMagicWord {
			a.logger.Errorf("found deleted feed for %s\n", dirNameWithPath)
			return nil, file.ErrDeletedFeed
		}

		var meta *file.MetaData
		err = json.Unmarshal(metaBytes, &meta)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		var fileInode file.INode
		err = json.Unmarshal(fileInodeBytes, &fileInode)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		r := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
		data, err = io.ReadAll(r)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	}
	dirSnapShot.Name = inode.Meta.Name
	dirSnapShot.ContentType = dir.MimeTypeDirectory
	dirSnapShot.CreationTime = strconv.FormatInt(inode.Meta.CreationTime, 10)
	dirSnapShot.AccessTime = strconv.FormatInt(inode.Meta.AccessTime, 10)
	dirSnapShot.ModificationTime = strconv.FormatInt(inode.Meta.ModificationTime, 10)
	dirSnapShot.Mode = inode.Meta.Mode
	err = a.getSnapShotForDir(dirSnapShot, fd, accountInfo, inode.FileOrDirNames, dirNameWithPath, p.Password)
	if err != nil {
		return nil, err
	}
	return dirSnapShot, nil
}

func (a *API) getSnapShotForDir(dirL *pod.DirSnapShot, fd *feed.API, accountInfo *account.Info, fileOrDirNames []string, dirNameWithPath, password string) error {
	var wg sync.WaitGroup
	dirChan := make(chan dir.Inode, len(fileOrDirNames))
	fileChan := make(chan file.MetaData, len(fileOrDirNames))
	errChan := make(chan error, len(fileOrDirNames))
	semaphore := make(chan struct{}, 4)
	missingCount := 0
	for _, fileOrDirName := range fileOrDirNames {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore slot

		go func(fileOrDirName string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore slot
			if strings.HasPrefix(fileOrDirName, "_D_") {

				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				dirPath := utils.CombinePathAndFile(dirNameWithPath, dirName)
				var (
					inode dir.Inode
					data  []byte
				)

				dirTopic := utils.HashString(utils.CombinePathAndFile(dirPath, dir.IndexFileName))
				_, indexBytes, err := fd.GetFeedData(dirTopic, accountInfo.GetAddress(), []byte(password), false)
				if err != nil { // skipcq: TCV-001
					topic := utils.HashString(dirName)
					_, data, err = fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(password), false)
					if err != nil {
						errChan <- fmt.Errorf("list dir : %v", err)
						return
					}
					err = inode.Unmarshal(data)
					if err != nil {
						errChan <- err
						return
					}
				} else {
					if string(indexBytes) == utils.DeletedFeedMagicWord {
						errChan <- file.ErrDeletedFeed
						return
					}

					var meta *file.MetaData
					err = json.Unmarshal(indexBytes, &meta)
					if err != nil {
						errChan <- err
						return
					}
					fileInodeBytes, _, err := a.client.DownloadBlob(meta.InodeAddress)
					if err != nil {
						errChan <- err
						return
					}

					var fileInode file.INode
					err = json.Unmarshal(fileInodeBytes, &fileInode)
					if err != nil {
						errChan <- err
						return
					}
					r := file.NewReader(fileInode, a.client, meta.Size, meta.BlockSize, meta.Compression, false)
					data, err = io.ReadAll(r)
					if err != nil {
						errChan <- err
						return
					}
					err = inode.Unmarshal(data)
					if err != nil {
						errChan <- err
						return
					}
				}
				dirChan <- inode
			} else if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimPrefix(fileOrDirName, "_F_")
				filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)

				fileTopic := utils.HashString(utils.CombinePathAndFile(filePath, ""))

				_, data, err := fd.GetFeedData(fileTopic, accountInfo.GetAddress(), []byte(password), false)
				if err != nil { // skipcq: TCV-001
					errChan <- fmt.Errorf("file mtdt : %s : %v", filePath, err)
					return
				}
				if string(data) == utils.DeletedFeedMagicWord { // skipcq: TCV-001
					return
				}
				var meta *file.MetaData
				err = json.Unmarshal(data, &meta)
				if err != nil { // skipcq: TCV-001
					errChan <- fmt.Errorf("file mtdt : %v", err)
					return
				}

				fileChan <- *meta
			}
		}(fileOrDirName)
	}

	// Close channels in a goroutine after all goroutines are done
	go func() {
		wg.Wait()
		close(dirChan)
		close(fileChan)
		close(errChan)
	}()

	for {
		select {
		case inode, ok := <-dirChan:
			if ok {

				dirItem := &pod.DirSnapShot{
					Name:             inode.Meta.Name,
					ContentType:      dir.MimeTypeDirectory,
					CreationTime:     strconv.FormatInt(inode.Meta.CreationTime, 10),
					AccessTime:       strconv.FormatInt(inode.Meta.AccessTime, 10),
					ModificationTime: strconv.FormatInt(inode.Meta.ModificationTime, 10),
					Mode:             inode.Meta.Mode,
					DirList:          make([]*pod.DirSnapShot, 0),
					FileList:         make([]file.MetaData, 0),
				}
				dirL.DirList = append(dirL.DirList, dirItem)
				err := a.getSnapShotForDir(dirItem, fd, accountInfo, inode.FileOrDirNames, utils.CombinePathAndFile(dirNameWithPath, inode.Meta.Name), password)
				if err != nil {
					return err
				}
			}
		case fileEntry, ok := <-fileChan:
			if ok {
				dirL.FileList = append(dirL.FileList, fileEntry)
			}
		case err := <-errChan:
			if err != nil {
				return err
			}
		case <-time.After(time.Hour):
			return fmt.Errorf("timeout while listing directory")
		}
		if len(dirL.FileList)+len(dirL.DirList)+missingCount == len(fileOrDirNames) {
			break
		}
	}
	return nil
}

// PodReceive - receive a pod from a sharingReference
func (a *API) PodReceive(sessionId, sharedPodName string, ref utils.Reference) (*pod.Info, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePod(sharedPodName, ref)
}

// IsPodExist checks if a pod exists
func (a *API) IsPodExist(podName, sessionId string) bool {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false
	}
	return ui.GetPod().IsPodPresent(podName)
}

// ForkPod forks a pod
func (a *API) ForkPod(podName, forkName, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if forkName == "" {
		return pod.ErrBlankPodName
	}

	if ui.GetPod().IsPodPresent(forkName) {
		return pod.ErrForkAlreadyExists
	}

	_, err := a.prepareOwnPod(ui, forkName)
	if err != nil {
		return err
	}

	return ui.GetPod().PodFork(podName, forkName)
}

// ForkPodFromRef forks a pod from a sharing reference
func (a *API) ForkPodFromRef(forkName, refString, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if refString == "" {
		return pod.ErrBlankPodSharingReference
	}
	if forkName == "" {
		return pod.ErrBlankPodName
	}

	if ui.GetPod().IsPodPresent(forkName) {
		return pod.ErrForkAlreadyExists
	}

	_, err := a.prepareOwnPod(ui, forkName)
	if err != nil {
		return err
	}

	return ui.GetPod().PodForkFromRef(forkName, refString)
}

func (*API) prepareOwnPod(ui *user.Info, podName string) (*pod.Info, error) {
	podPasswordBytes, _ := utils.GetRandBytes(pod.PasswordLength)
	podPassword := hex.EncodeToString(podPasswordBytes)

	// create the pod
	_, err := ui.GetPod().CreatePod(podName, "", podPassword)
	if err != nil {
		return nil, err
	}

	// open the pod
	pi, err := ui.GetPod().OpenPod(podName)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// ListPodInMarketplace lists a pod in the datahub marketplace
func (a *API) ListPodInMarketplace(sessionId, podName, title, desc, thumbnail string, price uint64, daysValid uint16, category [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	nameHash, err := a.users.GetNameHash(ui.GetUserName())
	if err != nil {
		return err
	}

	return ui.GetPod().ListPodInMarketplace(podName, title, desc, thumbnail, price, daysValid, category, nameHash)
}

// ChangePodListStatusInMarketplace changes the status of a pod in the datahub marketplace
func (a *API) ChangePodListStatusInMarketplace(sessionId string, subHash [32]byte, show bool) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	return ui.GetPod().PodStatusInMarketplace(subHash, show)
}

// RequestSubscription requests a subscription to a pod
func (a *API) RequestSubscription(sessionId string, subHash [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	nameHash, err := a.users.GetNameHash(ui.GetUserName())
	if err != nil {
		return err
	}

	return ui.GetPod().RequestSubscription(subHash, nameHash)
}

// ApproveSubscription approves a subscription to a pod
func (a *API) ApproveSubscription(sessionId, podName string, reqHash, nameHash [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	_, subscriberPublicKey, err := a.users.GetUserInfoFromENS(nameHash)
	if err != nil {
		return err
	}

	return ui.GetPod().ApproveSubscription(podName, reqHash, subscriberPublicKey)
}

// EncryptSubscription encrypts the subscription information
func (a *API) EncryptSubscription(sessionId, podName string, nameHash [32]byte) (string, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return "", errNilSubManager
	}

	_, subscriberPublicKey, err := a.users.GetUserInfoFromENS(nameHash)
	if err != nil {
		return "", err
	}

	return ui.GetPod().EncryptUploadSubscriptionInfo(podName, subscriberPublicKey)
}

// DecryptAndOpenSubscriptionPod decrypts pod info and opens the subscription pod
func (a *API) DecryptAndOpenSubscriptionPod(sessionId, reference string, sellerNameHash [32]byte) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return nil, errNilSubManager
	}

	_, publicKey, err := a.users.GetUserInfoFromENS(sellerNameHash)
	if err != nil {
		return nil, err
	}

	pi, err := ui.GetPod().OpenSubscribedPodFromReference(reference, publicKey)
	if err != nil {
		return nil, err
	}

	err = pi.GetDirectory().AddRootDir(pi.GetPodName(), pi.GetPodPassword(), pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(pi.GetPodName(), pi)
	return pi, nil

}

// SubscriptionInfo contains the subscription information
type SubscriptionInfo struct {
	SubHash      [32]byte
	PodName      string
	PodAddress   string
	Category     string
	InfoLocation []byte
	ValidTill    int64
}

// GetSubscriptions returns the list of subscriptions
func (a *API) GetSubscriptions(sessionId string) ([]SubscriptionInfo, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return nil, errNilSubManager
	}

	nameHash, err := a.users.GetNameHash(ui.GetUserName())
	if err != nil {
		return nil, err
	}

	subscriptions, err := ui.GetPod().GetSubscriptions(nameHash)
	if err != nil {
		return nil, err
	}

	subs := make([]SubscriptionInfo, len(subscriptions))
	for i, item := range subscriptions {
		info, err := ui.GetPod().GetSubscribablePodInfo(item.SubHash)
		if err != nil {
			return subs, err
		}
		var infoLocation = make([]byte, 32)
		copy(infoLocation, item.UnlockKeyLocation[:])
		sub := SubscriptionInfo{
			SubHash:      item.SubHash,
			PodName:      info.PodName,
			PodAddress:   info.PodAddress,
			InfoLocation: infoLocation,
			ValidTill:    item.ValidTill.Int64(),
			Category:     info.Category,
		}

		subs[i] = sub
	}

	return subs, nil
}

// GetSubscribablePodInfo returns the subscribable pod info
func (a *API) GetSubscribablePodInfo(sessionId string, subHash [32]byte) (*rpc.SubscriptionItemInfo, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	if a.sm == nil {
		return nil, errNilSubManager
	}
	return a.sm.GetSubscribablePodInfo(subHash)
}

// OpenSubscribedPod opens the subscribed pod
func (a *API) OpenSubscribedPod(sessionId string, subHash [32]byte, infoLocation string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return nil, errNilSubManager
	}
	sub, err := a.sm.GetSub(subHash)
	if err != nil {
		return nil, err
	}

	subHashString := utils.Encode(subHash[:])

	_, ownerPublicKey, err := a.users.GetUserInfoFromENS(sub.FdpSellerNameHash)
	if err != nil {
		return nil, err
	}

	// open the pod
	pi, err := ui.GetPod().OpenSubscribedPodFromReference(infoLocation, ownerPublicKey)
	if err != nil {
		return nil, err
	}
	err = pi.GetDirectory().AddRootDir(pi.GetPodName(), pi.GetPodPassword(), pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName("0x"+subHashString, pi)
	return pi, nil
}

// GetSubscribablePods returns the list of subscribable pods
func (a *API) GetSubscribablePods(sessionId string) ([]datahub.DataHubSub, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	if a.sm == nil {
		return nil, errNilSubManager
	}
	return ui.GetPod().GetMarketplace()
}

// GetSubsRequests returns the list of subscription requests
func (a *API) GetSubsRequests(sessionId string) ([]datahub.DataHubSubRequest, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	if a.sm == nil {
		return nil, errNilSubManager
	}
	return ui.GetPod().GetSubRequests()
}
