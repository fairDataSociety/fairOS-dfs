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
	//ErrBlankPodName
	ErrBlankPodName = errors.New("pod name cannot be blank")
	//ErrInvalidPodName
	ErrInvalidPodName = errors.New("pod does not exist")
	//ErrTooLongPodName
	ErrTooLongPodName = errors.New("pod name too long")
	//ErrPodAlreadyExists
	ErrPodAlreadyExists = errors.New("pod already exists")
	//ErrForkAlreadyExists
	ErrForkAlreadyExists = errors.New("pod with fork name already exists")
	//ErrMaxPodsReached
	ErrMaxPodsReached = errors.New("max number of pods reached")
	//ErrPodNotOpened
	ErrPodNotOpened = errors.New("pod not opened")
	//ErrInvalidDirectory
	ErrInvalidDirectory = errors.New("invalid directory name")
	//ErrTooLongDirectoryName
	ErrTooLongDirectoryName = errors.New("directory name too long")
	//ErrInvalidFile
	ErrInvalidFile = errors.New("file does not exist")
	//ErrMaximumPodLimit
	ErrMaximumPodLimit = errors.New("maximum number of pods has reached")
	//ErrBlankPodSharingReference
	ErrBlankPodSharingReference = errors.New("pod sharing reference cannot be blank")
)
