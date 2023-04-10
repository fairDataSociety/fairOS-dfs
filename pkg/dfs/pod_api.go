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
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/datahub"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// CreatePod
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
	if err != nil {
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

// OpenPod
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

// ClosePod
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

// PodStat
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

// SyncPod
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

// SyncPodAsync
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

// ListPods
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

// PodShare
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

// PodReceiveInfo
func (a *API) PodReceiveInfo(sessionId string, ref utils.Reference) (*pod.ShareInfo, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePodInfo(ref)
}

// PublicPodReceiveInfo
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

// PublicPodFileDownload
func (a *API) PublicPodFileDownload(pod *pod.ShareInfo, filePath string) (io.ReadCloser, uint64, error) {

	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.logger)
	topic := utils.HashString(filePath)
	_, metaBytes, err := fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password))
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

// PublicPodKVEntryGet
func (a *API) PublicPodKVEntryGet(pod *pod.ShareInfo, name, key string) ([]string, []byte, error) {

	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.logger)
	kvStore := c.NewKeyValueStore(pod.PodName, fd, accountInfo, address, a.client, a.logger)

	err := kvStore.OpenKVTable(name, pod.Password)
	if err != nil {
		return nil, nil, err
	}

	return kvStore.KVGet(name, key)
}

// PublicPodFileDownload
func (a *API) PublicPodDisLs(pod *pod.ShareInfo, dirPathToLs string) ([]dir.Entry, []file.Entry, error) {

	accountInfo := &account.Info{}
	address := utils.HexToAddress(pod.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, a.client, a.logger)

	dirNameWithPath := filepath.ToSlash(dirPathToLs)
	topic := utils.HashString(dirNameWithPath)
	_, data, err := fd.GetFeedData(topic, accountInfo.GetAddress(), []byte(pod.Password))
	if err != nil { // skipcq: TCV-001
		if dirNameWithPath == utils.PathSeparator {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("list dir : %v", err) // skipcq: TCV-001
	}

	dirInode := &dir.Inode{}
	err = dirInode.Unmarshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("list dir : %v", err)
	}

	listEntries := []dir.Entry{}
	var files []string
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			dirPath := utils.CombinePathAndFile(dirNameWithPath, dirName)
			dirTopic := utils.HashString(dirPath)

			_, data, err := fd.GetFeedData(dirTopic, accountInfo.GetAddress(), []byte(pod.Password))
			if err != nil { // skipcq: TCV-001
				return nil, nil, fmt.Errorf("list dir : %v", err)
			}
			var dirInode *dir.Inode
			err = json.Unmarshal(data, &dirInode)
			if err != nil { // skipcq: TCV-001
				return nil, nil, fmt.Errorf("list dir : %v", err)
			}
			entry := dir.Entry{
				Name:             dirInode.Meta.Name,
				ContentType:      dir.MimeTypeDirectory, // per RFC2425
				CreationTime:     strconv.FormatInt(dirInode.Meta.CreationTime, 10),
				AccessTime:       strconv.FormatInt(dirInode.Meta.AccessTime, 10),
				ModificationTime: strconv.FormatInt(dirInode.Meta.ModificationTime, 10),
				Mode:             dirInode.Meta.Mode,
			}
			listEntries = append(listEntries, entry)
		} else if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			files = append(files, filePath)
		}
	}

	fileEntries := []file.Entry{}
	for _, filePath := range files {
		fileTopic := utils.HashString(utils.CombinePathAndFile(filePath, ""))

		_, data, err := fd.GetFeedData(fileTopic, accountInfo.GetAddress(), []byte(pod.Password))
		if err != nil { // skipcq: TCV-001
			return nil, nil, fmt.Errorf("file mtdt : %v", err)
		}
		if string(data) == utils.DeletedFeedMagicWord { // skipcq: TCV-001
			continue
		}
		var meta *file.MetaData
		err = json.Unmarshal(data, &meta)
		if err != nil { // skipcq: TCV-001
			return nil, nil, fmt.Errorf("file mtdt : %v", err)
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

		fileEntries = append(fileEntries, entry)
	}

	return listEntries, fileEntries, nil
}

// PodReceive
func (a *API) PodReceive(sessionId, sharedPodName string, ref utils.Reference) (*pod.Info, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePod(sharedPodName, ref)
}

// IsPodExist
func (a *API) IsPodExist(podName, sessionId string) bool {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false
	}
	return ui.GetPod().IsPodPresent(podName)
}

// ForkPod
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

// ForkPodFromRef
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

// ListPodInMarketplace
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

// ChangePodListStatusInMarketplace
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

// RequestSubscription
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

// ApproveSubscription
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

// EncryptSubscription
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

// DecryptAndOpenSubscriptionPod
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

type SubscriptionInfo struct {
	SubHash      [32]byte
	PodName      string
	PodAddress   string
	Category     string
	InfoLocation []byte
	ValidTill    int64
}

// GetSubscriptions
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

// GetSubscribablePodInfo
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

// OpenSubscribedPod
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

// GetSubscribablePods
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

// GetSubscribablePods
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
