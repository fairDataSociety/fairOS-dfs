package act

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethersphere/bee/v2/pkg/api"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	actFile = "ACTs"
)

type List map[string]*Act

// Act represents an Access Control Trie (ACT) with its metadata, grantees, and associated content.
type Act struct {
	Name        string     `json:"name"`
	HistoryRef  []byte     `json:"historyRef"`
	GranteesRef []byte     `json:"granteesRef"`
	CreatedAt   time.Time  `json:"createdAt"`
	Content     []*Content `json:"content"`
}

// Content represents a pod or data reference associated with the ACT.
type Content struct {
	Reference      string        `json:"reference"`
	Topic          []byte        `json:"topic"`
	Owner          utils.Address `json:"owner"`
	OwnerPublicKey string        `json:"ownerPublicKey"`
	AddedAt        time.Time     `json:"addedAt"`
}

func (t *ACT) CreateUpdateACT(actName string, publicKeyGrant, publicKeyRevoke *ecdsa.PublicKey) (*Act, error) {
	if actName == "" {
		return nil, fmt.Errorf("act name is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	list, err := t.loadUserACTs()
	if err != nil {
		return nil, err
	}
	if publicKeyGrant != nil {
		gkBytes := crypto.EncodeSecp256k1PublicKey(publicKeyGrant)
		gkParsedPublicKey, err := btcec.ParsePubKey(gkBytes)
		if err != nil {
			return nil, err
		}
		publicKeyGrant = gkParsedPublicKey.ToECDSA()
	}
	if publicKeyRevoke != nil {
		rkBytes := crypto.EncodeSecp256k1PublicKey(publicKeyRevoke)
		rkParsedPublicKey, err := btcec.ParsePubKey(rkBytes)
		if err != nil {
			return nil, err
		}
		publicKeyRevoke = rkParsedPublicKey.ToECDSA()
	}

	var (
		resp       = &api.GranteesPostResponse{}
		grantList  []*ecdsa.PublicKey
		revokeList []*ecdsa.PublicKey
		owner      = t.acc.GetUserAccountInfo().GetAddress()
		topic      = fmt.Sprintf("%s-%s", actName, owner.String())
		topicBytes = utils.HashString(topic)
	)
	// check if act with name already exists
	act, ok := list[actName]
	if !ok {
		act = &Act{
			Name:        actName,
			CreatedAt:   time.Now(),
			HistoryRef:  swarm.ZeroAddress.Bytes(),
			GranteesRef: swarm.ZeroAddress.Bytes(),
			Content:     []*Content{},
		}
		grantList = []*ecdsa.PublicKey{publicKeyGrant}
		resp, err = t.act.CreateGrantee(context.Background(), swarm.NewAddress(act.HistoryRef), grantList)
		if err != nil {
			return nil, err
		}
		err = t.fd.CreateFeed(owner, topicBytes, resp.HistoryReference.Bytes(), nil)
		if err != nil {
			return nil, err
		}
	} else {
		if publicKeyGrant == nil && publicKeyRevoke == nil {
			return nil, fmt.Errorf("grant or revoke key is required")
		}
		if publicKeyGrant != nil {
			grantList = []*ecdsa.PublicKey{publicKeyGrant}
		} else {
			grantList = nil
		}
		if publicKeyRevoke != nil {
			revokeList = []*ecdsa.PublicKey{publicKeyRevoke}
		} else {
			revokeList = nil
		}

		resp, err = t.act.RevokeGrant(context.Background(), swarm.NewAddress(act.GranteesRef), swarm.NewAddress(act.HistoryRef), grantList, revokeList)
		if err != nil {
			return nil, err
		}
		err = t.fd.UpdateFeed(owner, topicBytes, resp.HistoryReference.Bytes(), nil, false)
		if err != nil {
			return nil, err
		}
	}

	act.GranteesRef = resp.Reference.Bytes()
	act.HistoryRef = resp.HistoryReference.Bytes()

	list[actName] = act
	err = t.storeUserACTs(list)
	if err != nil {
		return nil, err
	}
	return act, nil
}

func (t *ACT) GetACT(actName string) (*Act, error) {
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

	return act, nil
}

func (t *ACT) GetList() (List, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.loadUserACTs()
}
