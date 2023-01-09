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

package test_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestAddress(t *testing.T) {
	buf := make([]byte, 4096)
	_, err := rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	ch, err := utils.NewChunkWithSpan(buf)
	if err != nil {
		t.Fatal(err)
	}

	refBytes := ch.Address().Bytes()
	ref := utils.NewReference(refBytes)
	refHexString := ref.String()
	newRef, err := utils.ParseHexReference(refHexString)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(refBytes, newRef.Bytes()) {
		t.Fatalf("bytes are not equal")
	}
}

func TestChunkLengthWithSpan(t *testing.T) {
	buf := make([]byte, 5000)
	_, err := rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = utils.NewChunkWithSpan(buf)
	if err != nil && err.Error() != "max chunk size exceeded" {
		t.Fatal("error should be \"max chunk size exceeded\"")
	}
}

func TestChunkLengthWithoutSpan(t *testing.T) {
	buf := make([]byte, 6000)
	_, err := rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = utils.NewChunkWithoutSpan(buf)
	if err != nil && err.Error() != "max chunk size exceeded" {
		t.Fatal("error should be \"max chunk size exceeded\"")
	}
}

func TestDecode(t *testing.T) {
	_, err := utils.Decode("")
	if !errors.Is(err, utils.ErrEmptyString) {
		t.Fatal("err should be empty string")
	}

	_, err = utils.Decode("hello")
	if !errors.Is(err, utils.ErrMissingPrefix) {
		t.Fatal("err should be missing prefix")
	}

	addr := "0xhello"
	_, err = utils.Decode(addr)
	if err == nil {
		t.Fatal("err should be \"invalid hex string\"")
	}

	addr = "0x6F55fbFE6770A6b8d353a14045dc69fF1EF"
	_, err = utils.Decode(addr)
	if err != nil && err.Error() != utils.ErrOddLength.Error() {
		t.Fatal("err should be odd length")
	}

	addr = "0x6F55fbFE6770A6b8d353a14045dc69fF1EFa094b"
	_, err = utils.Decode(addr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRandBytes(t *testing.T) {
	b1, err := utils.GetRandBytes(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(b1) != 10 {
		t.Fatal("b1 length should be 10")
	}
	b2, err := utils.GetRandBytes(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(b2) != 10 {
		t.Fatal("b2 length should be 10")
	}
	if bytes.Equal(b1, b2) {
		t.Fatal("b1 and b2 should not be same")
	}
}

func TestGetRandString(t *testing.T) {
	s1, err := utils.GetRandString(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(s1) != 10 {
		t.Fatal("s1 length should be 10")
	}
	s2, err := utils.GetRandString(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(s2) != 10 {
		t.Fatal("s2 length should be 10")
	}
	if s1 == s2 {
		t.Fatal("s1 and s2 should not be same")
	}
}

func TestCombinePathAndFile(t *testing.T) {
	root1 := ""
	root2 := "/root"
	filename := "test.txt"

	path1 := utils.CombinePathAndFile(root1, filename)
	if path1 != "/"+filename {
		t.Fatal("path1 is wrong")
	}

	path2 := utils.CombinePathAndFile(root2, "")
	if path2 != root2 {
		t.Fatal("path2 is wrong")
	}

	path3 := utils.CombinePathAndFile(root2, filename)
	if path3 != "/root/test.txt" {
		t.Fatal("path3 is wrong")
	}
}
