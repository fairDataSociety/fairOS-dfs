package dir

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Chmod does all the validation for the existence of the file and changes file mode
func (d *Directory) Chmod(dirNameWithPath, podPassword string, mode uint32) error {
	topic := utils.HashString(dirNameWithPath)
	_, data, err := d.fd.GetFeedData(topic, d.getAddress(), []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		return fmt.Errorf("dir chmod: %v", err)
	}
	if string(data) == utils.DeletedFeedMagicWord {
		return ErrDirectoryNotPresent
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
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
	_, err = d.fd.UpdateFeed(topic, d.userAddress, metaBytes, []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		return err
	}
	d.AddToDirectoryMap(dirNameWithPath, &dirInode)
	return nil
}
