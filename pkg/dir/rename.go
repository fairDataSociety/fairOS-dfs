package dir

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) RenameDir(dirWithPath, newName string) error {
	parentPath := filepath.Dir(dirWithPath)
	dirName := filepath.Base(dirWithPath)

	// validation checks of the arguments
	if dirName == "" || strings.HasPrefix(dirName, utils.PathSeparator) {
		return ErrInvalidDirectoryName
	}

	if len(dirName) > nameLength {
		return ErrTooLongDirectoryName
	}

	if dirName == "/" {
		return fmt.Errorf("cannot rename root dir")
	}

	// check if directory already present
	totalPath := utils.CombinePathAndFile(parentPath, dirName)
	newTotalPath := utils.CombinePathAndFile(parentPath, newName)

	// check if parent path exists
	if d.GetDirFromDirectoryMap(parentPath) == nil {
		return ErrDirectoryNotPresent
	}

	if d.GetDirFromDirectoryMap(newTotalPath) != nil {
		return ErrDirectoryAlreadyPresent
	}

	err := d.mapChildrenToNewPath(totalPath, newTotalPath)
	if err != nil {
		return err
	}

	topic := utils.HashString(utils.CombinePathAndFile(totalPath, ""))
	newTopic := utils.HashString(utils.CombinePathAndFile(newTotalPath, ""))
	_, inodeData, err := d.fd.GetFeedData(topic, d.userAddress)
	if err != nil {
		return err
	}

	// unmarshall the data and rename the directory entry
	var inode *Inode
	err = json.Unmarshal(inodeData, &inode)
	if err != nil { // skipcq: TCV-001
		return err
	}
	inode.Meta.Name = newName
	inode.Meta.ModificationTime = time.Now().Unix()
	// upload meta
	fileMetaBytes, err := json.Marshal(inode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	_, err = d.fd.CreateFeed(newTopic, d.userAddress, fileMetaBytes)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// delete old meta
	// update with utils.DeletedFeedMagicWord
	_, err = d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil { // skipcq: TCV-001
		return err
	}
	err = d.fd.DeleteFeed(topic, d.userAddress)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// get the parent directory entry and add this new directory to its list of children
	err = d.RemoveEntryFromDir(parentPath, dirName, false)
	if err != nil {
		return err
	}
	err = d.AddEntryToDir(parentPath, newName, false)
	if err != nil {
		return err
	}

	return d.SyncDirectory(parentPath)
}

func (d *Directory) mapChildrenToNewPath(totalPath, newTotalPath string) error {
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(totalPath, fileName)
			newFilePath := utils.CombinePathAndFile(newTotalPath, fileName)
			topic := utils.HashString(filePath)
			_, metaBytes, err := d.fd.GetFeedData(topic, d.userAddress)
			if err != nil {
				return err
			}
			if string(metaBytes) == utils.DeletedFeedMagicWord {
				continue
			}

			p := &file.MetaData{}
			err = json.Unmarshal(metaBytes, p)
			if err != nil { // skipcq: TCV-001
				return err
			}
			newTopic := utils.HashString(newFilePath)
			// change previous meta.Name
			p.Path = newTotalPath
			p.ModificationTime = time.Now().Unix()
			// upload meta
			fileMetaBytes, err := json.Marshal(p)
			if err != nil { // skipcq: TCV-001
				return err
			}

			_, err = d.fd.CreateFeed(newTopic, d.userAddress, fileMetaBytes)
			if err != nil { // skipcq: TCV-001
				return err
			}

			// delete old meta
			// update with utils.DeletedFeedMagicWord
			_, err = d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
			if err != nil { // skipcq: TCV-001
				return err
			}
			err = d.fd.DeleteFeed(topic, d.userAddress)
			if err != nil { // skipcq: TCV-001
				return err
			}
		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			pathWithFile := utils.CombinePathAndFile(totalPath, dirName)
			newPathWithFile := utils.CombinePathAndFile(newTotalPath, dirName)
			err := d.mapChildrenToNewPath(pathWithFile, newPathWithFile)
			if err != nil { // skipcq: TCV-001
				return err
			}
			topic := utils.HashString(utils.CombinePathAndFile(pathWithFile, ""))
			newTopic := utils.HashString(utils.CombinePathAndFile(newPathWithFile, ""))
			_, inodeData, err := d.fd.GetFeedData(topic, d.userAddress)
			if err != nil {
				return err
			}

			// unmarshall the data and add the directory entry to the parent
			var inode *Inode
			err = json.Unmarshal(inodeData, &inode)
			if err != nil { // skipcq: TCV-001
				return err
			}
			inode.Meta.Path = newTotalPath
			inode.Meta.ModificationTime = time.Now().Unix()
			// upload meta
			fileMetaBytes, err := json.Marshal(inode)
			if err != nil { // skipcq: TCV-001
				return err
			}

			_, err = d.fd.CreateFeed(newTopic, d.userAddress, fileMetaBytes)
			if err != nil { // skipcq: TCV-001
				return err
			}

			// delete old meta
			// update with utils.DeletedFeedMagicWord
			_, err = d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
			if err != nil { // skipcq: TCV-001
				return err
			}
			err = d.fd.DeleteFeed(topic, d.userAddress)
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
	}
	return nil
}
