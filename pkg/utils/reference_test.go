package utils

import (
	"encoding/hex"
	"testing"
)

func TestNewReference(t *testing.T) {
	_, err := ParseHexReference("any ref")
	if err == nil {
		t.Fatal("parse should fail")
	}

	swarmRef := "5605d329affb61b438260842059412330e5c2eaa05fd57f5c0ce3d0180be7988"
	b, err := hex.DecodeString(swarmRef)
	if err != nil {
		t.Fatal(err)
	}
	ref := NewReference(b)
	swarmRef2, err := ParseHexReference(ref.String())
	if err != nil {
		t.Fatal(err)
	}
	if swarmRef2.String() != swarmRef {
		t.Fatal("swarm references do not match")
	}
}
