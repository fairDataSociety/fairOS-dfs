package mock

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type PodItem struct {
	Name     string         `json:"name"`
	Price    uint64         `json:"price"`
	Owner    common.Address `json:"owner"`
	IsListed bool           `json:"isListed"`
}

type requestInfo struct {
	Name       string         `json:"name"`
	Subscriber common.Address `json:"owner"`
}

type SubscriptionManager struct {
	lock            sync.Mutex
	listMap         map[string]*PodItem
	subscriptionMap map[string]*PodItem
	requestMap      map[string]requestInfo
}

// NewMockSubscriptionManager returns a new mock subscription manager client
func NewMockSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		listMap:         make(map[string]*PodItem),
		subscriptionMap: make(map[string]*PodItem),
	}
}

func (s *SubscriptionManager) AddPodToMarketplace(owner common.Address, pod string, price uint64) error {
	i := &PodItem{
		Name:     pod,
		Price:    price,
		Owner:    owner,
		IsListed: true,
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	s.listMap[owner.Hex()+pod] = i

	return nil
}

func (s *SubscriptionManager) HidePodFromMarketplace(owner common.Address, pod string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+pod]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	i.IsListed = false
	return nil
}

func (s *SubscriptionManager) RequestAccess(pod string, owner, subscriber common.Address) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+pod]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if !i.IsListed {
		return fmt.Errorf("pod not listed")
	}

	s.requestMap[owner.Hex()+subscriber.Hex()+pod] = requestInfo{
		Name:       pod,
		Subscriber: subscriber,
	}
	return nil
}

func (s *SubscriptionManager) AllowAccess(pod string, owner, subscriber common.Address) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	i, ok := s.listMap[owner.Hex()+pod]
	if !ok {
		return fmt.Errorf("pod not listed")
	}
	if !i.IsListed {
		return fmt.Errorf("pod not listed")
	}

	_, ok = s.requestMap[owner.Hex()+subscriber.Hex()+pod]
	if !ok {
		return fmt.Errorf("request not available")
	}

	s.subscriptionMap[subscriber.Hex()+pod] = i

	return nil
}

func (s *SubscriptionManager) GetSubscriptions(subscriber common.Address) []*PodItem {
	subscriberHex := subscriber.Hex()
	pods := []*PodItem{}
	for i, v := range s.subscriptionMap {
		if strings.HasPrefix(i, subscriberHex) {
			pods = append(pods, v)
		}
	}
	return pods
}

func (s *SubscriptionManager) GetAllSubscribablePods() []*PodItem {
	pods := []*PodItem{}
	for _, v := range s.listMap {
		pods = append(pods, v)
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
