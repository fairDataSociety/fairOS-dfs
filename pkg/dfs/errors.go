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

import "errors"

var (
	// ErrUserNotLoggedIn indicate the user is not logged-in
	ErrUserNotLoggedIn = errors.New("user not logged in")
	// ErrPodNotOpen indicates pod is not open
	ErrPodNotOpen = errors.New("pod not open")
	// ErrFileNotPresent indicates file is not present
	ErrFileNotPresent     = errors.New("file not present")
	ErrFileAlreadyPresent = errors.New("file already exist with new name")

	errPodAlreadyOpen = errors.New("pod already open")
	ErrBeeClient      = errors.New("could not connect to bee client")
	errEthClient      = errors.New("could not connect to eth backend")
	errReadOnlyPod    = errors.New("operation not permitted: read only pod")
)
