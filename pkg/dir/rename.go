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

// RenameDir renames a directory
func (d *Directory) RenameDir(dirNameWithPath, newDirNameWithPath, podPassword string) error {
	dirNameWithPath = filepath.ToSlash(dirNameWithPath)
	newDirNameWithPath = filepath.ToSlash(newDirNameWithPath)
	parentPath := filepath.ToSlash(filepath.Dir(dirNameWithPath))
	dirName := filepath.Base(dirNameWithPath)

	newParentPath := filepath.ToSlash(filepath.Dir(newDirNameWithPath))
	newDirName := filepath.Base(newDirNameWithPath)

	// validation checks of the arguments
	if dirName == "" || strings.HasPrefix(dirName, utils.PathSeparator) { // skipcq: TCV-001
		return ErrInvalidDirectoryName
	}

	if len(dirName) > nameLength { // skipcq: TCV-001
		return ErrTooLongDirectoryName
	}

	if dirName == "/" {
		return fmt.Errorf("cannot rename root dir")
	}

	// check if directory exists
	if d.GetInode(podPassword, dirNameWithPath) == nil { // skipcq: TCV-001
		return ErrDirectoryNotPresent
	}

	// check if parent directory exists
	if d.GetInode(podPassword, parentPath) == nil { // skipcq: TCV-001
		return ErrDirectoryNotPresent
	}
	if d.GetInode(podPassword, newDirNameWithPath) != nil {
		return ErrDirectoryAlreadyPresent
	}

	err := d.mapChildrenToNewPath(dirNameWithPath, newDirNameWithPath, podPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}

	topic := utils.HashString(dirNameWithPath)
	newTopic := utils.HashString(newDirNameWithPath)
	_, inodeData, err := d.fd.GetFeedData(topic, d.userAddress, []byte(podPassword))
	if err != nil {
		return err
	}

	// unmarshall the data and rename the directory entry
	var inode *Inode
	err = json.Unmarshal(inodeData, &inode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	inode.Meta.Name = newDirName
	inode.Meta.Path = newParentPath
	inode.Meta.ModificationTime = time.Now().Unix()

	// upload meta
	fileMetaBytes, err := json.Marshal(inode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	previousAddr, _, err := d.fd.GetFeedData(newTopic, d.userAddress, []byte(podPassword))
	if err == nil && previousAddr != nil {
		_, err = d.fd.UpdateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
		if err != nil { // skipcq: TCV-001
			return err
		}
	} else {
		_, err = d.fd.CreateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
		if err != nil { // skipcq: TCV-001
			return err
		}
	}

	// delete old meta
	// update with utils.DeletedFeedMagicWord
	_, err = d.fd.UpdateFeed(d.userAddress, topic, []byte(utils.DeletedFeedMagicWord), []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		return err
	}
	d.RemoveFromDirectoryMap(dirNameWithPath)

	// get the parent directory entry and add this new directory to its list of children
	err = d.RemoveEntryFromDir(parentPath, podPassword, dirName, false)
	if err != nil {
		return err
	}
	err = d.AddEntryToDir(newParentPath, podPassword, newDirName, false)
	if err != nil {
		return err
	}

	err = d.SyncDirectory(parentPath, podPassword)
	if err != nil {
		return err
	}

	if parentPath != newParentPath {
		err = d.SyncDirectory(newParentPath, podPassword)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Directory) mapChildrenToNewPath(totalPath, newTotalPath, podPassword string) error {
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(totalPath, fileName)
			newFilePath := utils.CombinePathAndFile(newTotalPath, fileName)
			topic := utils.HashString(filePath)
			_, metaBytes, err := d.fd.GetFeedData(topic, d.userAddress, []byte(podPassword))
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

			previousAddr, _, err := d.fd.GetFeedData(newTopic, d.userAddress, []byte(podPassword))
			if err == nil && previousAddr != nil {
				_, err = d.fd.UpdateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
				if err != nil { // skipcq: TCV-001
					return err
				}
			} else {
				_, err = d.fd.CreateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
				if err != nil { // skipcq: TCV-001
					return err
				}
			}

			// delete old meta
			// update with utils.DeletedFeedMagicWord
			_, err = d.fd.UpdateFeed(d.userAddress, topic, []byte(utils.DeletedFeedMagicWord), []byte(podPassword))
			if err != nil { // skipcq: TCV-001
				return err
			}
		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			pathWithDir := utils.CombinePathAndFile(totalPath, dirName)
			newPathWithDir := utils.CombinePathAndFile(newTotalPath, dirName)
			err := d.mapChildrenToNewPath(pathWithDir, newPathWithDir, podPassword)
			if err != nil { // skipcq: TCV-001
				return err
			}
			topic := utils.HashString(pathWithDir)
			newTopic := utils.HashString(newPathWithDir)
			_, inodeData, err := d.fd.GetFeedData(topic, d.userAddress, []byte(podPassword))
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
			previousAddr, _, err := d.fd.GetFeedData(newTopic, d.userAddress, []byte(podPassword))
			if err == nil && previousAddr != nil {
				_, err = d.fd.UpdateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
				if err != nil { // skipcq: TCV-001
					return err
				}
			} else {
				_, err = d.fd.CreateFeed(d.userAddress, newTopic, fileMetaBytes, []byte(podPassword))
				if err != nil { // skipcq: TCV-001
					return err
				}
			}

			// delete old meta
			// update with utils.DeletedFeedMagicWord
			_, err = d.fd.UpdateFeed(d.userAddress, topic, []byte(utils.DeletedFeedMagicWord), []byte(podPassword))
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
	}
	return nil
}
