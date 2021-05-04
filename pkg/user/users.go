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
	sessionMap   map[string]string
	sessionMu    *sync.RWMutex
	gatewayMode  bool
	cookieDomain string
	logger       logging.Logger
}

func NewUsers(dataDir string, client blockstore.Client, cookieDomain string, gatewayMode bool, logger logging.Logger) *Users {
	return &Users{
		dataDir:      dataDir,
		client:       client,
		userMap:      make(map[string]*Info),
		userMu:       &sync.RWMutex{},
		sessionMap:   make(map[string]string),
		sessionMu:    &sync.RWMutex{},
		gatewayMode: gatewayMode,
		cookieDomain: cookieDomain,
		logger:       logger,
	}
}

func (u *Users) addUserToMap(info *Info) {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	u.userMap[info.userAddress] = info
}

func (u *Users) removeUserFromMap(userAddressString string) {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	delete(u.userMap, userAddressString)
}

func (u *Users) getUserFromMap(userAddressString string) *Info {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	return u.userMap[userAddressString]
}

func (u *Users) isUserPresentInMap(userAddressString string) bool {
	u.userMu.Lock()
	defer u.userMu.Unlock()
	if _, ok := u.userMap[userAddressString]; ok {
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

// session management
func (u *Users) addSessionrToMap(sessionId string) {
	u.sessionMu.Lock()
	defer u.sessionMu.Unlock()
	u.sessionMap[sessionId] = ""
}

func (u *Users) removeSessionFromMap(sessionId string) {
	u.sessionMu.Lock()
	defer u.sessionMu.Unlock()
	delete(u.sessionMap, sessionId)
}

func (u *Users) isSessionPresentInMap(sessionId string) bool {
	u.sessionMu.Lock()
	defer u.sessionMu.Unlock()
	if _, ok := u.sessionMap[sessionId]; ok {
		return true
	}
	return false
}