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

import "github.com/fairdatasociety/fairOS-dfs/pkg/collection"

func (d *DfsAPI) DocCreate(sessionId, name string, si map[string]collection.IndexType) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().CreateDocumentDB(name, si)
}

func (d *DfsAPI) DocOpen(sessionId, name string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().OpenDocumentDB(name)
}

func (d *DfsAPI) DocDelete(sessionId, name string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().DeleteDocumentDB(name)
}

func (d *DfsAPI) DocList(sessionId string) (map[string]collection.DBSchema, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}

	return podInfo.GetDocStore().LoadDocumentDBSchemas()
}

func (d *DfsAPI) DocCount(sessionId, name, expr string) (uint64, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return 0, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return 0, err
	}

	return podInfo.GetDocStore().Count(name, expr)
}

func (d *DfsAPI) DocPut(sessionId, name string, value []byte) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().Put(name, value)
}

func (d *DfsAPI) DocGet(sessionId, name, id string) ([]byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}

	return podInfo.GetDocStore().Get(name, id)
}

func (d *DfsAPI) DocDel(sessionId, name, id string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().Del(name, id)
}

func (d *DfsAPI) DocFind(sessionId, name, expr string, limit int) ([][]byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}

	return podInfo.GetDocStore().Find(name, expr, limit)
}

func (d *DfsAPI) DocBatch(sessionId, name string) (*collection.DocBatch, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return nil, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return nil, err
	}

	return podInfo.GetDocStore().CreateDocBatch(name)
}

func (d *DfsAPI) DocBatchPut(sessionId string, doc []byte, docBatch *collection.DocBatch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().DocBatchPut(docBatch, doc)
}

func (d *DfsAPI) DocBatchWrite(sessionId string, docBatch *collection.DocBatch) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if ui.GetPodName() == "" {
		return ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(ui.GetPodName())
	if err != nil {
		return err
	}

	return podInfo.GetDocStore().DocBatchWrite(docBatch)
}
