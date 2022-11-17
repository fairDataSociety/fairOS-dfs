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

package common

type UserSignupRequest struct {
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

type UserLoginRequest struct {
	UserName string `json:"userName,omitempty"`
	Password string `json:"password,omitempty"`
}

type PodRequest struct {
	PodName       string `json:"podName,omitempty"`
	Password      string `json:"password,omitempty"`
	Reference     string `json:"reference,omitempty"`
	SharedPodName string `json:"sharedPodName,omitempty"`
}

type PodShareRequest struct {
	PodName       string `json:"podName,omitempty"`
	SharedPodName string `json:"sharedPodName,omitempty"`
}

type PodReceiveRequest struct {
	PodName       string `json:"podName,omitempty"`
	Reference     string `json:"sharingRef,omitempty"`
	SharedPodName string `json:"sharedPodName,omitempty"`
}

type FileSystemRequest struct {
	PodName       string `json:"podName,omitempty"`
	DirectoryPath string `json:"dirPath,omitempty"`
	DirectoryName string `json:"dirName,omitempty"`
	FilePath      string `json:"filePath,omitempty"`
	FileName      string `json:"fileName,omitempty"`
	Destination   string `json:"destUser,omitempty"`
}

type RenameRequest struct {
	PodName string `json:"podName,omitempty"`
	OldPath string `json:"oldPath,omitempty"`
	NewPath string `json:"newPath,omitempty"`
}

type FileReceiveRequest struct {
	PodName          string `json:"podName,omitempty"`
	SharingReference string `json:"sharingRef,omitempty"`
	DirectoryPath    string `json:"dirPath,omitempty"`
}

type KVRequest struct {
	PodName     string `json:"podName,omitempty"`
	TableName   string `json:"tableName,omitempty"`
	IndexType   string `json:"indexType,omitempty"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	StartPrefix string `json:"startPrefix,omitempty"`
	EndPrefix   string `json:"endPrefix,omitempty"`
	Limit       string `json:"limit,omitempty"`
	Memory      string `json:"memory,omitempty"`
}

type DocRequest struct {
	PodName       string `json:"podName,omitempty"`
	TableName     string `json:"tableName,omitempty"`
	ID            string `json:"id,omitempty"`
	Document      string `json:"doc,omitempty"`
	SimpleIndex   string `json:"si,omitempty"`
	CompoundIndex string `json:"ci,omitempty"`
	Expression    string `json:"expr,omitempty"`
	Mutable       bool   `json:"mutable,omitempty"`
	Limit         string `json:"limit,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}
