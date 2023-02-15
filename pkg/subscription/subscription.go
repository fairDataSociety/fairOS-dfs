package subscription

import (
	"fmt"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"

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
func (m *Manager) ApproveSubscription(podname string, subscriber common.Address) error {
	return m.sm.AllowAccess(podname, m.addr, subscriber)
}

// RequestSubscription will send a subscription request to the owner of the pod
// will create an escrow account and deposit the `price`
func (m *Manager) RequestSubscription(pod string, owner common.Address) error {
	return m.sm.RequestAccess(pod, owner, m.addr)
}

// GetSubscriptions will query the smart contract and list my subscriptions
func (m *Manager) GetSubscriptions() ([]*mock.PodItem, error) {
	return m.sm.GetSubscriptions(m.addr), nil
}

// GetMarketplace will query the smart contract make the `list` all the pod from the marketplace
func (m *Manager) GetMarketplace() ([]*mock.PodItem, error) {
	return m.sm.GetAllSubscribablePods(), nil
}
