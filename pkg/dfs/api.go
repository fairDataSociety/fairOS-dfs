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
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type DfsAPI struct {
	dataDir string
	client  blockstore.Client
	users   *user.Users
	logger  logging.Logger
}

func NewDfsAPI(dataDir, host, port, cookieDomain string, logger logging.Logger) (*DfsAPI, error) {
	c := bee.NewBeeClient(host, port, logger)
	if !c.CheckConnection() {
		return nil, ErrBeeClient
	}
	users := user.NewUsers(dataDir, c, cookieDomain, logger)
	return &DfsAPI{
		dataDir: dataDir,
		client:  c,
		users:   users,
		logger:  logger,
	}, nil
}

//
//  User related APIs
//
func (d *DfsAPI) CreateUser(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, string, error) {
	if !d.client.CheckConnection() {
		return "", "", ErrBeeClient
	}

	reference, rcvdMnemonic, userInfo, err := d.users.CreateNewUser(userName, passPhrase, mnemonic, d.dataDir, d.client, response, sessionId)
	if err != nil {
		return reference, rcvdMnemonic, err
	}

	err = d.users.CreateRootFeeds(userInfo)
	if err != nil {
		return reference, rcvdMnemonic, err
	}
	return reference, rcvdMnemonic, nil
}

func (d *DfsAPI) ImportUserUsingMnemonic(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, error) {
	reference, _, err := d.CreateUser(userName, passPhrase, mnemonic, response, sessionId)
	return reference, err
}

func (d *DfsAPI) ImportUserUsingAddress(userName, passPhrase, address string, response http.ResponseWriter, sessionId string) error {
	return d.users.ImportUsingAddress(userName, passPhrase, address, d.dataDir, d.client, response, sessionId)
}

func (d *DfsAPI) LoginUser(userName, passPhrase string, response http.ResponseWriter, sessionId string) error {
	return d.users.LoginUser(userName, passPhrase, d.dataDir, d.client, response, sessionId)
}

func (d *DfsAPI) LogoutUser(sessionId string, response http.ResponseWriter) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.LogoutUser(ui.GetUserName(), d.dataDir, sessionId, response)
}

func (d *DfsAPI) DeleteUser(passPhrase, sessionId string, response http.ResponseWriter) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.DeleteUser(ui.GetUserName(), d.dataDir, passPhrase, sessionId, response, ui)
}

func (d *DfsAPI) IsUserNameAvailable(userName string) bool {
	return d.users.IsUsernameAvailable(userName, d.dataDir)
}

func (d *DfsAPI) IsUserLoggedIn(userName string) bool {
	// check if a given user is logged in
	return d.users.IsUserNameLoggedIn(userName)
}

func (d *DfsAPI) ListAllUsers() ([]string, error) {
	return d.users.ListAllUsers(d.dataDir)
}

func (d *DfsAPI) SaveAvatar(sessionId string, data []byte) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.SaveAvatar(data, ui)
}

func (d *DfsAPI) GetAvatar(sessionId string) ([]byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return d.users.GetAvatar(ui)
}

func (d *DfsAPI) SaveName(firstName, lastName, middleName, surname, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return d.users.SaveName(firstName, lastName, middleName, surname, ui)
}

func (d *DfsAPI) GetName(sessionId string) (*user.Name, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetName(ui)
}

func (d *DfsAPI) SaveContact(phone, mobile string, address *user.Address, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return d.users.SaveContacts(phone, mobile, address, ui)
}

func (d *DfsAPI) GetContact(sessionId string) (*user.Contacts, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetContacts(ui)
}

func (d *DfsAPI) GetUserStat(sessionId string) (*user.Stat, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetUserStat(ui)
}

func (d *DfsAPI) GetUserSharingInbox(sessionId string) (*user.Inbox, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetSharingInbox(ui)
}

func (d *DfsAPI) GetUserSharingOutbox(sessionId string) (*user.Outbox, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetSharingOutbox(ui)
}

func (d *DfsAPI) ExportUser(sessionId string) (string, string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", "", ErrUserNotLoggedIn
	}
	return d.users.ExportUser(ui)
}

//
//  Pods related APIs
//
func (d *DfsAPI) CreatePod(podName, passPhrase, sessionId string) (*pod.Info, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// create the pod
	pi, err := ui.GetPod().CreatePod(podName, passPhrase)
	if err != nil {
		return nil, err
	}

	// open the pod
	_, err = ui.GetPod().OpenPod(podName, passPhrase)
	if err != nil {
		return nil, err
	}

	// Add podName in the login user session
	ui.SetPodName(podName)
	return pi, nil
}

func (d *DfsAPI) DeletePod(podName, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// delete the pod and close if it is opened
	err := ui.GetPod().DeletePod(podName)
	if err != nil {
		return err
	}

	// close the pod and delete it from login user session, if the delete is for a opened pod
	if ui.GetPodName() != "" && podName == ui.GetPodName() {
		// remove from the login session
		ui.RemovePodName()
	}

	return nil
}

func (d *DfsAPI) OpenPod(podName, passPhrase, sessionId string) (*pod.Info, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// close the already open pod
	if ui.GetPodName() != "" {
		err := ui.GetPod().ClosePod(ui.GetPodName())
		if err != nil {
			return nil, err
		}
	}

	// open the pod
	po, err := ui.GetPod().OpenPod(podName, passPhrase)
	if err != nil {
		return nil, err
	}

	// Add podName in the login user session
	ui.SetPodName(podName)
	return po, nil
}

func (d *DfsAPI) ClosePod(sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	// close the pod
	err := ui.GetPod().ClosePod(ui.GetPodName())
	if err != nil {
		return err
	}

	// delete podName in the login user session
	ui.RemovePodName()
	return nil
}

func (d *DfsAPI) PodStat(podName, sessionId string) (*pod.PodStat, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
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

func (d *DfsAPI) SyncPod(sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	// sync the pod
	err := ui.GetPod().SyncPod(ui.GetPodName())
	if err != nil {
		return err
	}
	return nil
}

func (d *DfsAPI) ListPods(sessionId string) ([]string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// list pods of a user
	pods, err := ui.GetPod().ListPods()
	if err != nil {
		return nil, err
	}
	return pods, nil
}

//
//  Directory related APIs
//

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

//
//  KV related APIs
//

func (d *DfsAPI) KVCreate(sessionId, name string) error {
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

	return podInfo.GetCollection().CreateKVTable(name)
}

func (d *DfsAPI) KVDelete(sessionId, name string) error {
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

	return podInfo.GetCollection().DeleteKVTable(name)
}

func (d *DfsAPI) KVOpen(sessionId, name string) error {
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

	return podInfo.GetCollection().OpenKVTable(name)
}

func (d *DfsAPI) KVList(sessionId string) (map[string][]string, error) {
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

	return podInfo.GetCollection().LoadKVTables()
}

func (d *DfsAPI) KVCount(sessionId, name string) (uint64, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return 0, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return 0, err
	}

	return podInfo.GetCollection().KVCount(name)
}

func (d *DfsAPI) KVPut(sessionId, name, key string, value []byte) error {
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

	return podInfo.GetCollection().KVPut(name, key, value)
}

func (d *DfsAPI) KVGet(sessionId, name, key string) ([]string, []byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, nil, err
	}

	return podInfo.GetCollection().KVGet(name, key)
}

func (d *DfsAPI) KVDel(sessionId, name, key string) ([]byte, error) {
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

	return podInfo.GetCollection().KVDelete(name, key)
}

func (d *DfsAPI) KVBatch(sessionId, name string, columns []string) (*collection.Batch, error) {
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

	return podInfo.GetCollection().KVBatch(name, columns)
}

func (d *DfsAPI) KVBatchPut(sessionId, key string, value []byte, batch *collection.Batch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	return batch.Put(key, value)
}

func (d *DfsAPI) KVBatchWrite(sessionId string, batch *collection.Batch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	return batch.Write()
}

func (d *DfsAPI) KVSeek(sessionId, name, start, end string, limit int64) (*collection.Iterator, error) {
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

	return podInfo.GetCollection().KVSeek(name, start, end, limit)
}

func (d *DfsAPI) KVGetNext(sessionId, name string) ([]string, string, []byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, "", nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, "", nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, "", nil, err
	}

	return podInfo.GetCollection().KVGetNext(name)
}
