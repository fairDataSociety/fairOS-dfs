package acl

import "github.com/ethereum/go-ethereum/ethclient"

type ACL struct {
	c *ethclient.Client
}

func (A ACL) CreateGroup(groupName, ownerAddress string) error {
	//TODO implement me
	panic("implement me")
}

func (A ACL) AddMember(groupName, ownerAddress, memberAddress string) error {
	//TODO implement me
	panic("implement me")
}

func (A ACL) RemoveMember(groupName, ownerAddress, memberAddress string) error {
	//TODO implement me
	panic("implement me")
}

func (A ACL) RemoveGroup(groupName, ownerAddress string) error {
	//TODO implement me
	panic("implement me")
}

func (A ACL) GetGroupMembers(groupName string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (A ACL) GetAllGroups(ownerAddress string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (A ACL) UpdateRole(groupName, ownerAddress, memberAddress, role string) error {
	//TODO implement me
	panic("implement me")
}

func NewACL(c *ethclient.Client) *ACL {
	return &ACL{
		c: c,
	}
}
