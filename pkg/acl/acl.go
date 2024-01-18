package acl

const (
	PermissionRead    = 1
	PermissionWrite   = 2
	PermissionExecute = 4
)

type ACL interface {
	CreateGroup(groupName, ownerAddress string) error
	AddMember(groupName, ownerAddress, memberAddress string, permission uint8) error
	RemoveMember(groupName, ownerAddress, memberAddress string) error
	RemoveGroup(groupName, ownerAddress string) error
	GetGroupMembers(groupName, ownerAddress string) (map[string]uint8, error)
	GetAllGroups(ownerAddress string) (map[string]map[string]uint8, error)
	UpdatePermission(groupName, ownerAddress, memberAddress string, permission uint8) error
	GetPermission(groupName, ownerAddress, memberAddress string) (uint8, error)
}
