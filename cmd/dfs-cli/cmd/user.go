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
	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"log"
	"net/http"
	"strings"
)

func userNew(userName string) {
	password := getPassword()
	newUser := common.UserRequest{
		UserName: userName,
		Password: password,
	}
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		fmt.Println("create user: error marshalling request")
		return
	}

	data, err := fdfsAPI.postReq(http.MethodPost, apiUserSignup, jsonData)
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
	fmt.Println("user created with address ", resp.Address)
	fmt.Println("Please store the following 12 words safely")
	fmt.Println("if you loose this, you cannot recover the data in-case of an emergency.")
	fmt.Println("you can also use this mnemonic to access the datain case this device is lost")
	fmt.Println("=============== Mnemonic ==========================")
	fmt.Println(resp.Mnemonic)
	fmt.Println("=============== Mnemonic ==========================")
}

func userImportUsingAddress(userName, address string) {
	password := getPassword()
	importUser := common.UserRequest{
		UserName: userName,
		Password: password,
		Address:  address,
	}
	jsonData, err := json.Marshal(importUser)
	if err != nil {
		log.Fatalf("import user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserImport, jsonData)
	if err != nil {
		fmt.Println("import user: ", err)
		return
	}
	var resp api.UserSignupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("import user: ", err)
		return
	}
	fmt.Println("imported user name: ", userName)
	fmt.Println("imported user address: ", resp.Address)
}

func userImportUsingMnemonic(userName, mnemonic string) {
	password := getPassword()
	importUser := common.UserRequest{
		UserName: userName,
		Password: password,
		Mnemonic:  mnemonic,
	}
	jsonData, err := json.Marshal(importUser)
	if err != nil {
		log.Fatalf("import user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserImport, jsonData)
	if err != nil {
		fmt.Println("import user: ", err)
		return
	}
	var resp api.UserSignupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("import user: ", err)
		return
	}
	fmt.Println("imported user name: ", userName)
	fmt.Println("imported user address: ", resp.Address)
}

func userLogin(userName string) {
	password := getPassword()
	loginUser := common.UserRequest{
		UserName: userName,
		Password: password,
	}
	jsonData, err := json.Marshal(loginUser)
	if err != nil {
		log.Fatalf("login user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserLogin, jsonData)
	if err != nil {
		fmt.Println("login user: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func deleteUser() {
	password := getPassword()
	delUser := common.UserRequest{
		Password: password,
	}
	jsonData, err := json.Marshal(delUser)
	if err != nil {
		fmt.Println("delete user: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiUserDelete, jsonData)
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

func exportUser() {
	data, err := fdfsAPI.postReq(http.MethodPost, apiUserExport, nil)
	if err != nil {
		fmt.Println("export user: ", err)
		return
	}
	var resp api.UserExportResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("export user: ", err)
		return
	}
	fmt.Println("user name:", resp.Name)
	fmt.Println("address  :", resp.Address)
}

func StatUser() {
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

func presentUser(userName string) {
	data, err := fdfsAPI.getReq(apiUserPresent, "user_name="+ userName)
	if err != nil {
		fmt.Println("user present: ", err)
		return
	}
	var resp api.UserPresentResponse
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
	data, err := fdfsAPI.getReq(apiUserIsLoggedin, "user_name="+ userName)
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
		fmt.Println("user is logged in")
	} else {
		fmt.Println("user is NOT logged in")
	}
}