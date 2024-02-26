package test_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/sirupsen/logrus"

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
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc1.GetUserAccountInfo(), mockClient, -1, 0, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	a1 := acc1.GetUserAccountInfo().GetAddress()
	addr1 := common.HexToAddress(a1.Hex())
	nameHash1, err := goens.NameHash(addr1.Hex())
	if err != nil {
		t.Fatal(err)
	}

	sm := mock2.NewMockSubscriptionManager()
	pod1 := pod.NewPod(mockClient, fd, acc1, tm, sm, -1, 0, logger)

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

	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, tm, sm, -1, 0, logger)
	a2 := acc2.GetUserAccountInfo().GetAddress()
	addr2 := common.HexToAddress(a2.Hex())
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

//

func TestEncryption(t *testing.T) {
	pvtSrt := "153cd9f51ceee5418270957c584b2c8de11f64df6fa2189087aaa89b7deea66e"
	VpvtSrt := "cb1b8338e66bd1a94e5a0e69a869e84ada4ef0e7bc18ddd7c7edcb25bcbd6312"
	pvtKey, err := crypto.HexToECDSA(pvtSrt)
	if err != nil {
		t.Fatal(err)
	}

	VpvtKey, err := crypto.HexToECDSA(VpvtSrt)
	if err != nil {
		t.Fatal(err)
	}

	pubKey := pvtKey.PublicKey
	VpubKey := VpvtKey.PublicKey
	crypto.PubkeyToAddress(pubKey)
	fmt.Println(pubKey)
	fmt.Println(crypto.PubkeyToAddress(pubKey).String())
	fmt.Println(crypto.FromECDSAPub(&pubKey))

	data := "This is a test string"
	fmt.Println(VpubKey.Curve)
	a, _ := VpubKey.Curve.ScalarMult(VpubKey.X, VpubKey.Y, pvtKey.D.Bytes())
	secret := sha256.Sum256(a.Bytes())
	fmt.Println(base64.URLEncoding.EncodeToString(secret[:]))
	encData, err := utils.EncryptBytes(secret[:], []byte(data))
	if err != nil {
		t.Fatal(err)
	}

	uEnc := base64.URLEncoding.EncodeToString(encData)
	fmt.Println(uEnc)

	b, _ := pubKey.Curve.ScalarMult(pubKey.X, pubKey.Y, VpvtKey.D.Bytes())
	secretB := sha256.Sum256(b.Bytes())

	data2, err := utils.DecryptBytes(secretB[:], encData)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data2))
}
