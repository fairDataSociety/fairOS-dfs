package utils

import (
	"encoding/hex"
	"testing"
)

func TestNewAddress(t *testing.T) {
	addStr := "6f55fbfe6770a6b8d353a14045dc69ff1efa094b"
	addHex := "0x6f55fbFe6770a6b8D353A14045dc69ff1EfA094B"
	b, err := hex.DecodeString(addStr)
	if err != nil {
		t.Fatal(err)
	}
	addr := NewAddress(b)

	if addr.String() != addStr {
		t.Fatal("address do not match")
	}
	if addr.Hex() != addHex {
		t.Fatal("address do not match")
	}
}
