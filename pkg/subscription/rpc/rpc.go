package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"
)

type SubscriptionManager interface {
	AddPodToMarketplace(owner common.Address, pod string, price uint64) error
	HidePodFromMarketplace(owner common.Address, pod string) error
	RequestAccess(pod string, owner, subscriber common.Address) error
	AllowAccess(pod string, owner, subscriber common.Address) error
	GetSubscriptions(subscriber common.Address) []*mock.PodItem
	GetAllSubscribablePods() []*mock.PodItem
	GetOwnSubscribablePods(owner common.Address) []*mock.PodItem
}
