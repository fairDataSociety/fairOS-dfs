package pod

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/ethersphere/bee/v2/pkg/swarm"

	aclController "github.com/fairdatasociety/fairOS-dfs/pkg/acl/acl"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
)

func (g *Group) AddMember(groupName string, memberAddress common.Address, memberPublicKey *ecdsa.PublicKey, permission uint8) ([]byte, error) {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return nil, err
	}
	if !g.checkIfPodPresent(groups, groupName) {
		return nil, ErrGroupDoesNotExist
	}

	// encrypt mnemonic DH key secret with member's public key
	a, _ := memberPublicKey.Curve.ScalarMult(memberPublicKey.X, memberPublicKey.Y, g.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
	secret := sha256.Sum256(a.Bytes())

	var gr *GroupItem
	for _, group := range groups.Groups {
		if group.Name == groupName {
			gr = &group
			break
		}
	}
	if gr == nil {
		return nil, ErrGroupDoesNotExist
	}

	seed, err := utils.DecryptBytes(crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey()), gr.Secret)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	encData, err := utils.EncryptBytes(secret[:], seed)
	if err != nil {
		return nil, err
	}

	address := g.acc.GetUserAccountInfo().GetAddress()
	commonAddr := common.HexToAddress(address.Hex())
	addressStr := commonAddr.Hex()

	// store group info and share the reference
	group := &GroupItem{
		Name:           gr.Name,
		Secret:         encData,
		OwnerPublicKey: gr.OwnerPublicKey,
		OwnerAddress:   gr.OwnerAddress,
		Password:       gr.Password,
	}

	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	ref, err := g.client.UploadBlob(0, "", "0", false, false, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	err = g.acl.AddMember(groupName, addressStr, memberAddress.String(), permission)
	if err != nil {
		return nil, err
	}

	return ref.Bytes(), nil
}

func (g *Group) AcceptGroupInvite(ref []byte) error {
	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return err
	}

	// download blob
	r, _, err := g.client.DownloadBlob(swarm.NewAddress(ref))
	if err != nil {
		return err
	}

	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// unmarshall into GroupItem
	group := &GroupItem{}
	err = json.Unmarshal(data, group)
	if err != nil {
		return err
	}
	address := g.acc.GetUserAccountInfo().GetAddress()
	commonAddr := common.HexToAddress(address.Hex())
	addressStr := commonAddr.Hex()

	// check smart contract for permission
	perm, err := g.acl.GetPermission(group.Name, group.OwnerAddress, addressStr)
	if err != nil {
		return err
	}

	if perm != aclController.PermissionRead && perm != aclController.PermissionWrite {
		return ErrPermissionDenied
	}
	// Save in te groups list
	groups.SharedGroups = append(groups.SharedGroups, *group)

	// Save the groups list
	return g.store(groups)
}

func (g *Group) RemoveMember(groupName string, memberAddress common.Address) error {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return err
	}
	if !g.checkIfPodPresent(groups, groupName) {
		return ErrGroupDoesNotExist
	}
	address := g.acc.GetUserAccountInfo().GetAddress()
	addressStr := common.HexToAddress(address.Hex()).Hex()
	return g.acl.RemoveMember(groupName, addressStr, memberAddress.String())
}

func (g *Group) UpdatePermission(groupName string, memberAddress common.Address, permission uint8) error {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return err
	}
	if !g.checkIfPodPresent(groups, groupName) {
		return ErrGroupDoesNotExist
	}
	address := g.acc.GetUserAccountInfo().GetAddress()
	addressStr := common.HexToAddress(address.Hex()).Hex()
	return g.acl.UpdatePermission(groupName, addressStr, memberAddress.String(), permission)
}

func (g *Group) GetPermission(groupName string) (uint8, error) {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return 0, err
	}
	if !g.checkIfPodPresent(groups, groupName) {
		return 0, ErrGroupDoesNotExist
	}

	var gr *GroupItem
	for _, group := range groups.Groups {
		if group.Name == groupName {
			gr = &group
			break
		}
	}
	if gr == nil {
		for _, group := range groups.SharedGroups {
			if group.Name == groupName {
				gr = &group
				break
			}
		}
	}
	if gr == nil {
		return 0, ErrGroupDoesNotExist
	}

	address := g.acc.GetUserAccountInfo().GetAddress()
	addressStr := common.HexToAddress(address.Hex()).Hex()
	return g.acl.GetPermission(groupName, gr.OwnerAddress, addressStr)
}

func (g *Group) GetGroupMembers(groupName string) (map[string]uint8, error) {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return nil, err
	}
	if !g.checkIfPodPresent(groups, groupName) {
		return nil, ErrGroupDoesNotExist
	}
	address := g.acc.GetUserAccountInfo().GetAddress()
	addressStr := common.HexToAddress(address.Hex()).Hex()
	return g.acl.GetGroupMembers(groupName, addressStr)
}

//
//func (g *Group) GetAllGroups() (map[string]map[string]uint8, error) {
//	address := g.acc.GetUserAccountInfo().GetAddress()
//	addressStr := common.HexToAddress(address.Hex()).Hex()
//	return g.acl.GetAllGroups(addressStr)
//}
