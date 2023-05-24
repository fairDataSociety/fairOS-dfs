/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http:// www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import "errors"

var (
	// ErrBlankPodName is returned when a pod name is empty.
	ErrBlankPodName = errors.New("pod name cannot be blank")
	// ErrInvalidPodName is returned when a pod name is invalid.
	ErrInvalidPodName = errors.New("pod does not exist")
	// ErrTooLongPodName is returned when a pod name is too long.
	ErrTooLongPodName = errors.New("pod name too long")
	// ErrPodAlreadyExists is returned when a pod already exists.
	ErrPodAlreadyExists = errors.New("pod already exists")
	// ErrForkAlreadyExists is returned when a pod fork already exists.
	ErrForkAlreadyExists = errors.New("pod with fork name already exists")
	// ErrMaxPodsReached is returned when the maximum number of pods is reached.
	ErrMaxPodsReached = errors.New("max number of pods reached")
	// ErrPodNotOpened is returned when a pod is not opened.
	ErrPodNotOpened = errors.New("pod not opened")
	// ErrInvalidDirectory is returned when a directory is invalid.
	ErrInvalidDirectory = errors.New("invalid directory name")
	// ErrTooLongDirectoryName is returned when a directory name is too long.
	ErrTooLongDirectoryName = errors.New("directory name too long")
	// ErrInvalidFile is returned when a file is invalid.
	ErrInvalidFile = errors.New("file does not exist")
	// ErrMaximumPodLimit is returned when the maximum number of pods is reached.
	ErrMaximumPodLimit = errors.New("maximum number of pods has reached")
	// ErrBlankPodSharingReference is returned when a pod sharing reference is blank.
	ErrBlankPodSharingReference = errors.New("pod sharing reference cannot be blank")
)
