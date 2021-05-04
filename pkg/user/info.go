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
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

type Info struct {
	name      string
	podName   string
	userAddress string
	sessionId string
	feedApi   *feed.API
	account   *account.Account
	file      *f.File
	dir       *d.Directory
	pod       *pod.Pod
}

func (i *Info) GetUserName() string {
	return i.name
}

func (i *Info) GetSessionId() string {
	return i.sessionId
}


func (i *Info) GetUserAddress() string {
	return i.userAddress
}

func (i *Info) GetPod() *pod.Pod {
	return i.pod
}

func (i *Info) GetAccount() *account.Account {
	return i.account
}

func (i *Info) GetFeed() *feed.API {
	return i.feedApi
}

func (i *Info) SetPodName(podName string) {
	i.podName = podName
}

func (i *Info) RemovePodName() {
	i.podName = ""
}

func (i *Info) GetPodName() string {
	return i.podName
}
