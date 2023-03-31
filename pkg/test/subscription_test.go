package test_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/stretchr/testify/assert"
	goens "github.com/wealdtech/go-ens/v3"
)

func TestSubscription(t *testing.T) {
	mockClient := mock.NewMockBeeClient()

	logger := logging.New(os.Stdout, 0)
	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc1.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	addr1 := common.HexToAddress(acc1.GetUserAccountInfo().GetAddress().Hex())
	nameHash1, err := goens.NameHash(addr1.Hex())
	if err != nil {
		t.Fatal(err)
	}

	sm := mock2.NewMockSubscriptionManager()
	pod1 := pod.NewPod(mockClient, fd, acc1, tm, sm, logger)

	randomLongPodName1, err := utils.GetRandString(64)
	if err != nil {
		t.Fatalf("error creating pod %s", randomLongPodName1)
	}

	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	p1, err := pod1.CreatePod(randomLongPodName1, "", podPassword)
	if err != nil {
		t.Fatal(err)
	}

	// make root dir so that other directories can be added
	err = p1.GetDirectory().MkRootDir(randomLongPodName1, podPassword, p1.GetPodAddress(), p1.GetFeed())
	if err != nil {
		t.Fatal(err)
	}

	// create some dir and files
	addFilesAndDirectories(t, p1, pod1, randomLongPodName1, podPassword)
	p1, err = pod1.OpenPod(randomLongPodName1)
	if err != nil {
		t.Fatal(err)
	}
	category := [32]byte{}
	err = pod1.ListPodInMarketplace(randomLongPodName1, randomLongPodName1, randomLongPodName1, "", 1, 10, category, nameHash1)
	if err != nil {
		t.Fatal(err)
	}

	// sub
	market, err := sm.GetAllSubscribablePods()
	if err != nil {
		t.Fatal(err)
	}

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, tm, sm, logger)
	addr2 := common.HexToAddress(acc2.GetUserAccountInfo().GetAddress().Hex())
	nameHash2, err := goens.NameHash(addr2.Hex())
	if err != nil {
		t.Fatal(err)
	}
	err = pod2.RequestSubscription(market[0].SubHash, nameHash2)
	if err != nil {
		t.Fatal(err)
	}

	requests, err := sm.GetSubRequests(addr1)
	if err != nil {
		t.Fatal(err)
	}

	if requests[0].SubHash != market[0].SubHash {
		t.Fatal("subhash mismatch")
	}

	err = pod1.ApproveSubscription(p1.GetPodName(), requests[0].RequestHash, acc2.GetUserAccountInfo().GetPublicKey())
	if err != nil {
		t.Fatal(err)
	}

	subs, err := pod2.GetSubscriptions(requests[0].FdpBuyerNameHash)
	if err != nil {
		t.Fatal(err)
	}

	podInfo, err := pod2.GetSubscribablePodInfo(subs[0].SubHash)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(subs), 1)
	assert.Equal(t, podInfo.PodName, randomLongPodName1)

	pi, err := pod2.OpenSubscribedPod(subs[0].UnlockKeyLocation[:], acc1.GetUserAccountInfo().GetPublicKey())
	if err != nil {
		t.Fatal("failed to open subscribed pod")
	}

	dirObject := pi.GetDirectory()

	dirInode1 := dirObject.GetDirFromDirectoryMap("/parentDir/subDir1")
	if dirInode1 == nil {
		t.Fatalf("invalid dir entry")
	}

	if dirInode1.Meta.Path != "/parentDir" {
		t.Fatalf("invalid path entry")
	}
	if dirInode1.Meta.Name != "subDir1" {
		t.Fatalf("invalid dir entry")
	}
	dirInode2 := dirObject.GetDirFromDirectoryMap("/parentDir/subDir2")
	if dirInode2 == nil {
		t.Fatalf("invalid dir entry")
	}
	if dirInode2.Meta.Path != "/parentDir" {
		t.Fatalf("invalid path entry")
	}
	if dirInode2.Meta.Name != "subDir2" {
		t.Fatalf("invalid dir entry")
	}

	fileObject := pi.GetFile()
	fileMeta1 := fileObject.GetInode(podPassword, "/parentDir/file1")
	if fileMeta1 == nil {
		t.Fatalf("invalid file meta")
	}

	if fileMeta1.Path != "/parentDir" {
		t.Fatalf("invalid path entry")
	}
	if fileMeta1.Name != "file1" {
		t.Fatalf("invalid file entry")
	}
	if fileMeta1.Size != uint64(100) {
		t.Fatalf("invalid file size")
	}
	if fileMeta1.BlockSize != file.MinBlockSize {
		t.Fatalf("invalid block size")
	}
	fileMeta2 := fileObject.GetInode(podPassword, "/parentDir/file2")
	if fileMeta2 == nil {
		t.Fatalf("invalid file meta")
	}
	if fileMeta2.Path != "/parentDir" {
		t.Fatalf("invalid path entry")
	}
	if fileMeta2.Name != "file2" {
		t.Fatalf("invalid file entry")
	}
	if fileMeta2.Size != uint64(200) {
		t.Fatalf("invalid file size")
	}
	if fileMeta2.BlockSize != file.MinBlockSize {
		t.Fatalf("invalid block size")
	}

}
