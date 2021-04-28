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

package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	inboxFeedName  = "inbox"
	outboxFeedName = "outbox"
)

type Inbox struct {
	Entries []SharingEntry `json:"entries"`
}

type Outbox struct {
	Entries []SharingEntry `json:"entries"`
}

type SharingEntry struct {
	FileName     string `json:"name"`
	PodName      string `json:"pod_name"`
	FileMetaHash string `json:"meta_ref"`
	Sender       string `json:"source_address"`
	Receiver     string `json:"dest_address"`
	SharedTime   string `json:"shared_time"`
}

type ReceiveFileInfo struct {
	FileName       string `json:"name"`
	Size           string `json:"size"`
	BlockSize      string `json:"block_size"`
	NumberOfBlocks string `json:"number_of_blocks"`
	ContentType    string `json:"content_type"`
	Compression    string `json:"compression"`
	PodName        string `json:"pod_name"`
	FileMetaHash   string `json:"meta_ref"`
	Sender         string `json:"source_address"`
	Receiver       string `json:"dest_address"`
	SharedTime     string `json:"shared_time"`
}

func (u *Users) ShareFileWithUser(podName, podFilePath, destinationRef string, userInfo *Info, pod *pod.Pod) (string, error) {
	// KVGet the meta reference of the file to share
	metaRef, fileName, err := pod.GetMetaReferenceOfFile(podName, podFilePath)
	if err != nil {
		return "", err
	}

	// Create a outbox entry
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	now := time.Now()
	sharingEntry := SharingEntry{
		FileName:     fileName,
		PodName:      userInfo.podName,
		FileMetaHash: utils.NewReference(metaRef).String(),
		Sender:       rootReference.String(),
		Receiver:     destinationRef,
		SharedTime:   strconv.FormatInt(now.Unix(), 10),
	}

	// get the outbox reference from outbox feed
	outboxRef, err := getFeedData(outboxFeedName, rootReference, userInfo.GetFeed())
	if err != nil {
		return "", err
	}

	// download the entire outbox file
	outboxFileBytes, respCode, err := u.client.DownloadBlob(outboxRef)
	if err != nil && respCode != http.StatusOK {
		return "", err
	}

	// unmarshall, add a new outbox entry, marshall the data again
	outbox := &Outbox{}
	err = json.Unmarshal(outboxFileBytes, outbox)
	if err != nil {
		return "", err
	}
	outbox.Entries = append(outbox.Entries, sharingEntry)
	outData, err := json.Marshal(outbox)
	if err != nil {
		return "", err
	}

	// store the new outbox file data
	newOutboxRef, err := u.client.UploadBlob(outData, true, true)
	if err != nil {
		return "", err
	}

	// update the outbox feed with the new outbox file reference
	err = putFeedData(outboxFeedName, rootReference, newOutboxRef, userInfo.GetFeed())
	if err != nil {
		return "", err
	}

	// marshall the entry
	data, err := json.Marshal(sharingEntry)
	if err != nil {
		return "", err
	}

	//encrypt data
	encryptedData, err := encryptData(data, now.Unix())
	if err != nil {
		return "", err
	}

	// upload the encrypted data and get the reference
	ref, err := u.client.UploadBlob(encryptedData, true, true)
	if err != nil {
		return "", err
	}

	// add now to the ref
	sharingRef := utils.NewSharingReference(ref, now.Unix())
	return sharingRef.String(), nil
}

func (u *Users) ReceiveFileFromUser(podName string, sharingRef utils.SharingReference, userInfo *Info, pod *pod.Pod, podDir string) (string, string, error) {
	metaRef := sharingRef.GetRef()
	unixTime := sharingRef.GetNonce()

	// get the encrypted meta
	encryptedData, respCode, err := u.client.DownloadBlob(metaRef)
	if err != nil || respCode != http.StatusOK {
		return "", "", err
	}

	// decrypt the data
	decryptedData, err := decryptData(encryptedData, unixTime)
	if err != nil {
		return "", "", err
	}

	// unmarshall the entry
	sharingEntry := SharingEntry{}
	err = json.Unmarshal(decryptedData, &sharingEntry)
	if err != nil {
		return "", "", err
	}

	// add the file to the pod directory specified
	fileName := sharingEntry.FileName
	sharingEntry.PodName = podName

	// check if pod is open
	if !pod.IsPodOpened(podName) {
		return "", "", fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := pod.GetPodInfoFromPodMap(podName)
	if err != nil {
		return "", "", err
	}

	path := pod.GetFilePath(podDir, podInfo)
	dir := podInfo.GetDirectory()

	_, dirInode, err := dir.GetDirNode(path, podInfo.GetFeed(), podInfo.GetAccountInfo().GetAddress())
	if err != nil {
		return "", "", err
	}

	// check if file is already present
	if podInfo.GetFile().IsFileAlreadyPresent(path) {
		return "", "", fmt.Errorf("file already present in the destination dir")
	}

	// modify the modification time
	dirInode.Meta.ModificationTime = time.Now().Unix()
	_, err = dir.UpdateDirectory(dirInode)
	if err != nil {
		return "", "", err
	}

	// Add to file path map
	err = podInfo.GetFile().AddFileToPath(path, sharingEntry.FileMetaHash)
	if err != nil {
		return "", "", err
	}

	// get the inbox reference from inbox feed
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	inboxRef, err := getFeedData(inboxFeedName, rootReference, userInfo.GetFeed())
	if err != nil {
		return "", "", err
	}

	// download the entire inbox file
	inboxFileBytes, respCode, err := u.client.DownloadBlob(inboxRef)
	if err != nil && respCode != http.StatusOK {
		return "", "", err
	}

	// unmarshall, add a new inbox entry, marshall the data again
	inbox := &Inbox{}
	err = json.Unmarshal(inboxFileBytes, inbox)
	if err != nil {
		return "", "", err
	}
	inbox.Entries = append(inbox.Entries, sharingEntry)
	inData, err := json.Marshal(inbox)
	if err != nil {
		return "", "", err
	}

	// store the new inbox file data
	newInboxRef, err := u.client.UploadBlob(inData, true, true)
	if err != nil {
		return "", "", err
	}

	// update the inbox feed with the new inbox file reference
	err = putFeedData(inboxFeedName, rootReference, newInboxRef, userInfo.GetFeed())
	if err != nil {
		return "", "", err
	}

	if podDir == utils.PathSeperator {
		podDir = ""
	}

	return podDir + utils.PathSeperator + fileName, sharingEntry.FileMetaHash, nil
}

func (u *Users) GetSharingInbox(userInfo *Info) (*Inbox, error) {
	// get the inbox reference from the inbox feed
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	inboxRef, err := getFeedData(inboxFeedName, rootReference, userInfo.GetFeed())
	if err != nil {
		return nil, err
	}

	if len(inboxRef) < utils.ReferenceLength {
		return nil, fmt.Errorf("empty inbox")
	}

	// download the entire inbox file
	inboxFileBytes, respCode, err := u.client.DownloadBlob(inboxRef)
	if err != nil && respCode != http.StatusOK {
		return nil, err
	}

	// unmarshall it and return the entire structure
	inbox := &Inbox{}
	err = json.Unmarshal(inboxFileBytes, inbox)
	if err != nil {
		return nil, err
	}
	return inbox, nil
}

func (u *Users) GetSharingOutbox(userInfo *Info) (*Outbox, error) {
	// get the outbox reference from the inbox feed
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	outboxRef, err := getFeedData(outboxFeedName, rootReference, userInfo.GetFeed())
	if err != nil {
		return nil, err
	}

	if len(outboxRef) < utils.ReferenceLength {
		return nil, fmt.Errorf("empty outbox")
	}

	// download the entire outbox file
	outboxFileBytes, respCode, err := u.client.DownloadBlob(outboxRef)
	if err != nil && respCode != http.StatusOK {
		return nil, err
	}

	outbox := &Outbox{}
	err = json.Unmarshal(outboxFileBytes, outbox)
	if err != nil {
		return nil, err
	}
	return outbox, nil
}

func encryptData(data []byte, now int64) ([]byte, error) {
	pk, err := account.CreateRandomKeyPair(now)
	if err != nil {
		return nil, err
	}
	pubKey := btcec.PublicKey{Curve: pk.PublicKey.Curve, X: pk.PublicKey.X, Y: pk.PublicKey.Y}
	return btcec.Encrypt(&pubKey, data)
}

func decryptData(data []byte, now int64) ([]byte, error) {
	pk, err := account.CreateRandomKeyPair(now)
	if err != nil {
		return nil, err
	}
	privateKey := btcec.PrivateKey{PublicKey: pk.PublicKey, D: pk.D}
	return btcec.Decrypt(&privateKey, data)
}

func (u *Users) ReceiveFileInfo(podName string, sharingRef utils.SharingReference, userInfo *Info, pod *pod.Pod) (*ReceiveFileInfo, error) {
	metaRef := sharingRef.GetRef()
	unixTime := sharingRef.GetNonce()

	// get the encrypted meta
	encryptedData, respCode, err := u.client.DownloadBlob(metaRef)
	if err != nil || respCode != http.StatusOK {
		return nil, err
	}

	// decrypt the data
	decryptedData, err := decryptData(encryptedData, unixTime)
	if err != nil {
		return nil, err
	}

	// unmarshall the entry
	sharingEntry := SharingEntry{}
	err = json.Unmarshal(decryptedData, &sharingEntry)
	if err != nil {
		return nil, err
	}

	fileMetaRef, err := utils.ParseHexReference(sharingEntry.FileMetaHash)
	if err != nil {
		return nil, err
	}

	fileMetaBytes, respCode, err := u.client.DownloadBlob(fileMetaRef.Bytes())
	if err != nil || respCode != http.StatusOK {
		return nil, err
	}

	meta := f.MetaData{}
	err = json.Unmarshal(fileMetaBytes, &meta)
	if err != nil {
		return nil, err
	}
	compression := meta.Compression
	if compression == "" {
		compression = "None"
	}

	fileInodeBytes, respCode, err := u.client.DownloadBlob(meta.InodeAddress)
	if err != nil || respCode != http.StatusOK {
		return nil, err
	}
	var fileInode f.INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil {
		return nil, err
	}

	info := ReceiveFileInfo{
		FileName:       meta.Name,
		Size:           strconv.FormatInt(int64(meta.Size), 10),
		BlockSize:      strconv.FormatInt(int64(meta.BlockSize), 10),
		NumberOfBlocks: strconv.FormatInt(int64(len(fileInode.Blocks)), 10),
		ContentType:    meta.ContentType,
		Compression:    compression,
		PodName:        sharingEntry.PodName,
		FileMetaHash:   sharingEntry.FileMetaHash,
		Sender:         sharingEntry.Sender,
		Receiver:       sharingEntry.Receiver,
		SharedTime:     sharingEntry.SharedTime,
	}
	return &info, nil
}
