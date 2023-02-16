package test_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription"
	rpcMock "github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/stretchr/testify/assert"
)

func TestSubscription(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()

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
	pod1 := pod.NewPod(mockClient, fd, acc1, tm, logger)
	addr1 := common.HexToAddress(acc1.GetUserAccountInfo().GetAddress().Hex())
	sm := rpcMock.NewMockSubscriptionManager()
	m := subscription.New(pod1, addr1, acc1.GetUserAccountInfo().GetPrivateKey(), ens, sm)
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

	err = m.ListPod(randomLongPodName1, common.HexToAddress(p1.GetPodAddress().Hex()), 1)
	if err != nil {
		t.Fatal(err)
	}

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, tm, logger)
	addr2 := common.HexToAddress(acc2.GetUserAccountInfo().GetAddress().Hex())

	m2 := subscription.New(pod2, addr2, acc2.GetUserAccountInfo().GetPrivateKey(), ens, sm)

	err = m2.RequestSubscription(common.HexToAddress(p1.GetPodAddress().Hex()), addr1)
	if err != nil {
		t.Fatal(err)
	}

	err = m.ApproveSubscription(p1.GetPodName(), common.HexToAddress(p1.GetPodAddress().Hex()), addr2, acc2.GetUserAccountInfo().GetPublicKey())
	if err != nil {
		t.Fatal(err)
	}

	subs, err := m2.GetSubscriptions()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].Name, randomLongPodName1)

	randomLongPodName2, err := utils.GetRandString(64)
	if err != nil {
		t.Fatalf("error creating pod %s", randomLongPodName2)
	}

	p2, err := pod1.CreatePod(randomLongPodName2, "", podPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = m2.RequestSubscription(common.HexToAddress(p2.GetPodAddress().Hex()), addr1)
	if err == nil {
		t.Fatal("pod is not listed")
	}

	err = m.ListPod(randomLongPodName2, common.HexToAddress(p2.GetPodAddress().Hex()), 1)
	if err != nil {
		t.Fatal(err)
	}

	err = m.DelistPod(common.HexToAddress(p2.GetPodAddress().Hex()))
	if err != nil {
		t.Fatal(err)
	}

	err = m2.RequestSubscription(common.HexToAddress(p2.GetPodAddress().Hex()), addr1)
	if err == nil {
		t.Fatal("pod is not listed")
	}

	pi, err := m2.OpenSubscribedPod(common.HexToAddress(p1.GetPodAddress().Hex()), acc1.GetUserAccountInfo().GetPublicKey())
	if err != nil {
		return
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
	fileMeta1 := fileObject.GetFromFileMap("/parentDir/file1")
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
	if fileMeta1.BlockSize != uint32(10) {
		t.Fatalf("invalid block size")
	}
	fileMeta2 := fileObject.GetFromFileMap("/parentDir/file2")
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
	if fileMeta2.BlockSize != uint32(20) {
		t.Fatalf("invalid block size")
	}
}
