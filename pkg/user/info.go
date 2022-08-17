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
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// Info is user information on fairOS
type Info struct {
	name       string
	sessionId  string
	feedApi    *feed.API
	account    *account.Account
	file       *f.File
	dir        *d.Directory
	pod        *pod.Pod
	openPods   map[string]*pod.Info
	openPodsMu *sync.RWMutex
}

// GetUserName return username
func (i *Info) GetUserName() string {
	return i.name
}

// GetSessionId get sessionId
func (i *Info) GetSessionId() string {
	return i.sessionId
}

// GetPod returns user pod handler
func (i *Info) GetPod() *pod.Pod {
	return i.pod
}

// GetAccount returns user account info
func (i *Info) GetAccount() *account.Account {
	return i.account
}

// GetFeed returns user feed handler
func (i *Info) GetFeed() *feed.API {
	return i.feedApi
}

// AddPodName adds pod to user user pod map
func (i *Info) AddPodName(podName string, podInfo *pod.Info) {
	i.openPodsMu.Lock()
	defer i.openPodsMu.Unlock()
	i.openPods[podName] = podInfo
}

// RemovePodName removes pod from user pod map
func (i *Info) RemovePodName(podName string) {
	i.openPodsMu.Lock()
	defer i.openPodsMu.Unlock()
	delete(i.openPods, podName)
}

// IsPodOpen checks if users pod is open
func (i *Info) IsPodOpen(podName string) bool {
	i.openPodsMu.RLock()
	defer i.openPodsMu.RUnlock()
	if _, ok := i.openPods[podName]; ok {
		return true
	}
	return false
}

// GetUserDirectory returns user directory handler
func (i *Info) GetUserDirectory() *d.Directory {
	return i.dir
}
