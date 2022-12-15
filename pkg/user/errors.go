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

package user

import "errors"

var (
	// ErrUserAlreadyLoggedIn is returned if username is already logged-in
	ErrUserAlreadyLoggedIn = errors.New("user already logged-in")

	// ErrInvalidUserName is returned if the username is invalid
	ErrInvalidUserName = errors.New("invalid user name")

	// ErrUserNameNotFound is returned if the username is invalid
	ErrUserNameNotFound = errors.New("no user available")

	// ErrUserAlreadyPresent is returned if username is already taken while signup
	ErrUserAlreadyPresent = errors.New("user name already present")

	// ErrUserNotLoggedIn is returned if user is not logged-in
	ErrUserNotLoggedIn = errors.New("user not logged-in")

	// ErrInvalidPassword is returned if password is invalid
	ErrInvalidPassword = errors.New("invalid password")
)
