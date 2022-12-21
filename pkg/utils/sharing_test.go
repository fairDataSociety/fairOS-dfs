package utils

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestNewSharingReference(t *testing.T) {
	_, err := ParseSharingReference("any ref")
	if err == nil {
		t.Fatal("parse should fail")
	}

	swarmRef := "5605d329affb61b438260842059412330e5c2eaa05fd57f5c0ce3d0180be7988"
	now := time.Now().Unix()
	b, err := hex.DecodeString(swarmRef)
	if err != nil {
		t.Fatal(err)
	}
	ref := NewSharingReference(b, now)
	swarmRef2, err := ParseSharingReference(ref.String())
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(swarmRef2.GetRef()) != swarmRef {
		t.Fatal("swarm references do not match")
	}
}
