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

// ExportUser gives back the information required to export the user from one dfs server
// import him in to another.
func (u *Users) ExportUser(ui *Info) (string, string, error) {
	address, err := u.getAddressFromUserName(ui.name, u.dataDir)
	if err != nil {
		return "", "", err
	}
	return ui.name, address.Hex(), nil
}
