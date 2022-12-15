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

import "errors"

var (
	//ErrInvalidPodName indicates pod does not exist
	ErrInvalidPodName = errors.New("pod does not exist")
	//ErrTooLongPodName indicates pod name is too long
	ErrTooLongPodName = errors.New("pod name too long")
	//ErrPodAlreadyExists indicates pod already exist
	ErrPodAlreadyExists = errors.New("pod already exists")
	//ErrMaxPodsReached indicates maximum number of pod has been created
	ErrMaxPodsReached = errors.New("max number of pods reached")
	//ErrPodNotOpened indicates pod is not yet opened
	ErrPodNotOpened = errors.New("pod not opened")
	//ErrInvalidDirectory indicates invalid directory name
	ErrInvalidDirectory = errors.New("invalid directory name")
	//ErrTooLongDirectoryName indicates directory name is too long
	ErrTooLongDirectoryName = errors.New("directory name too long")

	//ErrInvalidFile indicates that the file does not exist
	ErrInvalidFile = errors.New("file does not exist")
	//ErrMaximumPodLimit maximum number of pod has been created
	ErrMaximumPodLimit = errors.New("maximum number of pods has reached")
)
