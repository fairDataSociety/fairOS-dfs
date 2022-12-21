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
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type SharingEntry struct {
	Meta       *f.MetaData `json:"meta"`
	Sender     string      `json:"sourceAddress"`
	Receiver   string      `json:"destAddress"`
	SharedTime string      `json:"sharedTime"`
}

type SharingMetaData struct {
	Version          uint8  `json:"version"`
	Path             string `json:"filePath"`
	Name             string `json:"fileName"`
	SharedPassword   string `json:"sharedPassword"`
	Size             uint64 `json:"fileSize"`
	BlockSize        uint32 `json:"blockSize"`
	ContentType      string `json:"contentType"`
	Compression      string `json:"compression"`
	CreationTime     int64  `json:"creationTime"`
	AccessTime       int64  `json:"accessTime"`
	ModificationTime int64  `json:"modificationTime"`
	InodeAddress     []byte `json:"fileInodeReference"`
}

type ReceiveFileInfo struct {
	FileName       string `json:"name"`
	Size           string `json:"size"`
	BlockSize      string `json:"blockSize"`
	NumberOfBlocks string `json:"numberOfBlocks"`
	ContentType    string `json:"contentType"`
	Compression    string `json:"compression"`
	Sender         string `json:"sourceAddress"`
	Receiver       string `json:"destAddress"`
	SharedTime     string `json:"sharedTime"`
}

// ShareFileWithUser exports a file to another user by creating and uploading a new encrypted sharing file entry.
func (u *Users) ShareFileWithUser(podName, podPassword, podFileWithPath, destinationRef string, userInfo *Info, pod *pod.Pod, userAddress utils.Address) (string, error) {
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	meta, err := userInfo.file.GetMetaFromFileName(totalFilePath, podPassword, userAddress)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// Create an outbox entry
	now := time.Now()
	sharingEntry := SharingEntry{
		Meta:       meta,
		Sender:     userAddress.String(),
		Receiver:   destinationRef,
		SharedTime: strconv.FormatInt(now.Unix(), 10),
	}

	// marshall the entry
	data, err := json.Marshal(sharingEntry)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// encrypt data
	encryptedData, err := encryptData(data, now.Unix())
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// upload the encrypted data and get the reference
	ref, err := u.client.UploadBlob(encryptedData, true, true)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// add now to the ref
	sharingRef := utils.NewSharingReference(ref, now.Unix())
	return sharingRef.String(), nil
}

// ReceiveFileFromUser imports a exported file in to the current user and pod by reading the sharing file entry.
func (u *Users) ReceiveFileFromUser(podName string, sharingRef utils.SharingReference, userInfo *Info, pd *pod.Pod, podDir string) (string, error) {
	metaRef := sharingRef.GetRef()
	unixTime := sharingRef.GetNonce()

	// get the encrypted meta
	encryptedData, respCode, err := u.client.DownloadBlob(metaRef)
	if err != nil || respCode != http.StatusOK {
		return "", err
	} // skipcq: TCV-001

	// decrypt the data
	decryptedData, err := decryptData(encryptedData, unixTime)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// unmarshall the entry
	sharingEntry := SharingEntry{}
	err = json.Unmarshal(decryptedData, &sharingEntry)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	// check if pod is open
	if !pd.IsPodOpened(podName) {
		return "", pod.ErrPodNotOpened
	}

	podInfo, _, err := pd.GetPodInfoFromPodMap(podName)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	fileNameToAdd := sharingEntry.Meta.Name
	dir := podInfo.GetDirectory()
	file := podInfo.GetFile()
	totalPath := utils.CombinePathAndFile(podDir, fileNameToAdd)

	// check if file is already present
	if file.IsFileAlreadyPresent(totalPath) {
		return "", f.ErrFileAlreadyPresent
	}

	// Add to file path map
	now := time.Now().Unix()
	newMeta := f.MetaData{
		Version:          sharingEntry.Meta.Version,
		Path:             podDir,
		Name:             fileNameToAdd,
		Size:             sharingEntry.Meta.Size,
		BlockSize:        sharingEntry.Meta.BlockSize,
		ContentType:      sharingEntry.Meta.ContentType,
		Compression:      sharingEntry.Meta.Compression,
		CreationTime:     now,
		AccessTime:       now,
		ModificationTime: now,
		InodeAddress:     sharingEntry.Meta.InodeAddress,
	}

	file.AddToFileMap(totalPath, &newMeta)
	err = file.PutMetaForFile(&newMeta, podInfo.GetPodPassword())
	if err != nil { // skipcq: TCV-001
		return "", err
	}
	err = dir.AddEntryToDir(podDir, podInfo.GetPodPassword(), fileNameToAdd, true)
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	return totalPath, nil
}

func encryptData(data []byte, now int64) ([]byte, error) {
	pk, err := account.CreateRandomKeyPair(now)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	pubKey := btcec.PublicKey{Curve: pk.PublicKey.Curve, X: pk.PublicKey.X, Y: pk.PublicKey.Y}
	return btcec.Encrypt(&pubKey, data)
}

func decryptData(data []byte, now int64) ([]byte, error) {
	pk, err := account.CreateRandomKeyPair(now)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	privateKey := btcec.PrivateKey{PublicKey: pk.PublicKey, D: pk.D}
	return btcec.Decrypt(&privateKey, data)
}

// ReceiveFileInfo displays the information of the exported file. This is used to decide whether
// to import the file or not.
func (u *Users) ReceiveFileInfo(sharingRef utils.SharingReference) (*ReceiveFileInfo, error) {
	metaRef := sharingRef.GetRef()
	unixTime := sharingRef.GetNonce()

	// get the encrypted meta
	encryptedData, respCode, err := u.client.DownloadBlob(metaRef)
	if err != nil || respCode != http.StatusOK { // skipcq: TCV-001
		return nil, err
	}

	// decrypt the data
	decryptedData, err := decryptData(encryptedData, unixTime)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// unmarshall the entry
	sharingEntry := SharingEntry{}
	err = json.Unmarshal(decryptedData, &sharingEntry)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	fileInodeBytes, respCode, err := u.client.DownloadBlob(sharingEntry.Meta.InodeAddress)
	if err != nil || respCode != http.StatusOK { // skipcq: TCV-001
		return nil, err
	}

	var fileInode f.INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	info := ReceiveFileInfo{
		FileName:       sharingEntry.Meta.Name,
		Size:           strconv.FormatInt(int64(sharingEntry.Meta.Size), 10),
		BlockSize:      strconv.FormatInt(int64(sharingEntry.Meta.BlockSize), 10),
		NumberOfBlocks: strconv.FormatInt(int64(len(fileInode.Blocks)), 10),
		ContentType:    sharingEntry.Meta.ContentType,
		Compression:    sharingEntry.Meta.Compression,
		Sender:         sharingEntry.Sender,
		Receiver:       sharingEntry.Receiver,
		SharedTime:     sharingEntry.SharedTime,
	}
	return &info, nil
}
