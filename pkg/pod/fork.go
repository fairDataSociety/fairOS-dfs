package pod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// PodFork forks a pod with a different given name
func (p *Pod) PodFork(podName, forkName string) error {
	podName, err := CleanPodName(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	podInfo, _, err := p.GetPodInfo(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	forkName, err = CleanPodName(forkName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	forkInfo, _, err := p.GetPodInfo(forkName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	directory := podInfo.GetDirectory()
	rootInode := directory.GetInode(podInfo.GetPodPassword(), "/")
	if rootInode == nil {
		return fmt.Errorf("root inode not found")
	}
	return cloneFolder(podInfo, forkInfo, "/", rootInode)
}

// PodForkFromRef forks a pof from a given pod sharing reference
func (p *Pod) PodForkFromRef(forkName, refString string) error {
	ref, err := utils.ParseHexReference(refString)
	if err != nil {
		return nil
	}
	data, resp, err := p.client.DownloadBlob(ref.Bytes())
	if err != nil { // skipcq: TCV-001
		return err
	}
	if resp != http.StatusOK { // skipcq: TCV-001
		return fmt.Errorf("ReceivePod: could not download blob")
	}
	var shareInfo ShareInfo
	err = json.Unmarshal(data, &shareInfo)
	if err != nil { // skipcq: TCV-001
		return err
	}
	accountInfo := p.acc.GetEmptyAccountInfo()
	address := utils.HexToAddress(shareInfo.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, p.client, p.feedCacheSize, p.feedCacheTTL, p.logger)
	file := f.NewFile(shareInfo.PodName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
	dir := d.NewDirectory(shareInfo.PodName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)
	podInfo := &Info{
		podName:     shareInfo.PodName,
		podPassword: shareInfo.Password,
		userAddress: address,
		dir:         dir,
		file:        file,
		accountInfo: accountInfo,
		feed:        fd,
	}

	return p.forkPod(podInfo, forkName)
}

// PodForkCore creates a new pod with all the contents of a given pod
func (p *Pod) forkPod(podInfo *Info, forkName string) error {
	err := podInfo.GetDirectory().SyncDirectory("/", podInfo.GetPodPassword())
	if err != nil {
		return err
	}

	forkName, err = CleanPodName(forkName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	forkInfo, _, err := p.GetPodInfo(forkName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// sync from the root directory
	rootInode := podInfo.GetDirectory().GetDirFromDirectoryMap("/")
	return cloneFolder(podInfo, forkInfo, "/", rootInode)
}

func cloneFolder(source, dst *Info, dirNameWithPath string, dirInode *d.Inode) error {
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			meta := source.GetFile().GetInode(source.GetPodPassword(), filePath)
			r, _, err := source.GetFile().Download(filePath, source.GetPodPassword())
			if err != nil { // skipcq: TCV-001
				return err
			}

			err = dst.GetFile().Upload(r, meta.Name, int64(meta.Size), meta.BlockSize, 0, meta.Path, meta.Compression, dst.GetPodPassword())
			if err != nil { // skipcq: TCV-001
				return err
			}

			err = dst.GetDirectory().AddEntryToDir(dirNameWithPath, dst.GetPodPassword(), fileName, true)
			if err != nil { // skipcq: TCV-001
				return err
			}
		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			path := utils.CombinePathAndFile(dirNameWithPath, dirName)
			iNode := source.GetDirectory().GetInode(source.GetPodPassword(), path)
			err := dst.GetDirectory().MkDir(path, dst.GetPodPassword(), 0)
			if err != nil { // skipcq: TCV-001
				return err
			}
			err = cloneFolder(source, dst, path, iNode)
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
	}
	return nil
}
