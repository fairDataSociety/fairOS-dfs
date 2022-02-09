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
	"os"
	"path/filepath"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/spf13/afero"
)

const (
	userDirectoryName = "user"
)

func (u *Users) isUserMappingPresent(userName, dataDir string) bool {
	destDir := filepath.Join(dataDir, userDirectoryName)
	err := u.os.MkdirAll(destDir, 0700)
	if err != nil {
		return false
	}
	userFileName := filepath.Join(destDir, userName)
	info, err := u.os.Stat(userFileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (u *Users) storeUserNameToAddressFileMapping(userName, dataDir string, address utils.Address) error {
	destDir := filepath.Join(dataDir, userDirectoryName)
	err := u.os.MkdirAll(destDir, 0700)
	if err != nil {
		return err
	}
	userFileName := filepath.Join(destDir, userName)
	return afero.WriteFile(u.os, userFileName, address.ToBytes(), 0700)
}

func (u *Users) deleteUserMapping(userName, dataDir string) error {
	destDir := filepath.Join(dataDir, userDirectoryName)
	userFileName := filepath.Join(destDir, userName)
	return u.os.Remove(userFileName)
}

func (u *Users) getAddressFromUserName(userName, dataDir string) (utils.Address, error) {
	destDir := filepath.Join(dataDir, userDirectoryName)
	userFileName := filepath.Join(destDir, userName)
	data, err := afero.ReadFile(u.os, userFileName)
	if err != nil {
		return utils.ZeroAddress, err
	}
	return utils.NewAddress(data), nil
}

func (u *Users) GetUserMap(dataDir string) (map[string]string, error) {
	users := map[string]string{}
	destDir := filepath.Join(dataDir, userDirectoryName)
	files, err := afero.ReadDir(u.os, destDir)
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		addr, err := u.getAddressFromUserName(v.Name(), dataDir)
		if err != nil {
			continue
		}
		users[v.Name()] = addr.Hex()
	}
	return users, nil
}

func (u *Users) LoadUserMap(dataDir string, users map[string]string) error {
	for i, v := range users {
		addr := utils.HexToAddress(v)
		err := u.storeUserNameToAddressFileMapping(i, dataDir, addr)
		if err != nil {
			return err
		}
	}
	return nil
}
