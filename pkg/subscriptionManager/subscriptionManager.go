package subscriptionManager

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/datahub"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"
)

type SubscriptionManager interface {
	AddPodToMarketplace(podAddress, owner common.Address, pod, title, desc, thumbnail string, price uint64, daysValid uint16, category, nameHash [32]byte, key *ecdsa.PrivateKey) error
	HidePodFromMarketplace(owner common.Address, subHash [32]byte, hide bool, key *ecdsa.PrivateKey) error
	RequestAccess(subscriber common.Address, subHash, nameHash [32]byte, key *ecdsa.PrivateKey) error
	AllowAccess(owner common.Address, si *rpc.ShareInfo, requestHash, secret [32]byte, key *ecdsa.PrivateKey) error
	GetSubscription(subscriber common.Address, subHash, secret [32]byte) (*rpc.ShareInfo, error)
	GetSubscriptions(subscriber common.Address) ([]datahub.DataHubSubItem, error)
	GetAllSubscribablePods() ([]datahub.DataHubSub, error)
	GetOwnSubscribablePods(owner common.Address) ([]datahub.DataHubSub, error)
	GetSubscribablePodInfo(subHash [32]byte) (*rpc.SubscriptionItemInfo, error)
	GetSubRequests(owner common.Address) ([]datahub.DataHubSubRequest, error)
	GetSub(subHash [32]byte) (*datahub.DataHubSub, error)
}
