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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/cmd/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/api"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func isDirectoryPresent(podName, dirNameWithpath string) bool {
	args := fmt.Sprintf("podName=%s&dirPath=%s", podName, dirNameWithpath)
	data, err := fdfsAPI.getReq(apiDirIsPresent, args)
	if err != nil {
		fmt.Println("dir present: ", err)
		return false
	}
	var resp api.DirPresentResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		fmt.Println("dir present: ", err)
		return false
	}
	if !resp.Present {
		fmt.Println("dir present: ", resp.Error)
		return false
	}
	return resp.Present
}

func listFileAndDirectories(podName, dirNameWithpath string) (*api.ListFileResponse, error) {
	args := fmt.Sprintf("podName=%s&dirPath=%s", podName, dirNameWithpath)
	data, err := fdfsAPI.getReq(apiDirLs, args)
	if err != nil {
		return nil, err
	}
	var resp api.ListFileResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Directories == nil && resp.Files == nil {
		fmt.Println("empty directory")
	}
	for _, entry := range resp.Directories {
		fmt.Println("<Dir>: ", entry.Name)
	}
	for _, entry := range resp.Files {
		fmt.Println("<File>: ", entry.Name)
	}
	return &resp, nil
}

func statFileOrDirectory(podName, statElement string) {
	args := fmt.Sprintf("podName=%s&dirPath=%s", podName, statElement)
	data, err := fdfsAPI.getReq(apiDirStat, args)
	if err != nil {
		if strings.Contains(err.Error(), "directory not found") {
			args := fmt.Sprintf("podName=%s&filePath=%s", podName, statElement)
			data, err := fdfsAPI.getReq(apiFileStat, args)
			if err != nil {
				fmt.Println("stat failed: ", err)
				return
			}
			var resp file.Stats
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
			fmt.Println("PodName 	  : ", resp.PodName)
			fmt.Println("File Path	  : ", resp.FilePath)
			fmt.Println("File Name	  : ", resp.FileName)
			fmt.Println("File Size	  : ", resp.FileSize)
			fmt.Println("Block Size	  : ", resp.BlockSize)
			fmt.Println("Compression  	  : ", compression)
			fmt.Println("Content Type 	  : ", resp.ContentType)
			fmt.Println("Cr. Time	  : ", time.Unix(crTime, 0).String())
			fmt.Println("Mo. Time	  : ", time.Unix(accTime, 0).String())
			fmt.Println("Ac. Time	  : ", time.Unix(modTime, 0).String())
		} else {
			fmt.Println("stat: ", err)
			return
		}
	} else {
		var resp dir.Stats
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
		fmt.Println("PodName     	 : ", resp.PodName)
		fmt.Println("Dir Path    	 : ", resp.DirPath)
		fmt.Println("Dir Name	 : ", resp.DirName)
		fmt.Println("Cr. Time	 : ", time.Unix(crTime, 0).String())
		fmt.Println("Mo. Time	 : ", time.Unix(accTime, 0).String())
		fmt.Println("Ac. Time	 : ", time.Unix(modTime, 0).String())
		fmt.Println("No of Dir.	 : ", resp.NoOfDirectories)
		fmt.Println("No of Files      : ", resp.NoOfFiles)
	}
}

func mkdir(podName, dirNameWithpath string) {
	mkdirReq := common.FileSystemRequest{
		PodName:       podName,
		DirectoryPath: dirNameWithpath,
	}
	jsonData, err := json.Marshal(mkdirReq)
	if err != nil {
		fmt.Println("mkdir: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiDirMkdir, jsonData)
	if err != nil {
		fmt.Println("mkdir: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func rmDir(podName, dirNameWithpath string) {
	rmdirReq := common.FileSystemRequest{
		PodName:       podName,
		DirectoryPath: dirNameWithpath,
	}
	jsonData, err := json.Marshal(rmdirReq)
	if err != nil {
		fmt.Println("rmdir: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiDirRmdir, jsonData)
	if err != nil {
		fmt.Println("rmdir failed: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}

func uploadFile(fileName, podName, localFileWithPath, podDir, blockSize, compression string) {
	fd, err := os.Open(localFileWithPath)
	if err != nil {
		fmt.Println("upload failed: ", err)
		return
	}
	fi, err := fd.Stat()
	if err != nil {
		fmt.Println("upload failed: ", err)
		return
	}

	args := make(map[string]string)
	args["podName"] = podName
	args["dirPath"] = podDir
	args["blockSize"] = blockSize
	args["overwrite"] = "true"
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
	for _, response := range resp.Responses {
		fmt.Println(response.FileName, " : ", response.Message)
	}
}

func downloadFile(podName, localFileName, podFileName string) {
	// Create the local file fd
	out, err := os.Create(localFileName)
	if err != nil {
		fmt.Println("download failed: ", err)
		return
	}
	if err = out.Close(); err != nil {
		fmt.Println("download failed: ", err)
		return
	}

	args := make(map[string]string)
	args["podName"] = podName
	args["filePath"] = podFileName
	n, err := fdfsAPI.downloadMultipartFile(http.MethodPost, apiFileDownload, args, out)
	if err != nil {
		fmt.Println("download failed: ", err)
		return
	}
	fmt.Println("Downloaded ", n, " bytes")
}

func fileShare(podName, fileNameWithPath, destinationUser string) {
	rmdirReq := common.FileSystemRequest{
		PodName:     podName,
		FilePath:    fileNameWithPath,
		Destination: destinationUser,
	}
	jsonData, err := json.Marshal(rmdirReq)
	if err != nil {
		fmt.Println("share file: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodPost, apiFileShare, jsonData)
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
}

func fileReceiveInfo(podName, sharingRef string) {
	args := fmt.Sprintf("sharingRef=%s", sharingRef)
	data, err := fdfsAPI.getReq(apiFileReceiveInfo, args)
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
	fmt.Println("Sender         : ", resp.Sender)
	fmt.Println("Receiver       : ", resp.Receiver)
	fmt.Println("SharedTime     : ", shTime)
}

func fileReceive(podName, sharingRef, destDirectory string) {
	argsStr := fmt.Sprintf("podName=%s&sharingRef=%s&dirPath=%s", podName, sharingRef, destDirectory)
	data, err := fdfsAPI.getReq(apiFileReceive, argsStr)
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
}

func deleteFile(podName, fileNameWithPath string) {
	rmFileReq := common.FileSystemRequest{
		PodName:  podName,
		FilePath: fileNameWithPath,
	}
	jsonData, err := json.Marshal(rmFileReq)
	if err != nil {
		fmt.Println("rm file: error marshalling request")
		return
	}
	data, err := fdfsAPI.postReq(http.MethodDelete, apiFileDelete, jsonData)
	if err != nil {
		fmt.Println("rm failed: ", err)
		return
	}
	message := strings.ReplaceAll(string(data), "\n", "")
	fmt.Println(message)
}
