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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	DefaultPrompt    = "dfs"
	UserSeperator    = ">>>"
	PodSeperator     = ">>"
	PromptSeperator  = "> "
	DefaultSessionId = "12345678"
	APIVersion       = "/v0"
)

var (
	currentUser      string
	currentPod       string
	currentPrompt    string
	currentDirectory string
	fdfsAPI          *FdfsClient
)

const (
	API_USER_SIGNUP       = APIVersion + "/user/signup"
	API_USER_LOGIN        = APIVersion + "/user/login"
	API_USER_IMPORT       = APIVersion + "/user/import"
	API_USER_PRESENT      = APIVersion + "/user/present"
	API_USER_ISLOGGEDIN   = APIVersion + "/user/isloggedin"
	API_USER_LOGOUT       = APIVersion + "/user/logout"
	API_USER_AVATAR       = APIVersion + "/user/avatar"
	API_USER_NAME         = APIVersion + "/user/name"
	API_USER_CONTACT      = APIVersion + "/user/contact"
	API_USER_EXPORT       = APIVersion + "/user/export"
	API_USER_DELETE       = APIVersion + "/user/delete"
	API_USER_STAT         = APIVersion + "/user/stat"
	API_USER_SHARE_INBOX  = APIVersion + "/user/share/inbox"
	API_USER_SHARE_OUTBOX = APIVersion + "/user/share/inbox"
	API_POD_NEW           = APIVersion + "/pod/new"
	API_POD_OPEN          = APIVersion + "/pod/open"
	API_POD_CLOSE         = APIVersion + "/pod/close"
	API_POD_SYNC          = APIVersion + "/pod/sync"
	API_POD_DELETE        = APIVersion + "/pod/delete"
	API_POD_LS            = APIVersion + "/pod/ls"
	API_POD_STAT          = APIVersion + "/pod/stat"
	API_DIR_ISPRESENT     = APIVersion + "/dir/present"
	API_DIR_MKDIR         = APIVersion + "/dir/mkdir"
	API_DIR_RMDIR         = APIVersion + "/dir/rmdir"
	API_DIR_LS            = APIVersion + "/dir/ls"
	API_DIR_STAT          = APIVersion + "/dir/stat"
	API_FILE_DOWNLOAD     = APIVersion + "/file/download"
	API_FILE_UPLOAD       = APIVersion + "/file/upload"
	API_FILE_SHARE        = APIVersion + "/file/share"
	API_FILE_RECEIVE      = APIVersion + "/file/receive"
	API_FILE_RECEIVEINFO  = APIVersion + "/file/receiveinfo"
	API_FILE_DELETE       = APIVersion + "/file/delete"
	API_FILE_STAT         = APIVersion + "/file/stat"
)

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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_SIGNUP, args)
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
				data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_IMPORT, args)
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_IMPORT, args)
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_LOGIN, args)
			if err != nil {
				fmt.Println("login user: ", err)
				return
			}
			fmt.Println(string(data))
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_PRESENT, args)
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
			data, err := fdfsAPI.callFdfsApi(http.MethodDelete, API_USER_DELETE, args)
			if err != nil {
				fmt.Println("delete user: ", err)
				return
			}
			fmt.Println(string(data))
			currentUser = ""
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "logout":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_LOGOUT, nil)
			if err != nil {
				fmt.Println("logout user: ", err)
				return
			}
			fmt.Println(string(data))
			currentUser = ""
			currentPod = ""
			currentDirectory = ""
			currentPrompt = getCurrentPrompt()
		case "export":
			if currentUser == "" {
				fmt.Println("please login as  user to do the operation")
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_EXPORT, nil)
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
				_, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_NAME, args)
				if err != nil {
					fmt.Println("name: ", err)
					return
				}
			} else if len(blocks) == 2 {
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_USER_NAME, nil)
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
				address_line1 := blocks[4]
				address_line2 := blocks[5]
				state := blocks[6]
				zip := blocks[7]
				args := make(map[string]string)
				args["phone"] = phone
				args["mobile"] = mobile
				args["address_line_1"] = address_line1
				args["address_line_2"] = address_line2
				args["state_province_region"] = state
				args["zipcode"] = zip
				_, err := fdfsAPI.callFdfsApi(http.MethodPost, API_USER_CONTACT, args)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
			} else if len(blocks) == 2 {
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_USER_CONTACT, nil)
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
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_USER_SHARE_INBOX, nil)
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
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_USER_SHARE_OUTBOX, nil)
				if err != nil {
					fmt.Println("sharing outbox: ", err)
					return
				}
				var resp user.Inbox
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_POD_NEW, args)
			if err != nil {
				fmt.Println("could not create pod: ", err)
				return
			}
			fmt.Println(string(data))
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_POD_DELETE, args)
			if err != nil {
				fmt.Println("could not delete pod: ", err)
				return
			}
			fmt.Println(string(data))
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
			args["password"] = getPassword()
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_POD_OPEN, args)
			if err != nil {
				fmt.Println("pod open failed: ", err)
				return
			}
			fmt.Println(string(data))
			currentPod = podName
			currentDirectory = utils.PathSeperator
			currentPrompt = getCurrentPrompt()
		case "close":
			if !isPodOpened() {
				return
			}
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_POD_CLOSE, nil)
			if err != nil {
				fmt.Println("error logging out: ", err)
				return
			}
			fmt.Println(string(data))
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
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_POD_STAT, args)
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
			data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_POD_SYNC, nil)
			if err != nil {
				fmt.Println("could not sync pod: ", err)
				return
			}
			fmt.Println(string(data))
			currentPrompt = getCurrentPrompt()
		case "ls":
			data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_POD_LS, nil)
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
			currentPrompt = getCurrentPrompt()
		default:
			fmt.Println("invalid pod command!!")
			help()
		} // end of pod commands
	case "cd":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		dirTocd := blocks[1]
		args := make(map[string]string)
		args["dir"] = dirTocd
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_DIR_ISPRESENT, nil)
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
			if currentDirectory == utils.PathSeperator {
				currentDirectory = currentDirectory + dirTocd
			} else {
				currentDirectory = currentDirectory + utils.PathSeperator + dirTocd
			}
		} else {
			fmt.Println("User is not present: ", resp.Error)
		}
		currentPrompt = getCurrentPrompt()
	case "ls":
		if !isPodOpened() {
			return
		}
		args := make(map[string]string)
		args["dir"] = currentDirectory
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_DIR_LS, args)
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
		args := make(map[string]string)
		args["dir"] = blocks[1]
		data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_DIR_MKDIR, args)
		if err != nil {
			fmt.Println("mkdir failed: ", err)
			return
		}
		fmt.Println(string(data))
		currentPrompt = getCurrentPrompt()
	case "rmdir":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		args := make(map[string]string)
		args["dir"] = blocks[1]
		data, err := fdfsAPI.callFdfsApi(http.MethodPost, API_DIR_RMDIR, args)
		if err != nil {
			fmt.Println("rmdir failed: ", err)
			return
		}
		fmt.Println(string(data))
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
		blockSize := blocks[3]
		compression := blocks[4]
		data, err := fdfsAPI.uploadMultipartFile(API_FILE_UPLOAD, fileName, fi.Size(), fd, podDir, blockSize, compression)
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

		args := make(map[string]string)
		args["file"] = blocks[2]
		n, err := fdfsAPI.downloadMultipartFile(http.MethodPost, API_FILE_DOWNLOAD, args, out)
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
		args := make(map[string]string)
		args["file"] = blocks[2]
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_DIR_STAT, args)
		if err != nil {
			if err.Error() == "directory not found" {
				args["dir"] = blocks[2]
				data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_FILE_STAT, args)
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
				fmt.Println("stat: %w", err)
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
		args := make(map[string]string)
		args["file"] = blocks[2]
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_FILE_DELETE, args)
		if err != nil {
			fmt.Println("rm failed: ", err)
			return
		}
		fmt.Println(string(data))
		currentPrompt = getCurrentPrompt()
	case "share":
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		podFile := blocks[1]
		args := make(map[string]string)
		args["file"] = podFile
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_FILE_SHARE, args)
		if err != nil {
			fmt.Println("share: ", err)
			return
		}
		var resp api.SharingReference
		err = json.Unmarshal(data, &resp)
		if err != nil {
			fmt.Println("file share: ", err)
			return
		}
		fmt.Println("Sharing Reference: ", resp.Reference)
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
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_FILE_RECEIVE, args)
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
		data, err := fdfsAPI.callFdfsApi(http.MethodGet, API_FILE_RECEIVEINFO, args)
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
	fmt.Println("Please enter your password: ")
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
