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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	userDirectoryName = "user"
)

func (*Users) isUserMappingPresent(userName, dataDir string) bool {
	destDir := filepath.Join(dataDir, userDirectoryName)
	err := os.MkdirAll(destDir, 0700)
	if err != nil {
		return false
	}
	userFileName := filepath.Join(destDir, userName)
	info, err := os.Stat(userFileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (*Users) storeUserNameToAddressFileMapping(userName, dataDir string, address utils.Address) error {
	destDir := filepath.Join(dataDir, userDirectoryName)
	err := os.MkdirAll(destDir, 0700)
	if err != nil {
		return err
	}
	userFileName := filepath.Join(destDir, userName)
	return ioutil.WriteFile(userFileName, address.ToBytes(), 0700)
}

func (u *Users) deleteUserMapping(userName, dataDir string) error {
	destDir := filepath.Join(dataDir, userDirectoryName)
	userFileName := filepath.Join(destDir, userName)
	return os.Remove(userFileName)
}

func (*Users) getAddressFromUserName(userName, dataDir string) (utils.Address, error) {
	destDir := filepath.Join(dataDir, userDirectoryName)
	userFileName := filepath.Join(destDir, userName)
	data, err := ioutil.ReadFile(userFileName)
	if err != nil {
		return utils.ZeroAddress, err
	}
	return utils.NewAddress(data), nil
}
