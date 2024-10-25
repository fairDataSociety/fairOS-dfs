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

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func userNew(userName, mnemonic string) {
	password := getPassword()
	newUser := common.UserSignupRequest{
		UserName: userName,
		Password: password,
		Mnemonic: mnemonic,
	}
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		fmt.Println("create user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserSignupV2, jsonData)
	if err != nil {
		fmt.Println("create user: ", err)
		return
	}

	var resp api.UserSignupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("create user: ", err)
		return
	}
	if resp.Message == eth.ErrInsufficientBalance.Error() {
		fmt.Println("Failed to create new user")
		fmt.Println("Please fund your account with some eth and try again with the following command. This is not related with bee wallet")
		fmt.Printf(">>> user new %s %s\n", userName, resp.Mnemonic)
		fmt.Println("address :", resp.Address)
		fmt.Println("=============== Mnemonic ==========================")
		fmt.Println(resp.Mnemonic)
		fmt.Println("=============== Mnemonic ==========================")
		return
	}
	fmt.Println("user created with address ", resp.Address)
	fmt.Println("Please store the 12 words mnemonic safely")
	fmt.Println("if you loose that, you cannot recover the data in-case of an emergency.")
	fmt.Println("you can also use that mnemonic to access the data in-case this device is lost")
	fmt.Println("=============== Mnemonic Start ==========================")
	fmt.Println(resp.Mnemonic)
	fmt.Println("=============== Mnemonic End ==========================")
	fmt.Println("=============== PublicKey Start ==========================")
	fmt.Println(resp.PublicKey)
	fmt.Println("=============== PublicKey End ==========================")
	fdfsAPI.setAccessToken(resp.AccessToken)

	currentUser = userName
}

func userLogin(userName, apiEndpoint string) {
	password := getPassword()
	loginUser := common.UserSignupRequest{
		UserName: userName,
		Password: password,
	}
	jsonData, err := json.Marshal(loginUser)
	if err != nil {
		fmt.Println("login user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiEndpoint, jsonData)
	if err != nil {
		fmt.Println("login user: ", err)
		return
	}
	var resp api.UserSignupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("create user: ", err)
		return
	}
	fmt.Println("=============== PublicKey Start ==========================")
	fmt.Println(resp.PublicKey)
	fmt.Println("=============== PublicKey End ==========================")
	currentUser = userName
	message := strings.ReplaceAll(string(data), "\n", "")
	fdfsAPI.setAccessToken(resp.AccessToken)

	fmt.Println(message)
}

func signatureLogin(signature, apiEndpoint string) {
	password := getPassword()
	loginUser := common.UserSignatureLoginRequest{
		Signature: signature,
		Password:  password,
	}
	jsonData, err := json.Marshal(loginUser)
	if err != nil {
		fmt.Println("login user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiEndpoint, jsonData)
	if err != nil {
		fmt.Println("login user: ", err)
		return
	}
	var resp api.UserSignupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("create user: ", err)
		return
	}

	currentUser = resp.Address
	message := strings.ReplaceAll(string(data), "\n", "")
	fdfsAPI.setAccessToken(resp.AccessToken)

	fmt.Println(message)
}

func deleteUser(apiEndpoint string) {
	password := getPassword()
	delUser := common.UserSignupRequest{
		Password: password,
	}
	jsonData, err := json.Marshal(delUser)
	if err != nil {
		fmt.Println("delete user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiEndpoint, jsonData)
	if err != nil {
		fmt.Println("delete user: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func logoutUser() {
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserLogout, nil)
	if err != nil {
		fmt.Println("logout user: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func statUser() {
	data, err := fdfsAPI.getReq(apiUserStat, "")
	if err != nil {
		fmt.Println("user stat: ", err)
		return
	}
	var resp user.Stat
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("user stat: ", err)
		return
	}
	fmt.Println("user name: ", resp.Name)
	fmt.Println("Reference: ", resp.Reference)
}

func presentUser(userName, apiEndpoint string) {
	data, err := fdfsAPI.getReq(apiEndpoint, "userName="+userName)
	if err != nil {
		fmt.Println("user present: ", err)
		return
	}
	var resp api.PresentResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("import user: ", err)
		return
	}
	if resp.Present {
		fmt.Println("User is present")
	} else {
		fmt.Println("User is not present")
	}
}

func isUserLoggedIn(userName string) {
	data, err := fdfsAPI.getReq(apiUserIsLoggedin, "userName="+userName)
	if err != nil {
		fmt.Println("user loggedin: ", err)
		return
	}
	var resp api.LoginStatus
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("user loggedin: ", err)
		return
	}
	if resp.LoggedIn {
		fmt.Println("user is logged-in")
	} else {
		fmt.Println("user is NOT logged-in")
	}
}
