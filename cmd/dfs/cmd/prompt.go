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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	DefaultPrompt    = "dfs"
	UserSeperator    = ">>>"
	PodSeperator     = ">>"
	PromptSeperator  = "> "
	DefaultSessionId = "12345678"
)

var (
	currentUser    string
	currentPodInfo *pod.Info
	currentPrompt  string
	dfsAPI         *dfs.DfsAPI
	logger         logging.Logger
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "a REPL to interact with FairOS's dfs",
	Long: `A command prompt where you can interact with the distributed
file system of the FairOS.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger = logging.New(ioutil.Discard, 0)
		api, err := dfs.NewDfsAPI(dataDir, beeHost, beePort, logger)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		dfsAPI = api
		initPrompt()
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
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
			ref, mnemonic, err := dfsAPI.CreateUser(userName, "", "", nil, DefaultSessionId)
			if err != nil {
				fmt.Println("create user: ", err)
				return
			}
			fmt.Println("user created with address ", ref)
			fmt.Println("Please store the following 12 words safely")
			fmt.Println("if you loose this, you cannot recover the data in-case of an emergency.")
			fmt.Println("you can also use this mnemonic to access the datain case this device is lost")
			fmt.Println("=============== Mnemonic ==========================")
			fmt.Println(mnemonic)
			fmt.Println("=============== Mnemonic ==========================")
			currentUser = userName
			currentPodInfo = nil
			currentPrompt = getCurrentPrompt()
		case "export":
			name, address, err := dfsAPI.ExportUser(DefaultSessionId)
			if err != nil {
				fmt.Println("export user: ", err)
				return
			}
			fmt.Println("user name:", name)
			fmt.Println("address  :", address)
			currentPrompt = getCurrentPrompt()
		case "import":
			if len(blocks) == 4 {
				userName := blocks[2]
				address := blocks[3]
				err := dfsAPI.ImportUserUsingAddress(userName, "", address, nil, DefaultSessionId)
				if err != nil {
					fmt.Println("import user: ", err)
					return
				}
				fmt.Println("user imported")
				currentUser = userName
				currentPodInfo = nil
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
			_, err := dfsAPI.ImportUserUsingMnemonic(userName, "", mnemonic, nil, DefaultSessionId)
			if err != nil {
				fmt.Println("import user: ", err)
				return
			}
			currentUser = userName
			currentPodInfo = nil
			currentPrompt = getCurrentPrompt()
		case "del":
			err := dfsAPI.DeleteUser("", DefaultSessionId, nil)
			if err != nil {
				fmt.Println("delete user: ", err)
				return
			}
			currentUser = ""
			currentPodInfo = nil
			currentPrompt = getCurrentPrompt()
		case "login":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			err := dfsAPI.LoginUser(userName, "", nil, DefaultSessionId)
			if err != nil {
				fmt.Println("login user: ", err)
				return
			}
			currentUser = userName
			currentPodInfo = nil
			currentPrompt = getCurrentPrompt()
		case "logout":
			err := dfsAPI.LogoutUser(DefaultSessionId, nil)
			if err != nil {
				fmt.Println("logout user: ", err)
				return
			}
			currentUser = ""
			currentPodInfo = nil
			currentPrompt = getCurrentPrompt()
		case "present":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			userName := blocks[2]
			yes := dfsAPI.IsUserNameAvailable(userName)
			if yes {
				fmt.Println("true")
			} else {
				fmt.Println("false")
			}
			currentPrompt = getCurrentPrompt()
		case "ls":
			users := dfsAPI.ListAllUsers()
			if users != nil {
				for _, usr := range users {
					fmt.Println(usr)
				}
			}
			currentPrompt = getCurrentPrompt()
		case "name":
			if len(blocks) == 6 {
				firstName := blocks[2]
				middleName := blocks[3]
				lastName := blocks[4]
				surNmae := blocks[5]
				err := dfsAPI.SaveName(firstName, lastName, middleName, surNmae, DefaultSessionId)
				if err != nil {
					fmt.Println("name: ", err)
					return
				}
			} else if len(blocks) == 2 {
				name, err := dfsAPI.GetName(DefaultSessionId)
				if err != nil {
					fmt.Println("name: ", err)
					return
				}
				fmt.Println("first_name : ", name.FirstName)
				fmt.Println("middle_name: ", name.MiddleName)
				fmt.Println("last_name  : ", name.LastName)
				fmt.Println("surname    : ", name.SurName)
			}
			currentPrompt = getCurrentPrompt()
		case "contact":
			if len(blocks) == 8 {
				phone := blocks[2]
				mobile := blocks[3]
				address_line1 := blocks[4]
				address_line2 := blocks[5]
				state := blocks[6]
				zip := blocks[7]
				addr := &user.Address{
					AddressLine1: address_line1,
					AddressLine2: address_line2,
					State:        state,
					ZipCode:      zip,
				}
				err := dfsAPI.SaveContact(phone, mobile, addr, DefaultSessionId)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
			} else if len(blocks) == 2 {
				contacts, err := dfsAPI.GetContact(DefaultSessionId)
				if err != nil {
					fmt.Println("contact: ", err)
					return
				}
				fmt.Println("phone        : ", contacts.Phone)
				fmt.Println("mobile       : ", contacts.Mobile)
				fmt.Println("address_line1: ", contacts.Addr.AddressLine1)
				fmt.Println("address_line2: ", contacts.Addr.AddressLine2)
				fmt.Println("state        : ", contacts.Addr.State)
				fmt.Println("zipcode      : ", contacts.Addr.ZipCode)
			}
			currentPrompt = getCurrentPrompt()
		case "share":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"inbox/outbox\" argument ")
				return
			}
			switch blocks[2] {
			case "inbox":
				inbox, err := dfsAPI.GetUserSharingInbox(DefaultSessionId)
				if err != nil {
					fmt.Println("sharing inbox: ", err)
					return
				}
				for _, entry := range inbox.Entries {
					fmt.Println(entry)
				}
			case "outbox":
				outbox, err := dfsAPI.GetUserSharingOutbox(DefaultSessionId)
				if err != nil {
					fmt.Println("sharing outbox: ", err)
					return
				}
				for _, entry := range outbox.Entries {
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
			podInfo, err := dfsAPI.CreatePod(podName, "", DefaultSessionId)
			if err != nil {
				fmt.Println("could not create pod: ", err)
				return
			}
			currentPodInfo = podInfo
			currentPrompt = getCurrentPrompt()
		case "del":
			lastPrompt := currentPrompt
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			err := dfsAPI.DeletePod(podName, DefaultSessionId)
			if err != nil {
				fmt.Println("could not delete pod: ", err)
				return
			}
			fmt.Println("successfully deleted pod: ", podName)
			if podName == currentPodInfo.GetCurrentPodNameOnly() {
				currentPrompt = DefaultPrompt
			} else {
				currentPrompt = lastPrompt
			}
		case "open":
			if len(blocks) < 3 {
				fmt.Println("invalid command. Missing \"name\" argument ")
				return
			}
			podName := blocks[2]
			podInfo, err := dfsAPI.OpenPod(podName, "", DefaultSessionId)
			if err != nil {
				fmt.Println("Open failed: ", err)
				return
			}
			currentPodInfo = podInfo
			currentPrompt = getCurrentPrompt()
		case "close":
			if !isPodOpened() {
				return
			}
			err := dfsAPI.ClosePod(DefaultSessionId)
			if err != nil {
				fmt.Println("error logging out: ", err)
				return
			}
			currentPrompt = DefaultPrompt + " " + UserSeperator
			currentPodInfo = nil
		case "stat":
			if !isPodOpened() {
				return
			}
			podStat, err := dfsAPI.PodStat(currentPodInfo.GetCurrentPodNameOnly(), DefaultSessionId)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			crTime, err := strconv.ParseInt(podStat.CreationTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			accTime, err := strconv.ParseInt(podStat.AccessTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			modTime, err := strconv.ParseInt(podStat.ModificationTime, 10, 64)
			if err != nil {
				fmt.Println("error getting stat: ", err)
				return
			}
			fmt.Println("Version          : ", podStat.Version)
			fmt.Println("pod Name         : ", podStat.PodName)
			fmt.Println("Path             : ", podStat.PodPath)
			fmt.Println("Creation Time    :", time.Unix(crTime, 0).String())
			fmt.Println("Access Time      :", time.Unix(accTime, 0).String())
			fmt.Println("Modification Time:", time.Unix(modTime, 0).String())
			currentPrompt = getCurrentPrompt()
		case "sync":
			if !isPodOpened() {
				return
			}
			err := dfsAPI.SyncPod(DefaultSessionId)
			if err != nil {
				fmt.Println("could not sync pod: ", err)
				return
			}
			fmt.Println("pod synced.")
			currentPrompt = getCurrentPrompt()
		case "ls":
			pods, err := dfsAPI.ListPods(DefaultSessionId)
			if err != nil {
				fmt.Println("error while listing pods: %w", err)
				return
			}
			for _, pod := range pods {
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
		podInfo, err := dfsAPI.ChangeDirectory(blocks[1], DefaultSessionId)
		if err != nil {
			fmt.Println("cd failed: ", err)
			return
		}
		currentPodInfo = podInfo
		currentPrompt = getCurrentPrompt()
	case "ls":
		if !isPodOpened() {
			return
		}
		entries, err := dfsAPI.ListDir("", DefaultSessionId)
		if err != nil {
			fmt.Println("ls failed: ", err)
			return
		}
		for _, entry := range entries {
			if entry.ContentType == "inode/directory" {
				fmt.Println("<Dir>: ", entry.Name)
			} else {
				fmt.Println("<File>: ", entry.Name)
			}

		}
		currentPrompt = getCurrentPrompt()
	case "download":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 3 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		err := dfsAPI.CopyToLocal(blocks[1], blocks[2], DefaultSessionId)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
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
		ref, err := dfsAPI.UploadFile(fileName, DefaultSessionId, fi.Size(), fd, podDir, blockSize, compression)
		if err != nil {
			fmt.Println("upload failed: ", err)
			return
		}
		fmt.Println("reference : ", ref)
		currentPrompt = getCurrentPrompt()
	case "mkdir":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		err := dfsAPI.Mkdir(blocks[1], DefaultSessionId)
		if err != nil {
			fmt.Println("mkdir failed: ", err)
			return
		}
		currentPrompt = getCurrentPrompt()
	case "rmdir":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		err := dfsAPI.RmDir(blocks[1], DefaultSessionId)
		if err != nil {
			fmt.Println("rmdir failed: ", err)
			return
		}
		currentPrompt = getCurrentPrompt()
	case "cat":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		err := dfsAPI.Cat(blocks[1], DefaultSessionId)
		if err != nil {
			fmt.Println("cat failed: ", err)
			return
		}
		currentPrompt = getCurrentPrompt()
	case "stat":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		ds, err := dfsAPI.DirectoryStat(blocks[1], DefaultSessionId, true)
		if err != nil {
			if err.Error() == "directory not found" {
				fs, err := dfsAPI.FileStat(blocks[1], DefaultSessionId)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				crTime, err := strconv.ParseInt(fs.CreationTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				accTime, err := strconv.ParseInt(fs.AccessTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				modTime, err := strconv.ParseInt(fs.ModificationTime, 10, 64)
				if err != nil {
					fmt.Println("stat failed: ", err)
					return
				}
				compression := fs.Compression
				if compression == "" {
					compression = "None"
				}
				fmt.Println("Account 	   	: ", fs.Account)
				fmt.Println("PodName 	   	: ", fs.PodName)
				fmt.Println("File Path	   	: ", fs.FilePath)
				fmt.Println("File Name	   	: ", fs.FileName)
				fmt.Println("File Size	   	: ", fs.FileSize)
				fmt.Println("Block Size	   	: ", fs.BlockSize)
				fmt.Println("Compression   		: ", compression)
				fmt.Println("Content Type  		: ", fs.ContentType)
				fmt.Println("Cr. Time	   	: ", time.Unix(crTime, 0).String())
				fmt.Println("Mo. Time	   	: ", time.Unix(accTime, 0).String())
				fmt.Println("Ac. Time	   	: ", time.Unix(modTime, 0).String())
				for _, b := range fs.Blocks {
					blkStr := fmt.Sprintf("%s, 0x%s, %s bytes, %s bytes", b.Name, b.Reference, b.Size, b.CompressedSize)
					fmt.Println(blkStr)
				}
			} else {
				fmt.Println("stat: %w", err)
				return
			}
		} else {
			crTime, err := strconv.ParseInt(ds.CreationTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			accTime, err := strconv.ParseInt(ds.AccessTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			modTime, err := strconv.ParseInt(ds.ModificationTime, 10, 64)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			fmt.Println("Account 	   	: ", ds.Account)
			fmt.Println("PodAddress    		: ", ds.PodAddress)
			fmt.Println("PodName 	   	: ", ds.PodName)
			fmt.Println("Dir Path	   	: ", ds.DirPath)
			fmt.Println("Dir Name	   	: ", ds.DirName)
			fmt.Println("Cr. Time	   	: ", time.Unix(crTime, 0).String())
			fmt.Println("Mo. Time	   	: ", time.Unix(accTime, 0).String())
			fmt.Println("Ac. Time	   	: ", time.Unix(modTime, 0).String())
			fmt.Println("No of Dir.	   	: ", ds.NoOfDirectories)
			fmt.Println("No of Files   		: ", ds.NoOfFiles)
		}
		currentPrompt = getCurrentPrompt()
	case "pwd":
		if !isPodOpened() {
			return
		}
		if currentPodInfo.IsCurrentDirRoot() {
			fmt.Println("/")
		} else {
			podDir := currentPodInfo.GetCurrentPodPathAndName()
			curDir := strings.TrimPrefix(currentPodInfo.GetCurrentDirPathAndName(), podDir)
			fmt.Println(curDir)
		}
		currentPrompt = getCurrentPrompt()
	case "rm":
		if !isPodOpened() {
			return
		}
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		err := dfsAPI.DeleteFile(blocks[1], DefaultSessionId)
		if err != nil {
			fmt.Println("rm failed: ", err)
			return
		}
		currentPrompt = getCurrentPrompt()
	case "share":
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		podFile := blocks[1]
		sharingRef, err := dfsAPI.ShareFile(podFile, currentUser, DefaultSessionId)
		if err != nil {
			fmt.Println("share: ", err)
			return
		}
		fmt.Println("Sharing Reference: ", sharingRef)
		currentPrompt = getCurrentPrompt()
	case "receive":
		if len(blocks) < 3 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		sharingRefString := blocks[1]
		podDir := blocks[2]

		sharingRef, err := utils.ParseSharingReference(sharingRefString)
		if err != nil {
			fmt.Println("receive: ", err)
			return
		}
		filePath, metaRef, err := dfsAPI.ReceiveFile(DefaultSessionId, sharingRef, podDir)
		if err != nil {
			fmt.Println("receive: ", err)
			return
		}
		fmt.Println("file path  : ", filePath)
		fmt.Println("reference  : ", metaRef)
		currentPrompt = getCurrentPrompt()
	case "receiveinfo":
		if len(blocks) < 2 {
			fmt.Println("invalid command. Missing one or more arguments")
			return
		}
		sharingRefString := blocks[1]
		sharingRef, err := utils.ParseSharingReference(sharingRefString)
		if err != nil {
			fmt.Println("receive info: ", err)
			return
		}
		ri, err := dfsAPI.ReceiveInfo(DefaultSessionId, sharingRef)
		if err != nil {
			fmt.Println("receive info: ", err)
			return
		}
		shTime, err := strconv.ParseInt(ri.SharedTime, 10, 64)
		if err != nil {
			fmt.Println(" info: ", err)
			return
		}
		fmt.Println("FileName       : ", ri.FileName)
		fmt.Println("Size           : ", ri.Size)
		fmt.Println("BlockSize      : ", ri.BlockSize)
		fmt.Println("NumberOfBlocks : ", ri.NumberOfBlocks)
		fmt.Println("ContentType    : ", ri.ContentType)
		fmt.Println("Compression    : ", ri.Compression)
		fmt.Println("PodName        : ", ri.PodName)
		fmt.Println("FileMetaHash   : ", ri.FileMetaHash)
		fmt.Println("Sender         : ", ri.Sender)
		fmt.Println("Receiver       : ", ri.Receiver)
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
	dirPrompt := getCurrentDirPrompt()
	if dirPrompt != "" {
		currPrompt = currPrompt + " " + dirPrompt + " " + PromptSeperator
	}
	return currPrompt
}

func isPodOpened() bool {
	if currentPodInfo == nil {
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
	if currentPodInfo != nil {
		return currentPodInfo.GetCurrentPodNameOnly()
	} else {
		return ""
	}
}

func getCurrentDirPrompt() string {
	currentDir := ""
	if currentPodInfo != nil {
		if currentPodInfo.IsCurrentDirRoot() {
			return utils.PathSeperator
		}
		podPathAndName := currentPodInfo.GetCurrentPodPathAndName()
		pathExceptPod := strings.TrimPrefix(currentPodInfo.GetCurrentDirPathOnly(), podPathAndName)
		currentDir = pathExceptPod + utils.PathSeperator + currentPodInfo.GetCurrentDirNameOnly()
	}
	return currentDir
}
