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

package pod

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) UploadFile(podName, fileName string, fileSize int64, fd io.Reader, podDir, blockSize, compression string) (string, error) {
	if !p.isPodOpened(podName) {
		return "", fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return "", err
	}

	if podInfo.accountInfo.IsReadOnlyPod() {
		return "", ErrReadOnlyPod
	}

	dir := podInfo.GetDirectory()

	bs, err := humanize.ParseBytes(blockSize)
	if err != nil {
		return "", err
	}

	path := p.getFilePath(podDir, podInfo)

	_, dirInode, err := dir.GetDirNode(path, podInfo.GetFeed(), podInfo.GetAccountInfo())
	if err != nil {
		return "", err
	}

	fpath := path + utils.PathSeperator + fileName
	if podInfo.file.IsFileAlreadyPResent(fpath) {
		return "", fmt.Errorf("file already present in the destination dir")
	}
	ref, err := podInfo.file.Upload(fd, fileName, fileSize, uint32(bs), fpath, compression)
	if err != nil {
		return "", err
	}
	dirInode.Hashes = append(dirInode.Hashes, ref)

	dirInode.Meta.ModificationTime = time.Now().Unix()
	topic, err := dir.UpdateDirectory(dirInode)
	if err != nil {
		return "", err
	}

	if path != podInfo.GetCurrentPodPathAndName() {
		err = p.UpdateTillThePod(podName, podInfo.GetDirectory(), topic, path, true)
		if err != nil {
			return "", err
		}
	}

	return utils.NewReference(ref).String(), nil
}

func (p *Pod) getFilePath(podDir string, podInfo *Info) string {
	var path string
	if podDir == utils.PathSeperator || podDir == podInfo.GetCurrentPodPathAndName() {
		return podInfo.GetCurrentPodPathAndName()
	}

	// this is a full path.. so use it as it is
	if strings.HasPrefix(podDir, "/") {
		return podInfo.GetCurrentPodPathAndName() + podDir
	}

	if podInfo.IsCurrentDirRoot() {
		if podDir == "." {
			path = podInfo.GetCurrentPodPathAndName()
		} else {
			path = podInfo.GetCurrentPodPathAndName() + utils.PathSeperator + podDir
		}
	} else {
		if podDir == "." {
			path = podInfo.GetCurrentDirPathAndName()
		} else {
			path = podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + podDir
		}
	}
	return path
}
