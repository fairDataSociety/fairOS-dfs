package subscription

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc"
)

type Manager struct {
	pod        *pod.Pod
	addr       common.Address
	ens        ensm.ENSManager
	privateKey *ecdsa.PrivateKey
	sm         rpc.SubscriptionManager
}

type PodItem struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Price   uint64 `json:"price"`
	Owner   string `json:"owner"`
}

func New(pod *pod.Pod, addr common.Address, privateKey *ecdsa.PrivateKey, ensm ensm.ENSManager, sm rpc.SubscriptionManager) *Manager {
	return &Manager{
		pod:        pod,
		addr:       addr,
		ens:        ensm,
		sm:         sm,
		privateKey: privateKey,
	}
}

// ListPod will save the pod info in the subscription smart contract with its owner and price
// we keep the pod info in the smart contract, with a `list` flag
func (m *Manager) ListPod(podname string, podAddress common.Address, price uint64) error {
	if !m.pod.IsOwnPodPresent(podname) {
		return fmt.Errorf("pod not present")
	}
	return m.sm.AddPodToMarketplace(podAddress, m.addr, podname, price)
}

// DelistPod will make the `list` flag false for the pod so that it's not listed in the pod marketplace
func (m *Manager) DelistPod(podAddress common.Address) error {
	return m.sm.HidePodFromMarketplace(podAddress, m.addr)
}

// ApproveSubscription will send a subscription request to the owner of the pod
func (m *Manager) ApproveSubscription(podName string, podAddress, subscriber common.Address, subscriberPublicKey *ecdsa.PublicKey) error {
	a, _ := subscriberPublicKey.Curve.ScalarMult(subscriberPublicKey.X, subscriberPublicKey.Y, m.privateKey.D.Bytes())
	secret := sha256.Sum256(a.Bytes())

	ref, err := m.pod.PodShare(podName, "")
	if err != nil {
		return err
	}
	encRef, err := utils.EncryptBytes(secret[:], []byte(ref))
	if err != nil {
		return err
	}
	return m.sm.AllowAccess(podAddress, m.addr, subscriber, string(encRef))
}

// RequestSubscription will send a subscription request to the owner of the pod
// will create an escrow account and deposit the `price`
func (m *Manager) RequestSubscription(podAddress, owner common.Address) error {
	return m.sm.RequestAccess(podAddress, owner, m.addr)
}

// GetSubscriptions will query the smart contract and list my subscriptions
func (m *Manager) GetSubscriptions() ([]*mock.SubbedItem, error) {
	return m.sm.GetSubscriptions(m.addr), nil
}

// GetMarketplace will query the smart contract make the `list` all the pod from the marketplace
func (m *Manager) GetMarketplace() ([]*mock.PodItem, error) {
	return m.sm.GetAllSubscribablePods(), nil
}

// OpenSubscribedPod will open a subscribed pod
func (m *Manager) OpenSubscribedPod(podAddress common.Address, ownerPublicKey *ecdsa.PublicKey) (*pod.Info, error) {
	a, _ := ownerPublicKey.Curve.ScalarMult(ownerPublicKey.X, ownerPublicKey.Y, m.privateKey.D.Bytes())
	secret := sha256.Sum256(a.Bytes())
	item := m.sm.GetSubscription(podAddress, m.addr)
	refBytes, err := utils.DecryptBytes(secret[:], []byte(item.Secret))
	if err != nil {
		return nil, err
	}
	reference, err := utils.ParseHexReference(string(refBytes))
	if err != nil {
		return nil, err
	}

	return m.pod.OpenFromReference(reference)
}
