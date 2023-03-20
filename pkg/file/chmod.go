package file

import (
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Chmod does all the validation for the existence of the file and changes file mode
func (f *File) Chmod(podFileWithPath, podPassword string, mode uint32) error {
	// TODO check valid mode
	if f.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return feed.ErrReadOnlyFeed
	}
	// check if file present
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(podPassword, totalFilePath) {
		return ErrFileNotFound
	}

	meta := f.GetInode(podPassword, totalFilePath)
	if meta == nil { // skipcq: TCV-001
		return ErrFileNotFound
	}

	if meta.Mode == S_IFREG|mode {
		return nil
	}
	meta.Mode = S_IFREG | mode
	meta.AccessTime = time.Now().Unix()

	err := f.updateMeta(meta, podPassword)
	if err != nil {
		return err
	}
	f.AddToFileMap(totalFilePath, meta)
	return nil
}
