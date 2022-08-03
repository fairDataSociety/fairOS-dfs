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

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// SharingReference is used for sharing files
type SharingReference struct {
	r []byte
	n int64
}

// NewSharingReference creates a reference from swarm reference and time
func NewSharingReference(b []byte, n int64) SharingReference {
	return SharingReference{r: b, n: n}
}

// ParseSharingReference creates a SharingReference from a SharingReference string
func ParseSharingReference(s string) (a SharingReference, err error) {
	refLen := ReferenceLength * 2
	timeLen := len(strconv.FormatInt(time.Now().Unix(), 10))
	if len(s) > refLen+timeLen { // skipcq: TCV-001
		refLen = encryptedRefLength * 2
	}
	if len(s) < refLen+1 { // skipcq: TCV-001
		return a, fmt.Errorf("invalid reference length")
	}
	b, err := hex.DecodeString(s[:refLen])
	if err != nil {
		return a, err
	}
	n, err := strconv.ParseInt(s[refLen:], 10, 64)
	if err != nil {
		return a, err
	}
	return NewSharingReference(b, n), nil
}

func (ref SharingReference) String() string { // skipcq: TCV-001
	refStr := hex.EncodeToString(ref.r)
	numString := strconv.FormatInt(ref.n, 10)
	return refStr + numString
}

func (ref SharingReference) GetRef() []byte {
	return ref.r
}

func (ref SharingReference) GetNonce() int64 {
	return ref.n
}
