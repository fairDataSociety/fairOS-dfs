/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dfs

import (
	"context"
	"encoding/hex"

	SwarmMail "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/smail"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// CreatePod
func (a *API) CreatePod(podName, sessionId string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// open the pod
	pi, err := a.prepareOwnPod(ui, podName)
	if err != nil {
		return nil, err
	}

	// Add podName in the login user session
	ui.AddPodName(podName, pi)
	return pi, nil
}

// DeletePod deletes a pod
func (a *API) DeletePod(podName, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// delete all the directory, files, and database tables under this pod from
	// the Swarm network.
	podInfo, _, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	directory := podInfo.GetDirectory()

	// check if this is a shared pod
	if podInfo.GetFeed().IsReadOnlyFeed() {
		// delete the pod and close if it is opened
		err = ui.GetPod().DeleteSharedPod(podName)
		if err != nil {
			return err
		}

		// close the pod if it is open
		if ui.IsPodOpen(podName) {
			// remove from the login session
			ui.RemovePodName(podName)
		}
		return nil
	}

	err = directory.RmRootDir(podInfo.GetPodPassword())
	if err != nil {
		return err
	}

	// delete the pod and close if it is opened
	err = ui.GetPod().DeleteOwnPod(podName)
	if err != nil {
		return err
	}

	// close the pod if it is open
	if ui.IsPodOpen(podName) {
		// remove from the login session
		ui.RemovePodName(podName)
	}

	return nil
}

// OpenPod
func (a *API) OpenPod(podName, sessionId string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	// return if pod already open
	if ui.IsPodOpen(podName) {
		podInfo, _, err := ui.GetPod().GetPodInfoFromPodMap(podName)
		if err != nil {
			return nil, err
		}
		return podInfo, nil
	}
	// open the pod
	pi, err := ui.GetPod().OpenPod(podName)
	if err != nil {
		return nil, err
	}
	err = pi.GetDirectory().AddRootDir(pi.GetPodName(), pi.GetPodPassword(), pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(podName, pi)
	return pi, nil
}

// OpenPodAsync
func (a *API) OpenPodAsync(ctx context.Context, podName, sessionId string) (*pod.Info, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	// return if pod already open
	if ui.IsPodOpen(podName) {
		podInfo, _, err := ui.GetPod().GetPodInfoFromPodMap(podName)
		if err != nil {
			return nil, err
		}
		return podInfo, nil
	}
	// open the pod
	pi, err := ui.GetPod().OpenPodAsync(ctx, podName)
	if err != nil {
		return nil, err
	}
	err = pi.GetDirectory().AddRootDir(pi.GetPodName(), pi.GetPodPassword(), pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(podName, pi)
	return pi, nil
}

// ClosePod
func (a *API) ClosePod(podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// close the pod
	err := ui.GetPod().ClosePod(podName)
	if err != nil {
		return err
	}

	// delete podName in the login user session
	ui.RemovePodName(podName)
	return nil
}

// PodStat
func (a *API) PodStat(podName, sessionId string) (*pod.Stat, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// get the pod stat
	podStat, err := ui.GetPod().PodStat(podName)
	if err != nil {
		return nil, err
	}
	return podStat, nil
}

// SyncPod
func (a *API) SyncPod(podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// sync the pod
	err := ui.GetPod().SyncPod(podName)
	if err != nil {
		return err
	}
	return nil
}

// SyncPodAsync
func (a *API) SyncPodAsync(ctx context.Context, podName, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// sync the pod
	err := ui.GetPod().SyncPodAsync(ctx, podName)
	if err != nil {
		return err
	}
	return nil
}

// ListPods
func (a *API) ListPods(sessionId string) ([]string, []string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// list pods of a user
	pods, sharedPods, err := ui.GetPod().ListPods()
	if err != nil {
		return nil, nil, err
	}
	return pods, sharedPods, nil
}

// PodList lists all available pods in json format
func (a *API) PodList(sessionId string) (*pod.List, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// list pods of a user
	return ui.GetPod().PodList()
}

// PodShare
func (a *API) PodShare(podName, sharedPodName, sessionId string) (string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", ErrUserNotLoggedIn
	}

	// get the pod stat
	address, err := ui.GetPod().PodShare(podName, sharedPodName)
	if err != nil {
		return "", err
	}
	return address, nil
}

// PodReceiveInfo
func (a *API) PodReceiveInfo(sessionId string, ref utils.Reference) (*pod.ShareInfo, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePodInfo(ref)
}

// PodReceive
func (a *API) PodReceive(sessionId, sharedPodName string, ref utils.Reference) (*pod.Info, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().ReceivePod(sharedPodName, ref)
}

// IsPodExist
func (a *API) IsPodExist(podName, sessionId string) bool {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return false
	}
	return ui.GetPod().IsPodPresent(podName)
}

// ForkPod
func (a *API) ForkPod(podName, forkName, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	if forkName == "" {
		return pod.ErrBlankPodName
	}

	if ui.GetPod().IsPodPresent(forkName) {
		return pod.ErrForkAlreadyExists
	}

	_, err := a.prepareOwnPod(ui, forkName)
	if err != nil {
		return err
	}

	return ui.GetPod().PodFork(podName, forkName)
}

// ForkPodFromRef
func (a *API) ForkPodFromRef(forkName, refString, sessionId string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if refString == "" {
		return pod.ErrBlankPodSharingReference
	}
	if forkName == "" {
		return pod.ErrBlankPodName
	}

	if ui.GetPod().IsPodPresent(forkName) {
		return pod.ErrForkAlreadyExists
	}

	_, err := a.prepareOwnPod(ui, forkName)
	if err != nil {
		return err
	}

	return ui.GetPod().PodForkFromRef(forkName, refString)
}

func (a *API) prepareOwnPod(ui *user.Info, podName string) (*pod.Info, error) {
	podPasswordBytes, _ := utils.GetRandBytes(pod.PasswordLength)
	podPassword := hex.EncodeToString(podPasswordBytes)

	// create the pod
	_, err := ui.GetPod().CreatePod(podName, "", podPassword)
	if err != nil {
		return nil, err
	}

	// open the pod
	pi, err := ui.GetPod().OpenPod(podName)
	if err != nil {
		return nil, err
	}

	// create the root directory
	err = pi.GetDirectory().MkRootDir(pi.GetPodName(), podPassword, pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// ListPodInMarketplace
func (a *API) ListPodInMarketplace(podName, title, desc, thumbnail, sessionId string, price uint64, category [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	nameHash, err := a.users.GetNameHash(ui.GetUserName())
	if err != nil {
		return err
	}

	return ui.GetPod().ListPodInMarketplace(podName, title, desc, thumbnail, price, category, nameHash)
}

// ChangePodListStatusInMarketplace
func (a *API) ChangePodListStatusInMarketplace(sessionId string, subHash [32]byte, show bool) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	return ui.GetPod().PodStatusInMarketplace(subHash, show)
}

// RequestSubscription
func (a *API) RequestSubscription(sessionId string, subHash [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	nameHash, err := a.users.GetNameHash(ui.GetUserName())
	if err != nil {
		return err
	}
	return ui.GetPod().RequestSubscription(subHash, nameHash)
}

// ApproveSubscription
func (a *API) ApproveSubscription(sessionId, podName string, reqHash, nameHash [32]byte) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return errNilSubManager
	}

	_, subscriberPublicKey, err := a.users.GetUserInfoFromENS(nameHash)
	if err != nil {
		return err
	}

	return ui.GetPod().ApproveSubscription(podName, reqHash, subscriberPublicKey)
}

type SubscriptionInfo struct {
	SubHash      [32]byte
	PodName      string
	PodAddress   string
	InfoLocation []byte
	ValidTill    int64
}

// GetSubscriptions
func (a *API) GetSubscriptions(sessionId string, start, limit uint64) ([]*SubscriptionInfo, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	if a.sm == nil {
		return nil, errNilSubManager
	}

	subscriptions, err := ui.GetPod().GetSubscriptions(start, limit)
	if err != nil {
		return nil, err
	}

	subs := []*SubscriptionInfo{}
	for _, item := range subscriptions {
		info, err := ui.GetPod().GetSubscribablePodInfo(item.SubHash)
		if err != nil {
			return subs, err
		}
		sub := &SubscriptionInfo{
			SubHash:      item.SubHash,
			PodName:      info.PodName,
			PodAddress:   info.PodAddress,
			InfoLocation: item.UnlockKeyLocation[:],
			ValidTill:    item.ValidTill.Int64(),
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

// OpenSubscribedPod
func (a *API) OpenSubscribedPod(sessionId string, subHash [32]byte) (*pod.Info, error) {

	sub, err := a.sm.GetSub(subHash)
	if err != nil {
		return nil, err
	}

	_, ownerPublicKey, err := a.users.GetUserInfoFromENS(sub.FdpSellerNameHash)
	if err != nil {
		return nil, err
	}

	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// open the pod
	pi, err := ui.GetPod().OpenSubscribedPod(subHash, ownerPublicKey)
	if err != nil {
		return nil, err
	}
	err = pi.GetDirectory().AddRootDir(pi.GetPodName(), pi.GetPodPassword(), pi.GetPodAddress(), pi.GetFeed())
	if err != nil {
		return nil, err
	}
	// Add podName in the login user session
	ui.AddPodName(pi.GetPodName(), pi)
	return pi, nil
}

// GetSubscribablePods
func (a *API) GetSubscribablePods(sessionId string) ([]SwarmMail.SwarmMailSub, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return ui.GetPod().GetMarketplace()
}
