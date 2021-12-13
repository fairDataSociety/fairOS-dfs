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
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
)

// KVCreate does validation checks and calls the create KVtable function.
func (d *DfsAPI) KVCreate(sessionId, podName, name string, indexType collection.IndexType) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().CreateKVTable(name, indexType)
}

// KVDelete does validation checks and calls the delete KVtable function.
func (d *DfsAPI) KVDelete(sessionId, podName, name string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().DeleteKVTable(name)
}

// KVOpen does validation checks and calls the open KVtable function.
func (d *DfsAPI) KVOpen(sessionId, podName, name string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().OpenKVTable(name)
}

// KVList does validation checks and calls the list KVtable function.
func (d *DfsAPI) KVList(sessionId, podName string) (map[string][]string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().LoadKVTables()
}

// KVCount does validation checks and calls the count KVtable function.
func (d *DfsAPI) KVCount(sessionId, podName, name string) (*collection.KVCount, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVCount(name)
}

// KVPut does validation checks and calls the put KVtable function.
func (d *DfsAPI) KVPut(sessionId, podName, name, key string, value []byte) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().KVPut(name, key, value)
}

// KVGet does validation checks and calls the get KVtable function.
func (d *DfsAPI) KVGet(sessionId, podName, name, key string) ([]string, []byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, nil, err
	}

	return podInfo.GetKVStore().KVGet(name, key)
}

// KVDel does validation checks and calls the delete KVtable function.
func (d *DfsAPI) KVDel(sessionId, podName, name, key string) ([]byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVDelete(name, key)
}

// KVBatch does validation checks and calls the batch KVtable function.
func (d *DfsAPI) KVBatch(sessionId, podName, name string, columns []string) (*collection.Batch, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVBatch(name, columns)
}

// KVBatchPut does validation checks and calls the batch put KVtable function.
func (d *DfsAPI) KVBatchPut(sessionId, podName, key string, value []byte, batch *collection.Batch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	return batch.Put(key, value, false, false)
}

// KVBatchWrite does validation checks and calls the batch write KVtable function.
func (d *DfsAPI) KVBatchWrite(sessionId, podName string, batch *collection.Batch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	_, err := batch.Write("")
	return err
}

// KVSeek does validation checks and calls the seek KVtable function.
func (d *DfsAPI) KVSeek(sessionId, podName, name, start, end string, limit int64) (*collection.Iterator, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVSeek(name, start, end, limit)
}

// KVGetNext does validation checks and calls the get next KVtable function.
func (d *DfsAPI) KVGetNext(sessionId, podName, name string) ([]string, string, []byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, "", nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return nil, "", nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, "", nil, err
	}

	return podInfo.GetKVStore().KVGetNext(name)
}
