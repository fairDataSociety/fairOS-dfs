package test_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestNewSharingReference(t *testing.T) {
	_, err := utils.ParseSharingReference("any ref")
	if err == nil {
		t.Fatal("parse should fail")
	}

	swarmRef := "5605d329affb61b438260842059412330e5c2eaa05fd57f5c0ce3d0180be7988"
	now := time.Now().Unix()
	b, err := hex.DecodeString(swarmRef)
	if err != nil {
		t.Fatal(err)
	}
	ref := utils.NewSharingReference(b, now)
	swarmRef2, err := utils.ParseSharingReference(ref.String())
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(swarmRef2.GetRef()) != swarmRef {
		t.Fatal("swarm references do not match")
	}
}
