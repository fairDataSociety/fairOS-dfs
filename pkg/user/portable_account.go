package user

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	AuthVersion = "FDP-login-v1.0"
)

func (u *Users) uploadPortableAccount(accountInfo *account.Info, username, password string, data []byte, fd *feed.API) error {
	topic := utils.HashString(AuthVersion + username + password)
	_, err := fd.CreateFeedFromTopic(topic, accountInfo.GetAddress(), data)
	if err != nil {
		return err
	}
	return nil
}

func (u *Users) downloadPortableAccount(address utils.Address, username, password string, fd *feed.API) ([]byte, error) {
	topic := utils.HashString(AuthVersion + username + password)
	_, data, err := fd.GetFeedDataFromTopic(topic, address)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *Users) deletePortableAccount(address utils.Address, username, password string, fd *feed.API) error {
	topic := utils.HashString(AuthVersion + username + password)
	return fd.DeleteFeed(topic, address)
}
