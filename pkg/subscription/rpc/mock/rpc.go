package mock

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type PodItem struct {
	Name     string         `json:"name"`
	Price    uint64         `json:"price"`
	Address  common.Address `json:"address"`
	Owner    common.Address `json:"owner"`
	IsListed bool           `json:"isListed"`
}

type SubbedItem struct {
	Name    string         `json:"name"`
	Address common.Address `json:"address"`
	EndsAt  int64          `json:"ends_at"`
	Secret  string         `json:"secret"`
	Owner   common.Address `json:"owner"`
}

type requestInfo struct {
	Name       string         `json:"name"`
	Address    common.Address `json:"address"`
	Subscriber common.Address `json:"owner"`
}

type SubscriptionManager struct {
	lock            sync.Mutex
	listMap         map[string]*PodItem
	subscriptionMap map[string]*SubbedItem
	requestMap      map[string]requestInfo
}

// NewMockSubscriptionManager returns a new mock subscription manager client
func NewMockSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		listMap:         make(map[string]*PodItem),
		subscriptionMap: make(map[string]*SubbedItem),
		requestMap:      make(map[string]requestInfo),
	}
}

func (s *SubscriptionManager) AddPodToMarketplace(podAddress, owner common.Address, pod string, price uint64) error {
	i := &PodItem{
		Name:     pod,
		Price:    price,
		Address:  podAddress,
		Owner:    owner,
		IsListed: true,
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	s.listMap[owner.Hex()+podAddress.String()] = i
	return nil
}

func (s *SubscriptionManager) HidePodFromMarketplace(podAddress, owner common.Address) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+podAddress.String()]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	i.IsListed = false
	return nil
}

func (s *SubscriptionManager) RequestAccess(podAddress, owner, subscriber common.Address) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+podAddress.String()]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if !i.IsListed {
		return fmt.Errorf("pod not listed")
	}

	s.requestMap[owner.Hex()+subscriber.Hex()+podAddress.String()] = requestInfo{
		Name:       i.Name,
		Address:    podAddress,
		Subscriber: subscriber,
	}
	return nil
}

func (s *SubscriptionManager) AllowAccess(podAddress, owner, subscriber common.Address, secret string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+podAddress.String()]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if !i.IsListed {
		return fmt.Errorf("pod not listed")
	}

	_, ok = s.requestMap[owner.Hex()+subscriber.Hex()+podAddress.String()]
	if !ok {
		return fmt.Errorf("request not available")
	}

	item := &SubbedItem{
		Name:    i.Name,
		Address: i.Address,
		EndsAt:  time.Now().AddDate(0, 1, 0).Unix(),
		Secret:  secret,
		Owner:   owner,
	}
	s.subscriptionMap[subscriber.Hex()+podAddress.String()] = item

	return nil
}

func (s *SubscriptionManager) GetSubscriptions(subscriber common.Address) []*SubbedItem {
	subscriberHex := subscriber.Hex()
	pods := []*SubbedItem{}
	for i, v := range s.subscriptionMap {
		if strings.HasPrefix(i, subscriberHex) {
			pods = append(pods, v)
		}
	}
	return pods
}

func (s *SubscriptionManager) GetSubscription(podAddress, subscriber common.Address) *SubbedItem {
	return s.subscriptionMap[subscriber.Hex()+podAddress.String()]
}

func (s *SubscriptionManager) GetAllSubscribablePods() []*PodItem {
	pods := []*PodItem{}
	for _, v := range s.listMap {
		if v.IsListed {
			pods = append(pods, v)
		}
	}
	return pods
}

func (s *SubscriptionManager) GetOwnSubscribablePods(owner common.Address) []*PodItem {
	ownerHex := owner.Hex()
	pods := []*PodItem{}
	for i, v := range s.listMap {
		if strings.HasPrefix(i, ownerHex) {
			pods = append(pods, v)
		}
	}
	return pods
}
