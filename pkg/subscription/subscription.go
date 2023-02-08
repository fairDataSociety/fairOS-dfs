package subscription

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc"
)

type Manager struct {
	pod  *pod.Pod
	addr common.Address
	ens  ensm.ENSManager

	sm rpc.SubscriptionManager
}

type PodItem struct {
	Name  string `json:"name"`
	Price uint64 `json:"price"`
	Owner string `json:"owner"`
}

func New(pod *pod.Pod, addr common.Address, ensm ensm.ENSManager, sm rpc.SubscriptionManager) *Manager {
	return &Manager{
		pod:  pod,
		addr: addr,
		ens:  ensm,
		sm:   sm,
	}
}

// ListPod will save the pod info in the subscription smart contract with its owner and price
// we keep the pod info in the smart contract, with a `list` flag
func (m *Manager) ListPod(podname string, price uint64) error {
	if !m.pod.IsOwnPodPresent(podname) {
		return fmt.Errorf("pod not present")
	}
	return m.sm.AddPodToMarketplace(m.addr, podname, price)
}

// DelistPod will make the `list` flag false for the pod so that it's not listed in the pod marketplace
func (m *Manager) DelistPod(podname string) error {
	return m.sm.HidePodFromMarketplace(m.addr, podname)
}

// ApproveSubscription will send a subscription request to the owner of the pod
func (m *Manager) ApproveSubscription(podname string, subscriber string) error {
	subscriberAddr, err := m.ens.GetOwner(subscriber)
	if err != nil {
		return err
	}
	return m.sm.AllowAccess(podname, m.addr, subscriberAddr)
}

// RequestSubscription will send a subscription request to the owner of the pod
// will create an escrow account and deposit the `price`
func (m *Manager) RequestSubscription(pod string, owner string) error {
	ownerAddr, err := m.ens.GetOwner(owner)
	if err != nil {
		return err
	}

	return m.sm.RequestAccess(pod, ownerAddr, m.addr)
}

func (*Manager) GetSubscriptions() ([]PodItem, error) {
	// This will query the smart contract and list my subscriptions
	return nil, nil
}

func (*Manager) GetMarketplace() ([]PodItem, error) {
	// This will query the smart contract make the `list` all the pod from the marketplace
	return nil, nil
}
