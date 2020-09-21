/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package user

import (
	"encoding/json"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	avatarFeedName   = "Avatar"
	nameFeedName     = "Name"
	contactsFeedName = "Contacts"
)

type Name struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	SurName    string `json:"surname"`
}

type Contacts struct {
	Phone  string  `json:"phone_number"`
	Mobile string  `json:"mobile"`
	Addr   Address `json:"address"`
}

type Address struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	State        string `json:"state/Province/Region"`
	ZipCode      string `json:"zip_code"`
}

func (u *Users) SaveName(firstName, lastName, middleName, surName string, userInfo *Info) error {
	rootAddress := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	data, err := getFeedData(nameFeedName, rootAddress, userInfo.GetFeed())
	if err != nil {
		return err
	}
	name := &Name{}
	err = json.Unmarshal(data, name)
	if err != nil {
		return err
	}
	if firstName != "" {
		name.FirstName = firstName
	}
	if lastName != "" {
		name.LastName = lastName
	}
	if middleName != "" {
		name.MiddleName = middleName
	}
	if surName != "" {
		name.SurName = surName
	}

	nameData, err := json.Marshal(name)
	if err != nil {
		return err
	}
	return putFeedData(nameFeedName, rootAddress, nameData, userInfo.GetFeed())
}

func (u *Users) GetName(userInfo *Info) (*Name, error) {
	rootAddress := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	data, err := getFeedData(nameFeedName, rootAddress, userInfo.GetFeed())
	if err != nil {
		return nil, err
	}
	name := &Name{}
	err = json.Unmarshal(data, name)
	if err != nil {
		return nil, err
	}
	return name, nil
}

func (u *Users) SaveContacts(phone, mobile string, address *Address, userInfo *Info) error {
	rootAddress := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	data, err := getFeedData(contactsFeedName, rootAddress, userInfo.GetFeed())
	if err != nil {
		return err
	}
	contacts := &Contacts{}
	err = json.Unmarshal(data, contacts)
	if err != nil {
		return err
	}

	if phone != "" {
		contacts.Phone = phone
	}
	if mobile != "" {
		contacts.Mobile = mobile
	}
	if address != nil {
		contacts.Addr.AddressLine1 = address.AddressLine1
		contacts.Addr.AddressLine2 = address.AddressLine2
		contacts.Addr.State = address.State
		contacts.Addr.ZipCode = address.ZipCode
	}
	contactData, err := json.Marshal(contacts)
	if err != nil {
		return err
	}
	return putFeedData(contactsFeedName, rootAddress, contactData, userInfo.GetFeed())
}

func (u *Users) GetContacts(userInfo *Info) (*Contacts, error) {
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	data, err := getFeedData(contactsFeedName, rootReference, userInfo.GetFeed())
	if err != nil {
		return nil, err
	}
	contacts := &Contacts{}
	err = json.Unmarshal(data, contacts)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (u *Users) SaveAvatar(avatar []byte, userInfo *Info) error {
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	return putFeedData(avatarFeedName, rootReference, avatar, userInfo.GetFeed())
}

func (u *Users) GetAvatar(userInfo *Info) ([]byte, error) {
	rootReference := userInfo.GetAccount().GetAddress(account.UserAccountIndex)
	return getFeedData(avatarFeedName, rootReference, userInfo.GetFeed())
}

func getFeedData(fileName string, rootReference utils.Address, fd *feed.API) ([]byte, error) {
	topic := utils.HashString(fileName)
	_, data, err := fd.GetFeedData(topic, rootReference)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func putFeedData(fileName string, rootReference utils.Address, data []byte, fd *feed.API) error {
	topic := utils.HashString(fileName)
	_, err := fd.UpdateFeed(topic, rootReference, data)
	if err != nil {
		return err
	}
	return nil
}
