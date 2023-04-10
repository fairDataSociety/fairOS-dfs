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

package dir

import "errors"

var (
	// ErrInvalidDirectoryName is returned when the directory name is invalid
	ErrInvalidDirectoryName = errors.New("invalid directory name")
	// ErrTooLongDirectoryName is returned when the directory name is too long
	ErrTooLongDirectoryName = errors.New("too long directory name")
	// ErrDirectoryAlreadyPresent is returned when the directory is already present
	ErrDirectoryAlreadyPresent = errors.New("directory name already present")
	// ErrDirectoryNotPresent is returned when the directory is not present
	ErrDirectoryNotPresent = errors.New("directory not present")
	// ErrInvalidFileOrDirectoryName is returned when the file or directory name is invalid
	ErrInvalidFileOrDirectoryName = errors.New("invalid file or directory name")
)
