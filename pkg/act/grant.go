package act

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (t *ACT) GrantAccess(actName string, address swarm.Address) (*Content, error) {
	if actName == "" {
		return nil, fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return nil, err
	}
	// check if act with name already exists
	act, ok := list[actName]
	if !ok {
		return nil, ErrACTDoesNowExist
	}
	owner := t.acc.GetUserAccountInfo().GetAddress()
	addr, err := swarm.ParseHexAddress(act.HistoryRef)
	if err != nil {
		return nil, err
	}
	uploadResp, err := t.act.HandleUpload(context.Background(), address, addr)
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("%s-%s", actName, owner.String())
	topicBytes := utils.HashString(topic)
	err = t.fd.UpdateFeed(owner, topicBytes, uploadResp.HistoryReference.Bytes(), nil, false)
	if err != nil {
		return nil, err
	}
	opk, err := encodeKey(t.acc.GetUserAccountInfo().GetPublicKey())
	if err != nil {
		return nil, err
	}
	content := &Content{
		Reference:      uploadResp.Reference.String(),
		Topic:          topicBytes,
		Owner:          owner,
		OwnerPublicKey: opk,
		AddedAt:        time.Now(),
	}
	act.Content = append(act.Content, content)
	list[actName] = act
	err = t.storeUserACTs(list)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (t *ACT) GetPodAccess(actName string) (swarm.Address, error) {
	if actName == "" {
		return swarm.ZeroAddress, fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return swarm.ZeroAddress, err
	}
	// check if act with name already exists
	act, ok := list[actName]
	if !ok {
		return swarm.ZeroAddress, ErrACTDoesNowExist
	}
	if len(act.Content) == 0 {
		return swarm.ZeroAddress, ErrACTDoesNowExist
	}

	_, href, err := t.fd.GetFeedData(act.Content[0].Topic, act.Content[0].Owner, nil, false)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	ownerPubKey, err := parseKey(act.Content[0].OwnerPublicKey)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	reference, err := swarm.ParseHexAddress(act.Content[0].Reference)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	return t.act.HandleDownload(context.Background(), reference, swarm.NewAddress(href), ownerPubKey, time.Now().Unix())
}

func (t *ACT) SaveGrantedPod(actName string, c *Content) error {
	if actName == "" {
		return fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return err
	}
	// check if act with name already exists
	_, ok := list[actName]
	if ok {
		return ErrACTAlreadyExists
	}
	reference, err := swarm.ParseHexAddress(c.Reference)
	if err != nil {
		return err
	}
	a := &Act{
		Name:        actName,
		CreatedAt:   c.AddedAt,
		HistoryRef:  swarm.ZeroAddress.String(),
		GranteesRef: reference.String(),
		Content:     []*Content{c},
	}
	list[actName] = a
	return t.storeUserACTs(list)
}

func (t *ACT) GetGrantees(actName string) ([]string, error) {
	if actName == "" {
		return nil, fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return nil, err
	}
	act, ok := list[actName]
	if !ok {
		return nil, ErrACTDoesNowExist
	}
	addr, err := swarm.ParseHexAddress(act.GranteesRef)
	if err != nil {
		return nil, err
	}
	grantees, err := t.act.GetGrantees(context.Background(), addr)
	if err != nil {
		return nil, err
	}
	return encodeKeys(grantees)
}

func (t *ACT) GetContentList(actName string) ([]*Content, error) {
	if actName == "" {
		return nil, fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return nil, err
	}
	act, ok := list[actName]
	if !ok {
		return nil, ErrACTDoesNowExist
	}
	return act.Content, nil
}

func encodeKeys(keys []*ecdsa.PublicKey) ([]string, error) {
	encodedList := make([]string, 0, len(keys))
	for _, key := range keys {
		encoded, err := encodeKey(key)
		if err != nil {
			return nil, err
		}
		encodedList = append(encodedList, encoded)
	}
	return encodedList, nil
}

func encodeKey(key *ecdsa.PublicKey) (string, error) {
	if key == nil {
		return "", fmt.Errorf("nil key found")
	}
	return hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(key)), nil
}

func parseKey(g string) (*ecdsa.PublicKey, error) {
	h, err := hex.DecodeString(g)
	if err != nil {
		return nil, fmt.Errorf("failed to decode grantee: %w", err)
	}
	k, err := btcec.ParsePubKey(h)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grantee public key: %w", err)
	}
	return k.ToECDSA(), nil
}
