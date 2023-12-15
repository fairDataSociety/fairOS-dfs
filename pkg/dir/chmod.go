package dir

import (
	"fmt"
	"time"
)

// Chmod does all the validation for the existence of the file and changes file mode
func (d *Directory) Chmod(dirNameWithPath, podPassword string, mode uint32) error {
	dirInode, err := d.GetInode(podPassword, dirNameWithPath)
	if err != nil { // skipcq: TCV-001
		return fmt.Errorf("dir chmod: %v", err)
	}

	if dirInode.Meta == nil && dirInode.FileOrDirNames == nil { // skipcq: TCV-001
		return ErrDirectoryNotPresent
	}

	dirInode.Meta.Mode = S_IFDIR | mode
	dirInode.Meta.AccessTime = time.Now().Unix()
	return d.SetInode(podPassword, dirInode)
}
