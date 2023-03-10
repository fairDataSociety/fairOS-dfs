package file

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Status does all the validation for the existence of the file and checks file sync status
func (f *File) Status(podFileWithPath, podPassword string) (int64, int64, int64, error) {
	// check if file present
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(podPassword, totalFilePath) {
		return 0, 0, 0, ErrFileNotFound
	}

	tag := f.LoadFromTagMap(totalFilePath)
	if tag == 0 { // skipcq: TCV-001
		return 0, 0, 0, ErrFileTagPresent
	}

	return f.client.GetTag(tag)
}
