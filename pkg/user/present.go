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

// IsUsernameAvailable checks if a supplied user name is present in this dfs server.
func (u *Users) IsUsernameAvailable(userName, dataDir string) bool {
	return u.isUserMappingPresent(userName, dataDir)
}

// IsUsernameAvailableV2 checks if a supplied user name is present in blockchain
func (u *Users) IsUsernameAvailableV2(userName string) bool {
	addr, err := u.ens.GetOwner(userName)
	if err != nil { // skipcq: TCV-001
		return false
	}
	return addr.Hex() != zeroAddressHex
}
