package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"
)

type SubscriptionManager interface {
	AddPodToMarketplace(podAddress, owner common.Address, pod string, price uint64) error
	HidePodFromMarketplace(podAddress, owner common.Address) error
	RequestAccess(podAddress, owner, subscriber common.Address) error
	AllowAccess(podAddress, owner, subscriber common.Address, secret string) error
	GetSubscription(podAddress, subscriber common.Address) *mock.SubbedItem
	GetSubscriptions(subscriber common.Address) []*mock.SubbedItem
	GetAllSubscribablePods() []*mock.PodItem
	GetOwnSubscribablePods(owner common.Address) []*mock.PodItem
}
