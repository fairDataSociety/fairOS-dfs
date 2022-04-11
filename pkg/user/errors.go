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
	ErrUserAlreadyLoggedIn = errors.New("user already logged in")
	ErrInvalidUserName     = errors.New("invalid user name")
	ErrUserAlreadyPresent  = errors.New("user name already present")
	ErrUserNotLoggedIn     = errors.New("user not logged in")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrAccountNeedsFunding = "account needs funding, fund the account and try again with the mnemonic"
)
