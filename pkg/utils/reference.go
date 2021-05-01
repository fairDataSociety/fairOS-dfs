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

package utils

import "encoding/hex"

const (
	ReferenceLength    = 32
	encryptedRefLength = 64
)

type Reference struct {
	R []byte
}

func NewReference(b []byte) Reference {
	return Reference{R: b}
}
func ParseHexReference(s string) (a Reference, err error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return a, err
	}
	return NewReference(b), nil
}
func (ref Reference) String() string {
	return hex.EncodeToString(ref.R)
}
func (ref Reference) Bytes() []byte {
	return ref.R
}
