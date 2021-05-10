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

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

type Users struct {
	dataDir      string
	client       blockstore.Client
	userMap      map[string]*Info
	userMu       *sync.RWMutex
	cookieDomain string
	logger       logging.Logger
}

func NewUsers(dataDir string, client blockstore.Client, cookieDomain string, logger logging.Logger) *Users {
	return &Users{
		dataDir:      dataDir,
		client:       client,
		userMap:      make(map[string]*Info),
		userMu:       &sync.RWMutex{},
		cookieDomain: cookieDomain,
		logger:       logger,
	}
}

func (u *Users) addUserToMap(info *Info) {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	u.userMap[info.sessionId] = info
}

func (u *Users) removeUserFromMap(sessionId string) {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	delete(u.userMap, sessionId)
}

func (u *Users) getUserFromMap(sessionId string) *Info {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	return u.userMap[sessionId]
}

func (u *Users) isUserPresentInMap(sessionId string) bool {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	if _, ok := u.userMap[sessionId]; ok {
		return true
	}
	return false
}

func (u *Users) isUserNameInMap(userName string) bool {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	for _, ui := range u.userMap {
		if ui.name == userName {
			return true
		}
	}
	return false
}
