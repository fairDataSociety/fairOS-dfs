package subscriptionManager

import (
	"crypto/ecdsa"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"

	swarmMail "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/smail"

	"github.com/ethereum/go-ethereum/common"
)

type SubscriptionManager interface {
	AddPodToMarketplace(podAddress, owner common.Address, pod, title, desc, thumbnail string, price uint64, daysValid uint, category, nameHash [32]byte, key *ecdsa.PrivateKey) error
	HidePodFromMarketplace(owner common.Address, subHash [32]byte, hide bool, key *ecdsa.PrivateKey) error
	RequestAccess(subscriber common.Address, subHash, nameHash [32]byte, key *ecdsa.PrivateKey) error
	AllowAccess(owner common.Address, si *rpc.ShareInfo, requestHash, secret [32]byte, key *ecdsa.PrivateKey) error
	GetSubscription(subscriber common.Address, subHash, secret [32]byte) (*rpc.ShareInfo, error)
	GetSubscriptions(subscriber common.Address) ([]swarmMail.SwarmMailSubItem, error)
	GetAllSubscribablePods() ([]swarmMail.SwarmMailSub, error)
	GetOwnSubscribablePods(owner common.Address) ([]swarmMail.SwarmMailSub, error)
	GetSubscribablePodInfo(subHash [32]byte) (*rpc.SubscriptionItemInfo, error)
	GetSubRequests(owner common.Address) ([]swarmMail.SwarmMailSubRequest, error)
	GetSub(subHash [32]byte) (*swarmMail.SwarmMailSub, error)
}
