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

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
)

const (
	maxPodId = 65535
)

type Pod struct {
	fd     *feed.API
	acc    *account.Account
	client blockstore.Client
	podMap map[string]*Info //  podName -> dir
	podMu  *sync.RWMutex
	logger logging.Logger
}

func NewPod(client blockstore.Client, feed *feed.API, account *account.Account, logger logging.Logger) *Pod {
	return &Pod{
		fd:     feed,
		acc:    account,
		client: client,
		podMap: make(map[string]*Info),
		podMu:  &sync.RWMutex{},
		logger: logger,
	}
}

func (p *Pod) GetClient() blockstore.Client {
	return p.client
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

func (p *Pod) GetPodInfoFromPodMap(podName string) (*Info, error) {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	if podInfo, ok := p.podMap[podName]; ok {
		return podInfo, nil
	}
	return nil, fmt.Errorf("could not find pod: %s", podName)
}

func (p *Pod) GetFeed() *feed.API{
	return p.fd
}
func (p *Pod) GetAccount() *account.Account {
	return p.acc
}
