package dir

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
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
	metaBytes, err := json.Marshal(dirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}
	err = d.file.Upload(bufio.NewReader(bytes.NewBuffer(metaBytes)), indexFileName, int64(len(metaBytes)), file.MinBlockSize, 0, dirNameWithPath, "gzip", podPassword)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(dirNameWithPath, dirInode)
	return nil
}
