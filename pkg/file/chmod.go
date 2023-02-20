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
	if !f.IsFileAlreadyPresent(totalFilePath) {
		return ErrFileNotPresent
	}

	meta := f.GetFromFileMap(totalFilePath)
	if meta == nil { // skipcq: TCV-001
		return ErrFileNotFound
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
