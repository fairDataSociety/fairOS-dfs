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

import "github.com/fairdatasociety/fairOS-dfs/pkg/account"

type Stat struct {
	Name      string `json:"user_name"`
	Reference string `json:"reference"`
}

// GetUserStat shows the user information like user name and his address.
func (u *Users) GetUserStat(userInfo *Info) (*Stat, error) {
	if !u.IsUsernameAvailable(userInfo.name, u.dataDir) {
		return nil, ErrInvalidUserName
	}

	stat := &Stat{
		Name:      userInfo.name,
		Reference: userInfo.GetAccount().GetAddress(account.UserAccountIndex).Hex(),
	}
	return stat, nil
}
