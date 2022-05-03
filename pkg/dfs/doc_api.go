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

// DocCreate is a controller function which does all the checks before creating a documentDB.
func (d *DfsAPI) DocCreate(sessionId, podName, name string, indexes map[string]collection.IndexType, mutable bool) error {
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

	return podInfo.GetDocStore().CreateDocumentDB(name, indexes, mutable)
}

// DocOpen is a controller function which does all the checks before opening a documentDB.
func (d *DfsAPI) DocOpen(sessionId, podName, name string) error {
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

	return podInfo.GetDocStore().OpenDocumentDB(name)
}

// DocDelete is a controller function which does all the checks before deleting a documentDB.
func (d *DfsAPI) DocDelete(sessionId, podName, name string) error {
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

	return podInfo.GetDocStore().DeleteDocumentDB(name)
}

// DocList is a controller function which does all the checks before listing all the
// documentDB available in the pod.
func (d *DfsAPI) DocList(sessionId, podName string) (map[string]collection.DBSchema, error) {
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

	return podInfo.GetDocStore().LoadDocumentDBSchemas()
}

// DocCount is a controller function which does all the checks before counting
// all the documents ina documentDB.
func (d *DfsAPI) DocCount(sessionId, podName, name, expr string) (uint64, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return 0, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return 0, err
	}

	return podInfo.GetDocStore().Count(name, expr)
}

// DocPut is a controller function which does all the checks before inserting
// a document in the documentDB.
func (d *DfsAPI) DocPut(sessionId, podName, name string, value []byte) error {
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

	return podInfo.GetDocStore().Put(name, value)
}

// DocGet is a controller function which does all the checks before retrieving
//// a document in the documentDB.
func (d *DfsAPI) DocGet(sessionId, podName, name, id string) ([]byte, error) {
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

	return podInfo.GetDocStore().Get(name, id)
}

// DocDel is a controller function which does all the checks before deleting
// a documentDB.
func (d *DfsAPI) DocDel(sessionId, podName, name, id string) error {
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

	return podInfo.GetDocStore().Del(name, id)
}

// DocFind is a controller function which does all the checks before finding
// records from a documentDB.
func (d *DfsAPI) DocFind(sessionId, podName, name, expr string, limit int) ([][]byte, error) {
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

	return podInfo.GetDocStore().Find(name, expr, limit)
}

// DocBatch initiates a batch inserting session.
func (d *DfsAPI) DocBatch(sessionId, podName, name string) (*collection.DocBatch, error) {
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

	return podInfo.GetDocStore().CreateDocBatch(name)
}

// DocBatchPut inserts records in to a document batch.
func (d *DfsAPI) DocBatchPut(sessionId, podName string, doc []byte, docBatch *collection.DocBatch) error {
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

	return podInfo.GetDocStore().DocBatchPut(docBatch, doc, 0)
}

// DocBatchWrite commits the batch document insert.
func (d *DfsAPI) DocBatchWrite(sessionId, podName string, docBatch *collection.DocBatch) error {
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

	return podInfo.GetDocStore().DocBatchWrite(docBatch, "")
}

// DocIndexJson indexes a json files in to the document DB.
func (d *DfsAPI) DocIndexJson(sessionId, podName, name, podFileWithPath string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return ErrPodNotOpen
	}

	// check if file present
	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	file := podInfo.GetFile()
	if !file.IsFileAlreadyPresent(podFileWithPath) {
		return ErrFileNotPresent
	}

	return podInfo.GetDocStore().DocFileIndex(name, podFileWithPath)
}
