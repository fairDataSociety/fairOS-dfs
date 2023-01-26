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

	"golang.org/x/crypto/sha3"
)

const (
	// AddressLength defines the length for the address
	AddressLength = 20
)

// Address is a simplified implementation of eth address
type Address [AddressLength]byte

// NewAddress returns a new address from the given byte address
func NewAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func (a *Address) String() string {
	return hex.EncodeToString(a[:])
}

// Hex returns the hex of the corresponding address
func (a Address) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	sha := sha3.NewLegacyKeccak256()
	_, err := sha.Write([]byte(unchecksummed))
	if err != nil { // skipcq: TCV-001
		return ""
	}
	sumHash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := sumHash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// ToBytes returns the address bytes
func (a Address) ToBytes() []byte {
	return a[:]
}

// SetBytes sets address bytes
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// bytesToAddress creates a new address from given bytes
func bytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// HexToAddress creates a new address from given hex string
func HexToAddress(s string) Address { return bytesToAddress(fromHex(s)) }

func fromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex2Bytes(s)
}

func hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}
