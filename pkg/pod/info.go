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

package pod

import (
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	di "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type Info struct {
	podName         string
	user            utils.Address
	dir             *di.Directory
	file            *f.File
	accountInfo     *account.Info
	feed            *feed.API
	currentPodInode *di.DirInode
	curPodMu        sync.RWMutex
	currentDirInode *di.DirInode
	curDirMu        sync.RWMutex
	collection      *collection.KeyValue
}

func (i *Info) GetDirectory() *di.Directory {
	return i.dir
}

func (i *Info) getFile() *f.File {
	return i.file
}

func (i *Info) GetUser() utils.Address {
	return i.user
}

func (i *Info) GetAccountInfo() *account.Info {
	return i.accountInfo
}

func (i *Info) GetFeed() *feed.API {
	return i.feed
}

func (i *Info) GetCurrentPodInode() *di.DirInode {
	return i.currentPodInode
}
func (i *Info) GetCurrentDirInode() *di.DirInode {
	return i.currentDirInode
}

func (i *Info) SetCurrentPodInode(podInode *di.DirInode) {
	i.currentPodInode = podInode
}
func (i *Info) SetCurrentDirInode(podInode *di.DirInode) {
	i.currentDirInode = podInode
}

func (p *Info) IsCurrentDirRoot() bool {
	if p.currentDirInode.Meta.Path == utils.PathSeperator {
		return true
	} else {
		return false
	}
}

func (i *Info) GetCurrentPodPathOnly() string {
	return i.currentPodInode.Meta.Path
}

func (i *Info) GetCurrentPodNameOnly() string {
	return i.currentPodInode.Meta.Name
}

func (i *Info) GetCurrentPodPathAndName() string {
	return i.currentPodInode.Meta.Path + i.currentPodInode.Meta.Name
}

func (i *Info) GetCurrentDirPathOnly() string {
	return i.currentDirInode.Meta.Path
}

func (i *Info) GetCurrentDirNameOnly() string {
	return i.currentDirInode.Meta.Name
}

func (i *Info) GetCurrentDirPathAndName() string {
	return i.currentDirInode.Meta.Path + utils.PathSeperator + i.currentDirInode.Meta.Name
}

func (i *Info) GetCollection() *collection.KeyValue {
	return i.collection
}
