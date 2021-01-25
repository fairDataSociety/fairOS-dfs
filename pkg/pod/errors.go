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
	ErrInvalidPodName       = errors.New("invalid pod name")
	ErrTooLongPodName       = errors.New("pod name too long")
	ErrPodAlreadyExists     = errors.New("pod already exists")
	ErrMaxPodsReached       = errors.New("max number of pods reached")
	ErrPodNotOpened         = errors.New("pod not opened")
	ErrInvalidDirectory     = errors.New("invalid directory name")
	ErrTooLongDirectoryName = errors.New("directory name too long")
	ErrReadOnlyPod          = errors.New("operation not permitted: read only pod")
)
