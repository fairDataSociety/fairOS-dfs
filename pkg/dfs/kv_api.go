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

type KVGetter interface {
	KVGet(name, key string) ([]string, []byte, error)
	OpenKVTable(name, encryptionPassword string) error
}

// KVCreate does validation checks and calls the create KVtable function.
func (a *API) KVCreate(sessionId, podName, name string, indexType collection.IndexType) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().CreateKVTable(name, podInfo.GetPodPassword(), indexType)
}

// KVDelete does validation checks and calls the delete KVtable function.
func (a *API) KVDelete(sessionId, podName, name string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().DeleteKVTable(name, podInfo.GetPodPassword())
}

// KVOpen does validation checks and calls the open KVtable function.
func (a *API) KVOpen(sessionId, podName, name string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().OpenKVTable(name, podInfo.GetPodPassword())
}

// KVList does validation checks and calls the list KVtable function.
func (a *API) KVList(sessionId, podName string) (map[string][]string, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().LoadKVTables(podInfo.GetPodPassword())
}

// KVCount does validation checks and calls the count KVtable function.
func (a *API) KVCount(sessionId, podName, name string) (*collection.TableKeyCount, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVCount(name)
}

// KVPut does validation checks and calls the put KVtable function.
func (a *API) KVPut(sessionId, podName, name, key string, value []byte) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return err
	}

	return podInfo.GetKVStore().KVPut(name, key, value)
}

// KVGet does validation checks and calls the get KVtable function.
func (a *API) KVGet(sessionId, podName, name, key string) ([]string, []byte, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, nil, err
	}

	return podInfo.GetKVStore().KVGet(name, key)
}

// KVDel does validation checks and calls the delete KVtable function.
func (a *API) KVDel(sessionId, podName, name, key string) ([]byte, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVDelete(name, key)
}

// KVBatch does validation checks and calls the batch KVtable function.
func (a *API) KVBatch(sessionId, podName, name string, columns []string) (*collection.Batch, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVBatch(name, columns)
}

// KVBatchPut does validation checks and calls the batch put KVtable function.
func (a *API) KVBatchPut(sessionId, podName, key string, value []byte, batch *collection.Batch) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return batch.Put(key, value, false, false)
}

// KVBatchWrite does validation checks and calls the batch write KVtable function.
func (a *API) KVBatchWrite(sessionId, podName string, batch *collection.Batch) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	_, err := batch.Write("")
	return err
}

// KVSeek does validation checks and calls the seek KVtable function.
func (a *API) KVSeek(sessionId, podName, name, start, end string, limit int64) (*collection.Iterator, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, err
	}

	return podInfo.GetKVStore().KVSeek(name, start, end, limit)
}

// KVGetNext does validation checks and calls the get next KVtable function.
func (a *API) KVGetNext(sessionId, podName, name string) ([]string, string, []byte, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, "", nil, ErrUserNotLoggedIn
	}

	podInfo, _, err := ui.GetPod().GetPodInfo(podName)
	if err != nil {
		return nil, "", nil, err
	}

	return podInfo.GetKVStore().KVGetNext(name)
}
