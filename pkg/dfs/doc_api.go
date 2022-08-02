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
func (d *API) DocCreate(sessionId, podName, name string, indexes map[string]collection.IndexType, mutable bool) error {
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
func (d *API) DocOpen(sessionId, podName, name string) error {
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
func (d *API) DocDelete(sessionId, podName, name string) error {
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
func (d *API) DocList(sessionId, podName string) (map[string]collection.DBSchema, error) {
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
func (d *API) DocCount(sessionId, podName, name, expr string) (*collection.TableKeyCount, error) {
	keyCount := &collection.TableKeyCount{TableName: name}
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return keyCount, ErrUserNotLoggedIn
	}

	// check if pod open
	if !ui.IsPodOpen(podName) {
		return keyCount, ErrPodNotOpen
	}

	podInfo, err := ui.GetPod().GetPodInfoFromPodMap(podName)
	if err != nil {
		return keyCount, err
	}

	count, err := podInfo.GetDocStore().Count(name, expr)
	if err != nil {
		return keyCount, err
	}
	keyCount.Count = count
	return keyCount, nil
}

// DocPut is a controller function which does all the checks before inserting
// a document in the documentDB.
func (d *API) DocPut(sessionId, podName, name string, value []byte) error {
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
func (d *API) DocGet(sessionId, podName, name, id string) ([]byte, error) {
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
func (d *API) DocDel(sessionId, podName, name, id string) error {
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
func (d *API) DocFind(sessionId, podName, name, expr string, limit int) ([][]byte, error) {
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
func (d *API) DocBatch(sessionId, podName, name string) (*collection.DocBatch, error) {
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
func (d *API) DocBatchPut(sessionId, podName string, doc []byte, docBatch *collection.DocBatch) error {
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
func (d *API) DocBatchWrite(sessionId, podName string, docBatch *collection.DocBatch) error {
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
func (d *API) DocIndexJson(sessionId, podName, name, podFileWithPath string) error {
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
