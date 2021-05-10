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

type UserRequest struct {
	UserName string `json:"user_name,omitempty"`
	Password string `json:"password,omitempty"`
	Address  string `json:"address,omitempty"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

type PodRequest struct {
	PodName   string `json:"pod_name,omitempty"`
	Password  string `json:"password,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type FileSystemRequest struct {
	DirectoryPath string `json:"dir_path,omitempty"`
	DirectoryName string `json:"dir_name,omitempty"`
	FilePath      string `json:"file_path,omitempty"`
	FileName      string `json:"file_name,omitempty"`
	Destination   string `json:"dest_user,omitempty"`
}

type KVRequest struct {
	TableName   string `json:"table_name,omitempty"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	StartPrefix string `json:"start_prefix,omitempty"`
	EndPrefix   string `json:"end_prefix,omitempty"`
	Limit       string `json:"limit,omitempty"`
}

type DocRequest struct {
	TableName     string `json:"table_name,omitempty"`
	ID            string `json:"id,omitempty"`
	Document      string `json:"doc,omitempty"`
	SimpleIndex   string `json:"si,omitempty"`
	CompoundIndex string `json:"ci,omitempty"`
	Expression    string `json:"expr,omitempty"`
	Mutable       string `json:"mutable,omitempty"`
	Limit         string `json:"limit,omitempty"`
	FileName      string `json:"file_name,omitempty"`
}
