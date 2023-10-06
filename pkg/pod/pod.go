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

package pod

import (
	"fmt"
	"sync"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"
)

const (
	maxPodId = 65535

	// PasswordLength defines the pod password length
	PasswordLength = 32
)

// Pod is the main struct which acts on pods
type Pod struct {
	fd            *feed.API
	acc           *account.Account
	client        blockstore.Client
	podMap        map[string]*Info //  podName -> dir
	podMu         *sync.RWMutex
	logger        logging.Logger
	tm            taskmanager.TaskManagerGO
	sm            subscriptionManager.SubscriptionManager
	feedCacheSize int
	feedCacheTTL  time.Duration
}

// ListItem defines the structure for pod item
type ListItem struct {
	Name     string `json:"name"`
	Index    int    `json:"index"`
	Password string `json:"password"`
}

// SharedListItem defines the structure for shared pod item
type SharedListItem struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Password string `json:"password"`
}

// List lists all the pods
type List struct {
	Pods       []ListItem       `json:"pods"`
	SharedPods []SharedListItem `json:"sharedPods"`
}

// NewPod creates the main pod object which has all the methods related to the pods.
func NewPod(client blockstore.Client, feed *feed.API, account *account.Account, m taskmanager.TaskManagerGO, sm subscriptionManager.SubscriptionManager, feedCacheSize int, feedCacheTTL time.Duration, logger logging.Logger) *Pod {
	return &Pod{
		fd:            feed,
		acc:           account,
		client:        client,
		podMap:        make(map[string]*Info),
		podMu:         &sync.RWMutex{},
		logger:        logger,
		tm:            m,
		sm:            sm,
		feedCacheSize: feedCacheSize,
		feedCacheTTL:  feedCacheTTL,
	}
}

func (p *Pod) addPodToPodMap(podName string, podInfo *Info) {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	p.podMap[podName] = podInfo
}

func (p *Pod) removePodFromPodMap(podName string) {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	delete(p.podMap, podName)
}

// GetPodInfoFromPodMap returns the pod info for the given pod name.
func (p *Pod) GetPodInfoFromPodMap(podName string) (*Info, string, error) {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	if podInfo, ok := p.podMap[podName]; ok {
		return podInfo, podInfo.podPassword, nil
	}
	return nil, "", fmt.Errorf("could not find pod: %s", podName)
}

// GetFeed returns the feed object
func (p *Pod) GetFeed() *feed.API {
	return p.fd
}

// GetAccount returns the account object
func (p *Pod) GetAccount() *account.Account {
	return p.acc
}

// GetPodInfo returns the pod info for the given pod name.
func (p *Pod) GetPodInfo(podName string) (*Info, string, error) {
	pi, password, _ := p.GetPodInfoFromPodMap(podName)
	if pi != nil {
		return pi, password, nil
	}
	pi, err := p.OpenPod(podName)
	if err != nil {
		return nil, "", err
	}
	password = pi.GetPodPassword()
	return pi, password, nil
}
