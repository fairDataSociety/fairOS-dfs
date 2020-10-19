package collection

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	collectionFile = "collections"
)

func LoadUserCollections(fd *feed.API, user utils.Address) (map[string]string, error) {
	collections := make(map[string]string)
	topic := utils.HashString(collectionFile)
	_, data, err := fd.GetFeedData(topic, user)
	if err != nil {
		if err.Error() != "no feed updates found" {
			return collections, err
		}
	}

	buf := bytes.NewBuffer(data)
	rd := bufio.NewReader(buf)
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("loading collections: %w", err)
		}
		line = strings.Trim(line, "\n")
		lines := strings.Split(line, ",")
		collections[lines[0]] = lines[1]
	}
	return collections, nil
}

func StoreUserCollections(collections map[string]string, fd *feed.API, user utils.Address) error {
	buf := bytes.NewBuffer(nil)
	collectionLen := len(collections)
	for k, v := range collections {
		v := strings.Trim(v, "\n")
		if collectionLen > 1 && v == "" {
			continue
		}
		line := fmt.Sprintf("%s,%s", k, v)
		buf.WriteString(line + "\n")
	}

	topic := utils.HashString(collectionFile)
	_, err := fd.UpdateFeed(topic, user, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
