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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/tinygrasshopper/bettercsv"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	DefaultPrompt   = "dfs"
	UserSeperator   = ">>>"
	PodSeperator    = ">>"
	PromptSeperator = "> "
	APIVersion      = "/v0"
)

var (
	currentUser      string
	currentPod       string
	currentPrompt    string
	currentDirectory string
	fdfsAPI          *FdfsClient
)

const (
	apiUserSignup      = APIVersion + "/user/signup"
	apiUserLogin       = APIVersion + "/user/login"
	apiUserImport      = APIVersion + "/user/import"
	apiUserPresent     = APIVersion + "/user/present"
	apiUserIsLoggedin  = APIVersion + "/user/isloggedin"
	apiUserLogout      = APIVersion + "/user/logout"
	apiUserAvatar      = APIVersion + "/user/avatar"
	apiUserName        = APIVersion + "/user/name"
	apiUserContact     = APIVersion + "/user/contact"
	apiUserExport      = APIVersion + "/user/export"
	apiUserDelete      = APIVersion + "/user/delete"
	apiUserStat        = APIVersion + "/user/stat"
	apiUserShareInbox  = APIVersion + "/user/share/inbox"
	apiUserShareOutbox = APIVersion + "/user/share/outbox"
	apiPodNew          = APIVersion + "/pod/new"
	apiPodOpen         = APIVersion + "/pod/open"
	apiPodClose        = APIVersion + "/pod/close"
	apiPodSync         = APIVersion + "/pod/sync"
	apiPodDelete       = APIVersion + "/pod/delete"
	apiPodLs           = APIVersion + "/pod/ls"
	apiPodStat         = APIVersion + "/pod/stat"
	apiPodShare        = APIVersion + "/pod/share"
	apiPodReceive      = APIVersion + "/pod/receive"
	apiPodReceiveInfo  = APIVersion + "/pod/receiveinfo"
	apiDirIsPresent    = APIVersion + "/dir/present"
	apiDirMkdir        = APIVersion + "/dir/mkdir"
	apiDirRmdir        = APIVersion + "/dir/rmdir"
	apiDirLs           = APIVersion + "/dir/ls"
	apiDirStat         = APIVersion + "/dir/stat"
	apiFileDownload    = APIVersion + "/file/download"
	apiFileUpload      = APIVersion + "/file/upload"
	apiFileShare       = APIVersion + "/file/share"
	apiFileReceive     = APIVersion + "/file/receive"
	apiFileReceiveInfo = APIVersion + "/file/receiveinfo"
	apiFileDelete      = APIVersion + "/file/delete"
	apiFileStat        = APIVersion + "/file/stat"
	apiKVCreate        = APIVersion + "/kv/new"
	apiKVList          = APIVersion + "/kv/ls"
	apiKVOpen          = APIVersion + "/kv/open"
	apiKVDelete        = APIVersion + "/kv/delete"
	apiKVCount         = APIVersion + "/kv/count"
	apiKVEntryPut      = APIVersion + "/kv/entry/put"
	apiKVEntryGet      = APIVersion + "/kv/entry/get"
	apiKVEntryDelete   = APIVersion + "/kv/entry/del"
	apiKVLoadCSV       = APIVersion + "/kv/loadcsv"
	apiKVSeek          = APIVersion + "/kv/seek"
	apiKVSeekNext      = APIVersion + "/kv/seek/next"
	apiDocCreate       = APIVersion + "/doc/new"
	apiDocList         = APIVersion + "/doc/ls"
	apiDocOpen         = APIVersion + "/doc/open"
	apiDocCount        = APIVersion + "/doc/count"
	apiDocDelete       = APIVersion + "/doc/delete"
	apiDocFind         = APIVersion + "/doc/find"
	apiDocEntryPut     = APIVersion + "/doc/entry/put"
	apiDocEntryGet     = APIVersion + "/doc/entry/get"
	apiDocEntryDel     = APIVersion + "/doc/entry/del"
	apiDocLoadJson     = APIVersion + "/doc/loadjson"
)

type Message struct {
	Message string
	Code    int
}

func NewPrompt() {
	var err error
	fdfsAPI, err = NewFdfsClient(fdfsHost, fdfsPort)
	if err != nil {
		fmt.Println("could not create fdfs client")
		os.Exit(1)
	}
	if !fdfsAPI.CheckConnection() {
		fmt.Println("could not connect to fdfs server")
		os.Exit(2)
	}
}

func initPrompt() {
	currentPrompt = DefaultPrompt + " " + UserSeperator
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(currentPrompt),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("dfs"),
	)
	p.Run()
}

func changeLivePrefix() (string, bool) {
	return currentPrompt, true
}

var suggestions = []prompt.Suggest{
	{Text: "user new", Description: "create a new user"},
	{Text: "user del", Description: "delete a existing user"},
	{Text: "user login", Description: "login to a existing user"},
	{Text: "user logout", Description: "logout from a logged in user"},
	{Text: "user present", Description: "is user present"},
	{Text: "user ls", Description: "list all users"},
	{Text: "user name", Description: "sets and gets the user name information"},
	{Text: "user contact", Description: "sets and gets the user contact information"},
	{Text: "user share inbox", Description: "gets the information about the files received by the user"},
	{Text: "user share outbox", Description: "gets the information about the files shared by the user"},
	{Text: "user export ", Description: "exports the user"},
	{Text: "user import ", Description: "imports the user"},
	{Text: "user stat ", Description: "shows information about a user"},
	{Text: "pod new", Description: "create a new pod for a user"},
	{Text: "pod del", Description: "delete a existing pod of a user"},
	{Text: "pod open", Description: "open to a existing pod of a user"},
	{Text: "pod close", Description: "close a already opened pod of a user"},
	{Text: "pod ls", Description: "list all the existing pods of  auser"},
	{Text: "pod stat", Description: "show the metadata of a pod of a user"},
	{Text: "pod sync", Description: "sync the pod from swarm"},
	{Text: "kv new", Description: "create new key value store"},
	{Text: "kv delete", Description: "delete the  key value store"},
	{Text: "kv ls", Description: "lists all the key value stores"},
	{Text: "kv open", Description: "open already created key value store"},
	{Text: "kv get", Description: "get value from key"},
	{Text: "kv put", Description: "put key and value in kv store"},
	{Text: "kv del", Description: "delete key and value from the store"},
	{Text: "kv loadcsv", Description: "loads the csv file in to kv store"},
	{Text: "kv seek", Description: "seek to the given start prefix"},
	{Text: "kv getnext", Description: "get the next element"},
	{Text: "doc new", Description: "creates a new document store"},
	{Text: "doc delete", Description: "deletes a document store"},
	{Text: "doc open", Description: "open the document store"},
	{Text: "doc ls", Description: "list all document dbs"},
	{Text: "doc count", Description: "count the docs in the table satisfying the expression"},
	{Text: "doc find", Description: "find the docs in the table satisfying the expression and limit"},
	{Text: "doc put", Description: "insert a json document in to document store"},
	{Text: "doc get", Description: "get the document having the id from the store"},
	{Text: "doc del", Description: "delete the document having the id from the store"},
	{Text: "doc loadjson", Description: "load the json file in to the newly created document db"},
	{Text: "cd", Description: "change path"},
	{Text: "copyToLocal", Description: "copy file from dfs to local machine"},
	{Text: "copyFromLocal", Description: "copy file from local machine to dfs"},
	{Text: "share", Description: "share file with another user"},
	{Text: "receive", Description: "receive a shared file"},
	{Text: "exit", Description: "exit dfs-prompt"},
	{Text: "head", Description: "show few starting lines of a file"},
	{Text: "help", Description: "show usage"},
	{Text: "ls", Description: "list all the file and directories in the current path"},
	{Text: "mkdir", Description: "make a new directory"},
	{Text: "rmdir", Description: "remove a existing directory"},
	{Text: "pwd", Description: "show the current working directory"},
	{Text: "rm", Description: "remove a file"},
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

func executor(in string) {
	in = strings.TrimSpace(in)
	blocks := strings.Split(in, " ")
	switch blocks[0] {
	case "help":
		help()
	case "exit":
		os.Exit(0)
	case "user":
		if len(blocks) < 2 {
			log.Println("invalid command.")
			help()
			return
		}
		switch blocks[1] {
		case "new":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			args := make(map[string]string)
			args["user"] = userName
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserSignup, args)
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
			currentUser = userName
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "import":
			if len(blocks) == 4 {
				userName := blocks[2]
				address := blocks[3]
				args := make(map[string]string)
				args["user"] = userName
				args["address"] = address
				args["password"] = getPassword()
				data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserImport, args)
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
				currentUser = userName
				currentPod = ""
				currentDirectory = ""
				currentPrompt = getCurrentPrompt()
				return
			}
			if len(blocks) > 4 && len(blocks) < 15 {
				fmt.Println("invalid command. Missing arguments")
				return
			}
			userName := blocks[2]
			var mnemonic string
			for i := 3; i < 15; i++ {
				mnemonic = mnemonic + " " + blocks[i]
			}
			mnemonic = strings.TrimPrefix(mnemonic, " ")
			args := make(map[string]string)
			args["user"] = userName
			args["password"] = getPassword()
			args["mnemonic"] = mnemonic
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserImport, args)
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
			currentUser = userName
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "login":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			args := make(map[string]string)
			args["user"] = userName
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserLogin, args)
			if err != nil {
				fmt.Println("login user: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentUser = userName
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "present":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			args := make(map[string]string)
			args["user"] = userName
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserPresent, args)
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
			currentPrompt = getCurrentPrompt()
		case "ls":
			//users, err := dfsAPI.ListAllUsers()
			//if err != nil {
			//	fmt.Println("user ls: ", err)
			//	return
			//}
			//for _, usr := range users {
			//	fmt.Println(usr)
			//}
			currentPrompt = getCurrentPrompt()
		case "del":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			args := make(map[string]string)
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiUserDelete, args)
			if err != nil {
				fmt.Println("delete user: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentUser = ""
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "logout":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserLogout, nil)
			if err != nil {
				fmt.Println("logout user: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentUser = ""
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "export":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserExport, nil)
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
			currentPrompt = getCurrentPrompt()
		case "name":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			if len(blocks) == 6 {
				firstName := blocks[2]
				middleName := blocks[3]
				lastName := blocks[4]
				surNmae := blocks[5]
				args := make(map[string]string)
				args["first_name"] = firstName
				args["last_name"] = lastName
				args["middle_name"] = middleName
				args["surname"] = surNmae
				_, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserName, args)
				if err != nil {
					fmt.Println("name: ", err)
					return
				}
			} else if len(blocks) == 2 {
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserName, nil)
				if err != nil {
					fmt.Println("name: ", err)
					return
				}
				var resp user.Name
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("namer: ", err)
					return
				}
				fmt.Println("first_name : ", resp.FirstName)
				fmt.Println("middle_name: ", resp.MiddleName)
				fmt.Println("last_name  : ", resp.LastName)
				fmt.Println("surname    : ", resp.SurName)
			}
			currentPrompt = getCurrentPrompt()
		case "contact":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			if len(blocks) == 8 {
				phone := blocks[2]
				mobile := blocks[3]
				addressLine1 := blocks[4]
				addressLine2 := blocks[5]
				state := blocks[6]
				zip := blocks[7]
				args := make(map[string]string)
				args["phone"] = phone
				args["mobile"] = mobile
				args["address_line_1"] = addressLine1
				args["address_line_2"] = addressLine2
				args["state_province_region"] = state
				args["zipcode"] = zip
				_, err := fdfsAPI.callFdfsApi(http.MethodPost, apiUserContact, args)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
			} else if len(blocks) == 2 {
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserContact, nil)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
				var resp user.Contacts
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
				fmt.Println("phone        : ", resp.Phone)
				fmt.Println("mobile       : ", resp.Mobile)
				fmt.Println("address_line1: ", resp.Addr.AddressLine1)
				fmt.Println("address_line2: ", resp.Addr.AddressLine2)
				fmt.Println("state        : ", resp.Addr.State)
				fmt.Println("zipcode      : ", resp.Addr.ZipCode)
			}
			currentPrompt = getCurrentPrompt()
		case "share":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"inbox/outbox\" argument ")
				return
			}
			switch blocks[2] {
			case "inbox":
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserShareInbox, nil)
				if err != nil {
					fmt.Println("sharing inbox: ", err)
					return
				}
				var resp user.Inbox
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("sharing inbox: ", err)
					return
				}
				for _, entry := range resp.Entries {
					fmt.Println(entry)
				}
				currentPrompt = getCurrentPrompt()
			case "outbox":
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserShareOutbox, nil)
				if err != nil {
					fmt.Println("sharing outbox: ", err)
					return
				}
				var resp user.Outbox
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("sharing outbox: ", err)
					return
				}
				for _, entry := range resp.Entries {
					fmt.Println(entry)
				}
			}
			currentPrompt = getCurrentPrompt()
		case "loggedin":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			args := make(map[string]string)
			args["user"] = userName
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserIsLoggedin, args)
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
			currentPrompt = getCurrentPrompt()
		case "stat":
			if currentUser == "" {
				fmt.Println("please login as user to do the operation")
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiUserStat, nil)
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
			currentPrompt = getCurrentPrompt()
		case "avatar":
			if currentUser == "" {
				fmt.Println("please login as user to do the operation")
				return
			}
			if len(blocks) < 3 {
				// get avatar
				// Create the temp file
				tmpDir := os.TempDir()
				fd, err := ioutil.TempFile(tmpDir, "avatar")
				if err != nil {
					log.Fatal(err)
				}
				defer fd.Close()

				_, err = fdfsAPI.downloadMultipartFile(http.MethodGet, apiUserAvatar, nil, fd)
				if err != nil {
					fmt.Println("avatar download failed: ", err)
					return
				}
				fmt.Println("Avatar downloaded in ", fd.Name())
			} else {
				// put avatar
				fileName := filepath.Base(blocks[2])
				fd, err := os.Open(blocks[2])
				if err != nil {
					fmt.Println("avatar file open failed: ", err)
					return
				}
				fi, err := fd.Stat()
				if err != nil {
					fmt.Println("avatar file stat failed: ", err)
					return
				}
				data, err := fdfsAPI.uploadMultipartFile(apiUserAvatar, fileName, fi.Size(), fd, nil, "avatar", "false")
				if err != nil {
					fmt.Println("upload failed: ", err, string(data))
					return
				}
				message := strings.ReplaceAll(string(data), "\n", "")
				fmt.Println(message)
			}
			currentPrompt = getCurrentPrompt()
		default:
			fmt.Println("invalid user command")
		}
	case "pod":
		if currentUser == "" {
			fmt.Println("login as a user to execute these commands")
			return
		}
		if len(blocks) < 2 {
			log.Println("invalid command.")
			help()
			return
		}
		switch blocks[1] {
		case "new":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			args := make(map[string]string)
			args["pod"] = podName
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodNew, args)
			if err != nil {
				fmt.Println("could not create pod: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPod = podName
			currentDirectory = utils.PathSeperator
			currentPrompt = getCurrentPrompt()
		case "del":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			args := make(map[string]string)
			args["pod"] = podName
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiPodDelete, args)
			if err != nil {
				fmt.Println("could not delete pod: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "open":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			args := make(map[string]string)
			args["pod"] = podName

			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiPodLs, nil)
			if err != nil {
				fmt.Println("error while listing pods: %w", err)
				return
			}
			var resp api.PodListResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("pod stat: ", err)
				return
			}
			invalidPodName := true
			for _, pod := range resp.Pods {
				if pod == podName {
					args["password"] = getPassword()
					invalidPodName = false
				}
			}
			for _, pod := range resp.SharedPods {
				if pod == podName {
					args["password"] = ""
					invalidPodName = false
				}
			}
			if invalidPodName {
				fmt.Println("invalid pod name")
				break
			}

			data, err = fdfsAPI.callFdfsApi(http.MethodPost, apiPodOpen, args)
			if err != nil {
				fmt.Println("pod open failed: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPod = podName
			currentDirectory = utils.PathSeperator
			currentPrompt = getCurrentPrompt()
		case "close":
			if !isPodOpened() {
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodClose, nil)
			if err != nil {
				fmt.Println("error logging out: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "stat":
			if !isPodOpened() {
				return
			}
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			args := make(map[string]string)
			args["pod"] = podName
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiPodStat, args)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			var resp api.PodStatResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("pod stat: ", err)
				return
			}

			crTime, err := strconv.ParseInt(resp.CreationTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			accTime, err := strconv.ParseInt(resp.AccessTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			modTime, err := strconv.ParseInt(resp.ModificationTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			fmt.Println("Version          : ", resp.Version)
			fmt.Println("pod Name         : ", resp.PodName)
			fmt.Println("Path             : ", resp.PodPath)
			fmt.Println("Creation Time    :", time.Unix(crTime, 0).String())
			fmt.Println("Access Time      :", time.Unix(accTime, 0).String())
			fmt.Println("Modification Time:", time.Unix(modTime, 0).String())
			currentPrompt = getCurrentPrompt()
		case "sync":
			if !isPodOpened() {
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodSync, nil)
			if err != nil {
				fmt.Println("could not sync pod: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "ls":
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiPodLs, nil)
			if err != nil {
				fmt.Println("error while listing pods: %w", err)
				return
			}
			var resp api.PodListResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("pod stat: ", err)
				return
			}
			for _, pod := range resp.Pods {
				fmt.Println("<Pod>: ", pod)
			}
			for _, pod := range resp.SharedPods {
				fmt.Println("<Shared Pod>: ", pod)
			}
			currentPrompt = getCurrentPrompt()
		case "share":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			args := make(map[string]string)
			args["pod"] = podName
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodShare, args)
			if err != nil {
				fmt.Println("pod share failed: ", err)
				return
			}
			var sharingRef api.PodSharingReference
			err = json.Unmarshal(data, &sharingRef)
			if err != nil {
				fmt.Println("pod share failed: ", err)
				return
			}
			fmt.Println("Pod Sharing Reference : ", sharingRef.Reference)
			currentPrompt = getCurrentPrompt()
		case "receive":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podSharingReference := blocks[2]
			args := make(map[string]string)
			args["ref"] = podSharingReference
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodReceive, args)
			if err != nil {
				fmt.Println("pod receive failed: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "receiveinfo":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podSharingReference := blocks[2]
			args := make(map[string]string)
			args["ref"] = podSharingReference
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiPodReceiveInfo, args)
			if err != nil {
				fmt.Println("pod receive info failed: ", err)
				return
			}
			var podSharingInfo pod.ShareInfo
			err = json.Unmarshal(data, &podSharingInfo)
			if err != nil {
				fmt.Println("pod receive info failed: ", err)
				return
			}
			fmt.Println("Pod Name  : ", podSharingInfo.PodName)
			fmt.Println("Pod Ref.  : ", podSharingInfo.Address)
			fmt.Println("User Name : ", podSharingInfo.UserName)
			fmt.Println("User Ref. : ", podSharingInfo.UserAddress)
			fmt.Println("Shared Time : ", podSharingInfo.SharedTime)
			currentPrompt = getCurrentPrompt()

		default:
			fmt.Println("invalid pod command!!")
			help()
		} // end of pod commands
	case "kv":
		if currentUser == "" {
			fmt.Println("login as a user to execute these commands")
			return
		}
		if len(blocks) < 2 {
			log.Println("invalid command.")
			help()
			return
		}
		if !isPodOpened() {
			return
		}
		switch blocks[1] {
		case "new":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiKVCreate, args)
			if err != nil {
				fmt.Println("kv new: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "delete":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiKVDelete, args)
			if err != nil {
				fmt.Println("kv new: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "ls":
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiKVList, nil)
			if err != nil {
				fmt.Println("kv new: ", err)
				return
			}
			var resp api.Collections
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("kv ls: ", err)
				return
			}
			for _, table := range resp.Tables {
				fmt.Println("<KV>: ", table.Name)
			}
			currentPrompt = getCurrentPrompt()
		case "open":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiKVOpen, args)
			if err != nil {
				fmt.Println("kv open: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "count":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiKVCount, args)
			if err != nil {
				fmt.Println("kv open: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "put":
			if len(blocks) < 5 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			key := blocks[3]
			value := blocks[4]
			args := make(map[string]string)
			args["name"] = tableName
			args["key"] = key
			args["value"] = value
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiKVEntryPut, args)
			if err != nil {
				fmt.Println("kv put: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "get":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			key := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["key"] = key
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiKVEntryGet, args)
			if err != nil {
				fmt.Println("kv get: ", err)
				return
			}
			var resp api.KVResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("kv get: ", err)
				return
			}

			rdr := bytes.NewReader(resp.Values)
			csvReader := bettercsv.NewReader(rdr)
			csvReader.Comma = ','
			csvReader.Quote = '"'
			content, err := csvReader.ReadAll()
			if err != nil {
				fmt.Println("kv get: ", err)
				return
			}
			values := content[0]
			for i, name := range resp.Names {
				fmt.Println(name + " : " + values[i])
			}
			currentPrompt = getCurrentPrompt()
		case "del":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			key := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["key"] = key
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiKVEntryDelete, args)
			if err != nil {
				fmt.Println("kv del: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "loadcsv":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			fileName := filepath.Base(blocks[3])
			localCsvFile := blocks[3]

			fd, err := os.Open(localCsvFile)
			if err != nil {
				fmt.Println("loadcsv failed: ", err)
				return
			}
			fi, err := fd.Stat()
			if err != nil {
				fmt.Println("loadcsv failed: ", err)
				return
			}

			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.uploadMultipartFile(apiKVLoadCSV, fileName, fi.Size(), fd, args, "csv", "false")
			if err != nil {
				fmt.Println("loadcsv: ", err)
				return
			}
			var resp api.UploadFileResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("loadcsv: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "seek":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]

			var start string
			var end string
			var limit string
			if len(blocks) >= 4 {
				start = blocks[3]
			}
			if len(blocks) >= 5 {
				end = blocks[4]
			}

			if len(blocks) >= 6 {
				limit = blocks[5]
			}

			args := make(map[string]string)
			args["name"] = tableName
			args["start"] = start
			args["end"] = end
			args["limit"] = limit
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiKVSeek, args)
			if err != nil {
				fmt.Println("kv seek: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "getnext":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiKVSeekNext, args)
			if err != nil && !errors.Is(err, collection.ErrNoNextElement) {
				fmt.Println("kv get_next: ", err)
				return
			}

			if errors.Is(err, collection.ErrNoNextElement) {
				fmt.Println("no next element")
			} else {
				var resp api.KVResponse
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("kv get_next: ", err)
					return
				}

				rdr := bytes.NewReader(resp.Values)
				csvReader := bettercsv.NewReader(rdr)
				csvReader.Comma = ','
				csvReader.Quote = '"'
				content, err := csvReader.ReadAll()
				if err != nil {
					fmt.Println("kv get_next: ", err)
					return
				}
				values := content[0]
				for i, name := range resp.Names {
					fmt.Println(name + " : " + values[i])
				}
			}
			currentPrompt = getCurrentPrompt()
		default:
			fmt.Println("invalid kv command!!")
			help()
		}

	case "doc":
		if currentUser == "" {
			fmt.Println("login as a user to execute these commands")
			return
		}
		if len(blocks) < 2 {
			log.Println("invalid command.")
			help()
			return
		}
		if !isPodOpened() {
			return
		}
		switch blocks[1] {
		case "new":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			if len(blocks) == 4 {
				si := blocks[3]
				args["si"] = si
			}

			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiDocCreate, args)
			if err != nil {
				fmt.Println("doc new: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "ls":
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDocList, nil)
			if err != nil {
				fmt.Println("doc ls: ", err)
				return
			}
			var resp api.DocumentDBs
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("doc ls: ", err)
				return
			}
			for _, table := range resp.Tables {
				fmt.Println("<DOC>: ", table.Name)
				for fn, ft := range table.IndexedColumns {
					fmt.Println("     SI:", fn, ft)
				}
			}
			currentPrompt = getCurrentPrompt()
		case "open":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiDocOpen, args)
			if err != nil {
				fmt.Println("doc open: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "count":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]

			args := make(map[string]string)
			args["name"] = tableName
			if len(blocks) == 4 {
				args["expr"] = blocks[3]
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiDocCount, args)
			if err != nil {
				fmt.Println("doc count: ", err)
				return
			}
			count, err := strconv.ParseInt(string(data), 10, 64)
			if err != nil {
				fmt.Println("doc count: ", err)
				return
			}
			fmt.Println("Count = ", count)
			currentPrompt = getCurrentPrompt()
		case "delete":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiDocDelete, args)
			if err != nil {
				fmt.Println("doc del: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "find":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			expr := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["expr"] = expr
			if len(blocks) == 5 {
				args["limit"] = blocks[4]
			} else {
				args["limit"] = "10"
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDocFind, args)
			if err != nil {
				fmt.Println("doc find: ", err)
				return
			}
			var docs api.DocFindResponse
			err = json.Unmarshal(data, &docs)
			if err != nil {
				fmt.Println("doc find: ", err)
				return
			}
			for i, doc := range docs.Docs {
				fmt.Println("--- doc ", i)
				var d map[string]interface{}
				err = json.Unmarshal(doc, &d)
				if err != nil {
					fmt.Println("doc find: ", err)
					return
				}
				for k, v := range d {
					var valStr string
					switch val := v.(type) {
					case string:
						fmt.Println(k, "=", val)
					case float64:
						valStr = strconv.FormatFloat(val, 'E', -1, 10)
						fmt.Println(k, "=", valStr)
					case map[string]interface{}:
						for k1, v1 := range val {
							switch val1 := v1.(type) {
							case string:
								fmt.Println("   "+k1+" = ", val1)
							case float64:
								val2 := int64(val1)
								valStr = strconv.FormatInt(val2, 10)
								fmt.Println("   "+k1+" = ", valStr)
							default:
								fmt.Println("   "+k1+" = ", val1)
							}
						}
					default:
						fmt.Println(k, "=", val)
					}
				}

			}
			currentPrompt = getCurrentPrompt()
		case "put":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			value := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["doc"] = value
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiDocEntryPut, args)
			if err != nil {
				fmt.Println("doc put: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "get":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			idValue := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["id"] = idValue
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDocEntryGet, args)
			if err != nil {
				fmt.Println("doc get: ", err)
				return
			}

			var doc api.DocGetResponse
			err = json.Unmarshal(data, &doc)
			if err != nil {
				fmt.Println("doc get: ", err)
				return
			}
			var d map[string]interface{}
			err = json.Unmarshal(doc.Doc, &d)
			if err != nil {
				fmt.Println("doc get: ", err)
				return
			}
			for k, v := range d {
				fmt.Println(k, "=", v)
			}
			currentPrompt = getCurrentPrompt()
		case "del":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			idValue := blocks[3]
			args := make(map[string]string)
			args["name"] = tableName
			args["id"] = idValue
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiDocEntryDel, args)
			if err != nil {
				fmt.Println("doc del: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		case "loadjson":
			if len(blocks) < 4 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			tableName := blocks[2]
			fileName := filepath.Base(blocks[3])
			localCsvFile := blocks[3]

			fd, err := os.Open(localCsvFile)
			if err != nil {
				fmt.Println("loadjson failed: ", err)
				return
			}
			fi, err := fd.Stat()
			if err != nil {
				fmt.Println("loadjson failed: ", err)
				return
			}

			args := make(map[string]string)
			args["name"] = tableName
			data, err := fdfsAPI.uploadMultipartFile(apiDocLoadJson, fileName, fi.Size(), fd, args, "json", "false")
			if err != nil {
				fmt.Println("loadjson: ", err)
				return
			}
			var resp api.UploadFileResponse
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("loadjson: ", err)
				return
			}
			message := strings.ReplaceAll(string(data), "\n", "")
			fmt.Println(message)
			currentPrompt = getCurrentPrompt()
		default:
			fmt.Println("Invalid doc coammand")
			currentPrompt = getCurrentPrompt()
		}

	case "cd":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		dirTocd := blocks[1]

		// if cd'ing to previous dir, just do it
		if dirTocd == ".." && currentDirectory != utils.PathSeperator {
			currentDirectory = filepath.Dir(currentDirectory)
			currentPrompt = getCurrentPrompt()
			return
		}

		// if cd'ing to root dir, just do it
		if dirTocd == utils.PathSeperator {
			currentDirectory = utils.PathSeperator
			currentPrompt = getCurrentPrompt()
			return
		}

		// if cd'ing forward, we have to check if that dir is present
		if dirTocd != utils.PathSeperator {
			if currentDirectory == utils.PathSeperator {
				dirTocd = currentDirectory + dirTocd
			} else {
				dirTocd = currentDirectory + utils.PathSeperator + dirTocd
			}
		}

		args := make(map[string]string)
		args["dir"] = dirTocd
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDirIsPresent, args)
		if err != nil {
			fmt.Println("cd failed: ", err)
			return
		}
		var resp api.DirPresentResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("dir cd: ", err)
			return
		}
		if resp.Present {
			currentDirectory = dirTocd
		} else {
			fmt.Println("dir is not present: ", resp.Error)
		}
		currentPrompt = getCurrentPrompt()
	case "ls":
		if !isPodOpened() {
			return
		}
		args := make(map[string]string)
		args["dir"] = currentDirectory
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDirLs, args)
		if err != nil {
			fmt.Println("ls failed: ", err)
			return
		}
		var resp api.ListFileResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("dir ls: ", err)
			return
		}
		for _, entry := range resp.Entries {
			if entry.ContentType == "inode/directory" {
				fmt.Println("<Dir>: ", entry.Name)
			} else {
				fmt.Println("<File>: ", entry.Name)
			}
		}
		currentPrompt = getCurrentPrompt()
	case "mkdir":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		dirToMk := blocks[1]
		if dirToMk == "" {
			fmt.Println("invalid dir")
			return
		}

		if !strings.HasPrefix(dirToMk, utils.PathSeperator) {
			// then this path is not from root
			dirToMk = currentDirectory + utils.PathSeperator + dirToMk
		}

		args := make(map[string]string)
		args["dir"] = dirToMk

		data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiDirMkdir, args)
		if err != nil {
			fmt.Println("mkdir failed: ", err)
			return
		}
		message := strings.ReplaceAll(string(data), "\n", "")
		fmt.Println(message)
		currentPrompt = getCurrentPrompt()
	case "rmdir":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		dirToRm := blocks[1]
		if dirToRm == "" {
			fmt.Println("invalid dir")
			return
		}
		if !strings.HasPrefix(dirToRm, utils.PathSeperator) {
			// then this path is not from root
			if currentDirectory == utils.PathSeperator {
				dirToRm = currentDirectory + dirToRm
			} else {
				dirToRm = currentDirectory + utils.PathSeperator + dirToRm
			}
		}

		args := make(map[string]string)
		args["dir"] = dirToRm
		data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiDirRmdir, args)
		if err != nil {
			fmt.Println("rmdir failed: ", err)
			return
		}
		message := strings.ReplaceAll(string(data), "\n", "")
		fmt.Println(message)
		currentPrompt = getCurrentPrompt()
	case "upload":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 5 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		fileName := filepath.Base(blocks[1])
		fd, err := os.Open(blocks[1])
		if err != nil {
			fmt.Println("upload failed: ", err)
			return
		}
		fi, err := fd.Stat()
		if err != nil {
			fmt.Println("upload failed: ", err)
			return
		}
		podDir := blocks[2]
		if podDir == "." {
			podDir = currentDirectory
		}
		blockSize := blocks[3]
		compression := blocks[4]
		args := make(map[string]string)
		args["pod_dir"] = podDir
		args["block_size"] = blockSize
		data, err := fdfsAPI.uploadMultipartFile(apiFileUpload, fileName, fi.Size(), fd, args, "files", compression)
		if err != nil {
			fmt.Println("upload failed: ", err)
			return
		}
		var resp api.UploadFileResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("file upload: ", err)
			return
		}
		fmt.Println("reference : ", resp.References)
		currentPrompt = getCurrentPrompt()
	case "download":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 3 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		localDir := blocks[1]
		dirStat, err := os.Stat(localDir)
		if err != nil {
			fmt.Println("local path is not a present: ", err)
			return
		}

		if !dirStat.IsDir() {
			fmt.Println("local path is not a directory")
			return
		}

		// Create the file
		loalFile := filepath.Join(localDir + utils.PathSeperator + filepath.Base(blocks[2]))
		out, err := os.Create(loalFile)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
		defer out.Close()

		podFile := blocks[2]
		if !strings.HasPrefix(podFile, utils.PathSeperator) {
			if currentDirectory == utils.PathSeperator {
				podFile = currentDirectory + podFile
			} else {
				podFile = currentDirectory + utils.PathSeperator + podFile
			}
		}
		args := make(map[string]string)
		args["file"] = podFile
		n, err := fdfsAPI.downloadMultipartFile(http.MethodPost, apiFileDownload, args, out)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
		fmt.Println("Downloaded ", n, " bytes")
		currentPrompt = getCurrentPrompt()
	case "cat":
		//if !isPodOpened() {
		//	return
		//}
		//if len(blocks) < 2 {
		//	fmt.Println("invalid command. Missing one or more arguments")
		//	return
		//}
		//err := dfsAPI.Cat(blocks[1], DefaultSessionId)
		//if err != nil {
		//	fmt.Println("cat failed: ", err)
		//	return
		//}
		//currentPrompt = getCurrentPrompt()
	case "stat":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		statElement := blocks[1]
		if statElement == "" {
			return
		}
		if !strings.HasPrefix(statElement, utils.PathSeperator) {
			if currentDirectory == utils.PathSeperator {
				statElement = currentDirectory + statElement
			} else {
				statElement = currentDirectory + utils.PathSeperator + statElement
			}
		}
		args := make(map[string]string)
		args["dir"] = statElement
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiDirStat, args)
		if err != nil {
			if err.Error() == "dir stat: directory not found" {
				args := make(map[string]string)
				args["file"] = statElement
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, apiFileStat, args)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				var resp file.FileStats
				err = json.Unmarshal(data, &resp)
				if err != nil {
					fmt.Println("file stat: ", err)
					return
				}
				crTime, err := strconv.ParseInt(resp.CreationTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				accTime, err := strconv.ParseInt(resp.AccessTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				modTime, err := strconv.ParseInt(resp.ModificationTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				compression := resp.Compression
				if compression == "" {
					compression = "None"
				}
				fmt.Println("Account 	   	: ", resp.Account)
				fmt.Println("PodName 	   	: ", resp.PodName)
				fmt.Println("File Path	   	: ", resp.FilePath)
				fmt.Println("File Name	   	: ", resp.FileName)
				fmt.Println("File Size	   	: ", resp.FileSize)
				fmt.Println("Block Size	   	: ", resp.BlockSize)
				fmt.Println("Compression   		: ", compression)
				fmt.Println("Content Type  		: ", resp.ContentType)
				fmt.Println("Cr. Time	   	: ", time.Unix(crTime, 0).String())
				fmt.Println("Mo. Time	   	: ", time.Unix(accTime, 0).String())
				fmt.Println("Ac. Time	   	: ", time.Unix(modTime, 0).String())
				for _, b := range resp.Blocks {
					blkStr := fmt.Sprintf("%s, 0x%s, %s bytes, %s bytes", b.Name, b.Reference, b.Size, b.CompressedSize)
					fmt.Println(blkStr)
				}
			} else {
				fmt.Println("stat: ", err)
				return
			}
		} else {
			var resp dir.DirStats
			err = json.Unmarshal(data, &resp)
			if err != nil {
				fmt.Println("file stat: ", err)
				return
			}
			crTime, err := strconv.ParseInt(resp.CreationTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			accTime, err := strconv.ParseInt(resp.AccessTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			modTime, err := strconv.ParseInt(resp.ModificationTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			fmt.Println("Account 	   	: ", resp.Account)
			fmt.Println("PodAddress    		: ", resp.PodAddress)
			fmt.Println("PodName 	   	: ", resp.PodName)
			fmt.Println("Dir Path	   	: ", resp.DirPath)
			fmt.Println("Dir Name	   	: ", resp.DirName)
			fmt.Println("Cr. Time	   	: ", time.Unix(crTime, 0).String())
			fmt.Println("Mo. Time	   	: ", time.Unix(accTime, 0).String())
			fmt.Println("Ac. Time	   	: ", time.Unix(modTime, 0).String())
			fmt.Println("No of Dir.	   	: ", resp.NoOfDirectories)
			fmt.Println("No of Files   		: ", resp.NoOfFiles)
		}
		currentPrompt = getCurrentPrompt()
	case "pwd":
		if !isPodOpened() {
			return
		}
		fmt.Println(currentDirectory)
		currentPrompt = getCurrentPrompt()
	case "rm":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		rmFile := blocks[1]
		if rmFile == "" {
			return
		}
		if !strings.HasPrefix(rmFile, utils.PathSeperator) {
			if currentDirectory == utils.PathSeperator {
				rmFile = currentDirectory + rmFile
			} else {
				rmFile = currentDirectory + utils.PathSeperator + rmFile
			}
		}

		args := make(map[string]string)
		args["file"] = rmFile
		data, err := fdfsAPI.callFdfsApi(http.MethodDelete, apiFileDelete, args)
		if err != nil {
			fmt.Println("rm failed: ", err)
			return
		}
		message := strings.ReplaceAll(string(data), "\n", "")
		fmt.Println(message)
		currentPrompt = getCurrentPrompt()
	case "share":
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		podFile := blocks[1]

		if podFile == "" {
			return
		}
		if !strings.HasPrefix(podFile, utils.PathSeperator) {
			if currentDirectory == utils.PathSeperator {
				podFile = currentDirectory + podFile
			} else {
				podFile = currentDirectory + utils.PathSeperator + podFile
			}
		}

		args := make(map[string]string)
		args["file"] = podFile
		args["to"] = "add destination user address later"
		data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiFileShare, args)
		if err != nil {
			fmt.Println("share: ", err)
			return
		}
		var resp api.FileSharingReference
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("file share: ", err)
			return
		}
		fmt.Println("File Sharing Reference: ", resp.Reference)
		currentPrompt = getCurrentPrompt()
	case "receive":
		if len(blocks) < 3 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		sharingRefString := blocks[1]
		podDir := blocks[2]
		args := make(map[string]string)
		args["ref"] = sharingRefString
		args["dir"] = podDir
		data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiFileReceive, args)
		if err != nil {
			fmt.Println("receive: ", err)
			return
		}
		var resp api.ReceiveFileResponse
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("file receive: ", err)
			return
		}
		fmt.Println("file path  : ", resp.FileName)
		fmt.Println("reference  : ", resp.Reference)
		currentPrompt = getCurrentPrompt()
	case "receiveinfo":
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		sharingRefString := blocks[1]
		args := make(map[string]string)
		args["ref"] = sharingRefString
		data, err := fdfsAPI.callFdfsApi(http.MethodPost, apiFileReceiveInfo, args)
		if err != nil {
			fmt.Println("receive info: ", err)
			return
		}
		var resp user.ReceiveFileInfo
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("file receiveinfo: ", err)
			return
		}
		shTime, err := strconv.ParseInt(resp.SharedTime, 10, 64)
		if err != nil {
			fmt.Println(" info: ", err)
			return
		}
		fmt.Println("FileName       : ", resp.FileName)
		fmt.Println("Size           : ", resp.Size)
		fmt.Println("BlockSize      : ", resp.BlockSize)
		fmt.Println("NumberOfBlocks : ", resp.NumberOfBlocks)
		fmt.Println("ContentType    : ", resp.ContentType)
		fmt.Println("Compression    : ", resp.Compression)
		fmt.Println("PodName        : ", resp.PodName)
		fmt.Println("FileMetaHash   : ", resp.FileMetaHash)
		fmt.Println("Sender         : ", resp.Sender)
		fmt.Println("Receiver       : ", resp.Receiver)
		fmt.Println("SharedTime     : ", shTime)
		currentPrompt = getCurrentPrompt()
	case "mv":
		fmt.Println("not yet implemented")
	case "head":
		fmt.Println("not yet implemented")
	default:
		fmt.Println("invalid command")
	}
}

func help() {
	fmt.Println("Usage: <command> <sub-command> (args1) (args2) ...")
	fmt.Println("commands:")
	fmt.Println(" - user <new> (user-name) - create a new user and login as that user")
	fmt.Println(" - user <del> - deletes a logged in user")
	fmt.Println(" - user <login> (user-name) - login as a given user")
	fmt.Println(" - user <logout> - logout a logged in user")
	fmt.Println(" - user <present> (user-name) - returns true if the user is present, false otherwise")
	fmt.Println(" - user <ls> - lists all the user present in this instance")
	fmt.Println(" - user <name> (first_name) (middle_name) (last_name) (surname) - sets the user name information")
	fmt.Println(" - user <name> - gets the user name information")
	fmt.Println(" - user <contact> (phone) (mobile) (address_line1) (address_line2) (state) (zipcode) - sets the user contact information")
	fmt.Println(" - user <contact> gets the user contact information")
	fmt.Println(" - user <share> <inbox> - shows details of the files you have received from other users")
	fmt.Println(" - user <share> <outbox> - shows details of the files you have sent to other users")
	fmt.Println(" - user <export> - exports the given user")
	fmt.Println(" - user <import> (user-name) (address) - imports the user to another device")
	fmt.Println(" - user <import> (user-name) (12 word mnemonic) - imports the user if the device is lost")
	fmt.Println(" - user <stat> - shows information about a user")

	fmt.Println(" - pod <new> (pod-name) - create a new pod for the logged in user and opens the pod")
	fmt.Println(" - pod <del> (pod-name) - deletes a already created pod of the user")
	fmt.Println(" - pod <open> (pod-name) - open a already created pod")
	fmt.Println(" - pod <stat> (pod-name) - display meta information about a pod")
	fmt.Println(" - pod <sync> (pod-name) - sync the contents of a logged in pod from Swarm")
	fmt.Println(" - pod <close>  - close a opened pod")
	fmt.Println(" - pod <ls> - lists all the pods created for this account")

	fmt.Println(" - kv <new> (table-name) - creates a new key value store")
	fmt.Println(" - kv <delete> (table-name) - deletes the key value store")
	fmt.Println(" - kv <open> (table-name) - open the key value store")
	fmt.Println(" - kv <ls>  - list all collections")
	fmt.Println(" - kv <put> (table-name) (key) (value) - insertkey and value in to kv store")
	fmt.Println(" - kv <get> (table-name) (key) - get the value of the given key from the store")
	fmt.Println(" - kv <del> (table-name) (key) - remove the key and value from the store")
	fmt.Println(" - kv <loadcsv> (table-name) (local csv file) - load the csv file in to the newy created table")
	fmt.Println(" - kv <seek> (table-name) (start-key) (end-key) (limit) - seek nearst to start key")
	fmt.Println(" - kv <getnext> (table-name) - get the next element after seek")

	fmt.Println(" - doc <new> (table-name) (si=indexes) - creates a new document store")
	fmt.Println(" - doc <delete> (table-name) - deletes a document store")
	fmt.Println(" - doc <open> (table-name) - open the document store")
	fmt.Println(" - doc <ls>  - list all document dbs")
	fmt.Println(" - doc <count> (table-name) (expr) - count the docs in the table satisfying the expression")
	fmt.Println(" - doc <find> (table-name) (expr) (limit)- find the docs in the table satisfying the expression and limit")
	fmt.Println(" - doc <put> (table-name) (json) - insert a json document in to document store")
	fmt.Println(" - doc <get> (table-name) (id) - get the document having the id from the store")
	fmt.Println(" - doc <del> (table-name) (id) - delete the document having the id from the store")
	fmt.Println(" - doc <loadjson> (table-name) (local json file) - load the json file in to the newly created document db")

	fmt.Println(" - cd <directory name>")
	fmt.Println(" - ls ")
	fmt.Println(" - download <relative path of source file in pod, destination dir in local fs>")
	fmt.Println(" - upload <source file in local fs, destination directory in pod, block size (ex: 1Mb, 64Mb)>, compression true/false")
	fmt.Println(" - share <file name> -  shares a file with another user")
	fmt.Println(" - receive <sharing reference> <pod dir> - receives a file from another user")
	fmt.Println(" - receiveinfo <sharing reference> - shows the received file info before accepting the receive")
	fmt.Println(" - mkdir <directory name>")
	fmt.Println(" - rmdir <directory name>")
	fmt.Println(" - rm <file name>")
	fmt.Println(" - pwd - show present working directory")
	fmt.Println(" - cat  - stream the file to stdout")
	fmt.Println(" - stat <file name or directory name> - shows the information about a file or directory")
	fmt.Println(" - help - display this help")
	fmt.Println(" - exit - exits from the prompt")

}

func getCurrentPrompt() string {
	currPrompt := getUserPrompt()
	podPrompt := getPodPrompt()
	if podPrompt != "" {
		currPrompt = currPrompt + " " + podPrompt + " " + PodSeperator
	}
	dirPrompt := currentDirectory
	if dirPrompt != "" {
		currPrompt = currPrompt + " " + dirPrompt + " " + PromptSeperator
	}
	return currPrompt
}

func isPodOpened() bool {
	if currentPod == "" {
		fmt.Println("open the pod to do the operation")
		return false
	}
	return true
}

func getUserPrompt() string {
	if currentUser == "" {
		return DefaultPrompt + " " + UserSeperator
	} else {
		return DefaultPrompt + "@" + currentUser + " " + UserSeperator
	}
}

func getPodPrompt() string {
	if currentPod != "" {
		return currentPod
	} else {
		return ""
	}
}

func getPassword() (password string) {
	fmt.Print("Please enter your password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatalf("error reading password")
		return
	}
	fmt.Println("")
	passwd := string(bytePassword)
	password = strings.TrimSpace(passwd)
	return password
}
