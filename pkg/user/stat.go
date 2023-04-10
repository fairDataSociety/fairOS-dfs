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

// Stat is the structure of the user information.
type Stat struct {
	Name      string `json:"userName"`
	Reference string `json:"address"`
}

// GetUserStat shows the user information like username and his address.
func (u *Users) GetUserStat(userInfo *Info) (*Stat, error) {
	if !u.IsUsernameAvailableV2(userInfo.name) {
		return nil, ErrInvalidUserName
	}

	stat := &Stat{
		Name:      userInfo.name,
		Reference: userInfo.GetAccount().GetAddress(account.UserAccountIndex).Hex(),
	}
	return stat, nil
}
