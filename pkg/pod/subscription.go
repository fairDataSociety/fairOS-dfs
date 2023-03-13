package pod

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	swarmMail "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/smail"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// ListPodInMarketplace will save the pod info in the subscriptionManager smart contract with its owner and price
// we keep the pod info in the smart contract, with a `list` flag
func (p *Pod) ListPodInMarketplace(podName, title, desc, thumbnail string, price uint64, daysValid uint, category, nameHash [32]byte) error {
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return err
	}
	if !p.checkIfPodPresent(podList, podName) {
		return ErrInvalidPodName
	}

	strAddr, _ := p.getAddressPassword(podList, podName)
	if strAddr == "" { // skipcq: TCV-001
		return fmt.Errorf("pod does not exist")
	}

	podAddress := common.HexToAddress(strAddr)

	return p.sm.AddPodToMarketplace(podAddress, common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), podName, title, desc, thumbnail, price, daysValid, category, nameHash, p.acc.GetUserAccountInfo().GetPrivateKey())
}

// PodStatusInMarketplace will change the `list` flag for the pod so that it's not listed or gets re listed in the pod marketplace
func (p *Pod) PodStatusInMarketplace(subHash [32]byte, show bool) error {
	hide := !show
	return p.sm.HidePodFromMarketplace(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), subHash, hide, p.acc.GetUserAccountInfo().GetPrivateKey())
}

// ApproveSubscription will send a subscriptionManager request to the owner of the pod
func (p *Pod) ApproveSubscription(podName string, requestHash [32]byte, subscriberPublicKey *ecdsa.PublicKey) error {
	a, _ := subscriberPublicKey.Curve.ScalarMult(subscriberPublicKey.X, subscriberPublicKey.Y, p.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
	secret := sha256.Sum256(a.Bytes())

	shareInfo, err := p.GetPodSharingInfo(podName)
	if err != nil {
		return err
	}

	info := &rpc.ShareInfo{
		PodName:     shareInfo.PodName,
		Address:     shareInfo.Address,
		Password:    shareInfo.Password,
		UserAddress: shareInfo.UserAddress,
	}

	return p.sm.AllowAccess(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), info, requestHash, secret, p.acc.GetUserAccountInfo().GetPrivateKey())
}

// EncryptUploadSubscriptionInfo will upload sub pod info into swarm
func (p *Pod) EncryptUploadSubscriptionInfo(podName string, subscriberPublicKey *ecdsa.PublicKey) (string, error) {
	a, _ := subscriberPublicKey.Curve.ScalarMult(subscriberPublicKey.X, subscriberPublicKey.Y, p.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
	secret := sha256.Sum256(a.Bytes())

	shareInfo, err := p.GetPodSharingInfo(podName)
	if err != nil {
		return "", err
	}

	info := &rpc.ShareInfo{
		PodName:     shareInfo.PodName,
		Address:     shareInfo.Address,
		Password:    shareInfo.Password,
		UserAddress: shareInfo.UserAddress,
	}

	data, err := json.Marshal(info)
	if err != nil { // skipcq: TCV-001
		return "", err
	}
	encData, err := utils.EncryptBytes(secret[:], data)
	if err != nil {
		return "", err
	}

	ref, err := p.client.UploadBlob(encData, 0, false)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ref), nil
}

// RequestSubscription will send a subscriptionManager request to the owner of the pod
// will create an escrow account and deposit the `price`
func (p *Pod) RequestSubscription(subHash, nameHash [32]byte) error {
	return p.sm.RequestAccess(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), subHash, nameHash, p.acc.GetUserAccountInfo().GetPrivateKey())
}

// GetSubscriptions will query the smart contract and list my subscriptions
func (p *Pod) GetSubscriptions(start, limit uint64) ([]swarmMail.SwarmMailSubItem, error) {
	return p.sm.GetSubscriptions(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), start, limit)
}

// GetMarketplace will query the smart contract make the `list` all the pod from the marketplace
func (p *Pod) GetMarketplace() ([]swarmMail.SwarmMailSub, error) {
	return p.sm.GetAllSubscribablePods()
}

// GetSubscribablePodInfo will query the smart contract and get info by subHash
func (p *Pod) GetSubscribablePodInfo(subHash [32]byte) (*rpc.SubscriptionItemInfo, error) {
	return p.sm.GetSubscribablePodInfo(subHash)
}

// OpenSubscribedPod will open a subscribed pod
func (p *Pod) OpenSubscribedPod(subHash [32]byte, ownerPublicKey *ecdsa.PublicKey) (*Info, error) {
	a, _ := ownerPublicKey.Curve.ScalarMult(ownerPublicKey.X, ownerPublicKey.Y, p.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
	secret := sha256.Sum256(a.Bytes())
	info, err := p.sm.GetSubscription(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()), subHash, secret)
	if err != nil {
		return nil, err
	}

	shareInfo := &ShareInfo{
		PodName:     info.PodName,
		Address:     info.Address,
		Password:    info.Password,
		UserAddress: info.UserAddress,
	}
	return p.OpenFromShareInfo(shareInfo)
}

// OpenSubscribedPodFromReference will open a subscribed pod
func (p *Pod) OpenSubscribedPodFromReference(reference string, ownerPublicKey *ecdsa.PublicKey) (*Info, error) {
	a, _ := ownerPublicKey.Curve.ScalarMult(ownerPublicKey.X, ownerPublicKey.Y, p.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
	secret := sha256.Sum256(a.Bytes())

	ref, err := hex.DecodeString(reference)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	encData, resp, err := p.client.DownloadBlob(ref)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if resp != http.StatusOK { // skipcq: TCV-001
		return nil, fmt.Errorf("OpenSubscribedPodFromReference: could not get subscription info")
	}

	data, err := utils.DecryptBytes(secret[:], encData)
	if err != nil {
		return nil, err
	}
	var info *rpc.ShareInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}

	shareInfo := &ShareInfo{
		PodName:     info.PodName,
		Address:     info.Address,
		Password:    info.Password,
		UserAddress: info.UserAddress,
	}
	return p.OpenFromShareInfo(shareInfo)
}

// GetSubRequests will get all owners sub requests
func (p *Pod) GetSubRequests() ([]swarmMail.SwarmMailSubRequest, error) {
	return p.sm.GetSubRequests(common.HexToAddress(p.acc.GetUserAccountInfo().GetAddress().Hex()))
}
