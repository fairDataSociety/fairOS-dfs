package file

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Status does all the validation for the existence of the file and checks file sync status
func (f *File) Status(podFileWithPath, podPassword string) (int64, int64, int64, error) {
	// check if file present
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(totalFilePath) {
		return 0, 0, 0, ErrFileNotPresent
	}

	meta := f.GetFromFileMap(totalFilePath)
	if meta == nil { // skipcq: TCV-001
		return 0, 0, 0, ErrFileNotFound
	}

	return f.client.GetTag(meta.Tag)
}
