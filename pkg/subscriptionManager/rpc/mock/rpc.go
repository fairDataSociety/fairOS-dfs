package mock

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	goens "github.com/wealdtech/go-ens/v3"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"

	swarmMail "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/smail"

	"github.com/ethereum/go-ethereum/common"
)

type SubscriptionManager struct {
	lock            sync.Mutex
	listMap         map[string]*swarmMail.SwarmMailSub
	subscriptionMap map[string]*swarmMail.SwarmMailSubItem
	requestMap      map[string]*swarmMail.SwarmMailSubRequest
	subPodInfo      map[string]*rpc.SubscriptionItemInfo
	subscribedMap   map[string][]byte
}

func (s *SubscriptionManager) GetSubscribablePodInfo(subHash [32]byte) (*rpc.SubscriptionItemInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.subPodInfo[utils.Encode(subHash[:])], nil
}

// NewMockSubscriptionManager returns a new mock subscriptionManager manager client
func NewMockSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		listMap:         make(map[string]*swarmMail.SwarmMailSub),
		subscriptionMap: make(map[string]*swarmMail.SwarmMailSubItem),
		requestMap:      make(map[string]*swarmMail.SwarmMailSubRequest),
		subPodInfo:      make(map[string]*rpc.SubscriptionItemInfo),
		subscribedMap:   make(map[string][]byte),
	}
}

func (s *SubscriptionManager) AddPodToMarketplace(podAddress, owner common.Address, pod, title, desc, thumbnail string, price uint64, category, nameHash [32]byte, key *ecdsa.PrivateKey) error {
	subHash, err := goens.NameHash(owner.Hex() + podAddress.String())
	if err != nil {
		return err
	}
	i := &swarmMail.SwarmMailSub{
		SubHash:           subHash,
		FdpSellerNameHash: nameHash,
		Seller:            owner,
		SwarmLocation:     [32]byte{},
		Price:             new(big.Int).SetUint64(price),
		Active:            true,
		Earned:            nil,
		Bids:              0,
		Sells:             0,
		Reports:           0,
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	s.listMap[utils.Encode(subHash[:])] = i
	s.subPodInfo[utils.Encode(subHash[:])] = &rpc.SubscriptionItemInfo{
		PodName:    pod,
		PodAddress: podAddress.Hex(),
	}

	return nil
}

func (s *SubscriptionManager) HidePodFromMarketplace(owner common.Address, subHash [32]byte, hide bool, key *ecdsa.PrivateKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[utils.Encode(subHash[:])]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if i.Seller != owner {
		return fmt.Errorf("not the owner")
	}
	i.Active = !hide
	return nil
}

func (s *SubscriptionManager) RequestAccess(subscriber common.Address, subHash, nameHash [32]byte, key *ecdsa.PrivateKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[utils.Encode(subHash[:])]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if !i.Active {
		return fmt.Errorf("pod not listed")
	}
	reqHash, err := goens.NameHash(subscriber.Hex() + utils.Encode(nameHash[:]))
	if err != nil {
		return err
	}
	s.requestMap[utils.Encode(reqHash[:])] = &swarmMail.SwarmMailSubRequest{
		FdpBuyerNameHash: nameHash,
		Buyer:            subscriber,
		SubHash:          subHash,
		RequestHash:      reqHash,
	}
	return nil
}

func (s *SubscriptionManager) GetSubRequests(owner common.Address) ([]swarmMail.SwarmMailSubRequest, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	requests := []swarmMail.SwarmMailSubRequest{}
	for _, r := range s.requestMap {
		sub := s.listMap[utils.Encode(r.SubHash[:])]
		if sub.Seller == owner {
			requests = append(requests, *r)
		}
	}

	return requests, nil
}

func (s *SubscriptionManager) AllowAccess(owner common.Address, si *rpc.ShareInfo, requestHash, secret [32]byte, key *ecdsa.PrivateKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	i, ok := s.requestMap[utils.Encode(requestHash[:])]
	if !ok {
		return fmt.Errorf("request not available")
	}

	item := &swarmMail.SwarmMailSubItem{
		SubHash:           i.SubHash,
		UnlockKeyLocation: [32]byte{},
		ValidTill:         new(big.Int).SetInt64(time.Now().AddDate(0, 1, 0).Unix()),
	}

	s.subscriptionMap[i.Buyer.Hex()+utils.Encode(requestHash[:])] = item

	dt, err := json.Marshal(si)
	if err != nil {
		return err
	}

	encDt, err := utils.EncryptBytes(secret[:], dt)
	if err != nil {
		return err
	}

	s.subscribedMap[utils.Encode(i.SubHash[:])] = encDt

	return nil
}

func (s *SubscriptionManager) GetSubscriptions(subscriber common.Address, _, _ uint64) ([]swarmMail.SwarmMailSubItem, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	subscriberHex := subscriber.Hex()
	pods := []swarmMail.SwarmMailSubItem{}
	for i, v := range s.subscriptionMap {
		if strings.HasPrefix(i, subscriberHex) {
			pods = append(pods, *v)
		}
	}

	return pods, nil
}

func (s *SubscriptionManager) GetSubscription(subscriber common.Address, subHash, secret [32]byte) (*rpc.ShareInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	encDt := s.subscribedMap[utils.Encode(subHash[:])]
	dt, err := utils.DecryptBytes(secret[:], encDt)
	if err != nil {
		return nil, err
	}

	ip := &rpc.ShareInfo{}
	err = json.Unmarshal(dt, ip)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (s *SubscriptionManager) GetAllSubscribablePods() ([]swarmMail.SwarmMailSub, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	pods := []swarmMail.SwarmMailSub{}
	for _, v := range s.listMap {
		if v.Active {
			pods = append(pods, *v)
		}
	}
	return pods, nil
}

func (s *SubscriptionManager) GetOwnSubscribablePods(owner common.Address) ([]swarmMail.SwarmMailSub, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	pods := []swarmMail.SwarmMailSub{}
	for _, v := range s.listMap {
		if v.Seller == owner {
			pods = append(pods, *v)
		}
	}
	return pods, nil
}

func (s *SubscriptionManager) GetSub(subHash [32]byte) (*swarmMail.SwarmMailSub, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[utils.Encode(subHash[:])]
	if ok {
		return i, nil
	}
	return nil, fmt.Errorf("pod not found")
}
