package pod

import (
	"strings"

	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// PodFork creates a new pod with all the contents of a given pod
func (p *Pod) PodFork(podName, forkName string) error {
	podName, err := CleanPodName(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	forkName, err = CleanPodName(forkName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if !p.IsPodOpened(podName) {
		return ErrPodNotOpened
	}

	if !p.IsPodOpened(forkName) {
		return ErrPodNotOpened
	}

	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	forkInfo, _, err := p.GetPodInfoFromPodMap(forkName)
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
			meta := source.GetFile().GetFromFileMap(filePath)

			r, _, err := source.GetFile().Download(filePath, source.GetPodPassword())
			if err != nil { // skipcq: TCV-001
				return err
			}

			err = dst.GetFile().Upload(r, meta.Name, int64(meta.Size), meta.BlockSize, meta.Path, meta.Compression, dst.GetPodPassword())
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
			iNode := source.GetDirectory().GetDirFromDirectoryMap(path)
			err := dst.GetDirectory().MkDir(path, dst.GetPodPassword())
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
