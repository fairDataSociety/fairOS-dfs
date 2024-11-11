package act_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/act"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/sirupsen/logrus"
)

func TestACT(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))

	accounts := []*account.Account{}
	for i := 0; i < 10; i++ {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accounts = append(accounts, acc)
	}

	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock2.NewMockSubscriptionManager()
	pods := []string{"test1", "test2"}
	acts := []string{"testact1", "testact2"}
	_ = sm
	_ = pods
	t.Run("create-first-act", func(t *testing.T) {
		acc := accounts[0]
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		ownerACT := act.NewACT(mockClient, fd, acc, tm, logger)
		actName := acts[0]
		for i := 1; i < 10; i++ {
			acc := accounts[i]
			_, err := ownerACT.CreateUpdateACT(actName, acc.GetUserAccountInfo().GetPublicKey(), nil)
			if err != nil {
				t.Fatal(err)
			}
			<-time.After(1 * time.Second)
		}
		_, err := ownerACT.GetACT(actName)
		if err != nil {
			t.Fatal(err)
		}

		pubKeys, err := ownerACT.GetGrantees(actName)
		if err != nil {
			t.Fatal(err)
		}
		if len(pubKeys) != 9 {
			t.Fatal("pubkeys not matching")
		}
	})
	t.Run("create-first-then-revoke-act", func(t *testing.T) {
		acc := accounts[0]
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		ownerACT := act.NewACT(mockClient, fd, acc, tm, logger)
		actName := acts[1]
		for i := 1; i < 2; i++ {
			acc := accounts[i]
			accPrivKey, _ := btcec.PrivKeyFromBytes(acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
			_, err := ownerACT.CreateUpdateACT(actName, accPrivKey.PubKey().ToECDSA(), nil)
			if err != nil {
				t.Fatal(err)
			}
			<-time.After(1 * time.Second)

		}
		a, err := ownerACT.GetACT(actName)
		if err != nil {
			t.Fatal(err)
		}
		_, err = swarm.ParseHexAddress(a.GranteesRef)
		if err != nil {
			t.Fatal(err)
		}
		pubKeys, err := ownerACT.GetGrantees(actName)
		if err != nil {
			t.Fatal(err)
		}
		if len(pubKeys) != 1 {
			t.Fatal("pubkeys not matching")
		}

		for i := 1; i < 2; i++ {
			acc := accounts[i]
			_, err := ownerACT.CreateUpdateACT(actName, nil, acc.GetUserAccountInfo().GetPublicKey())
			if err != nil {
				t.Fatal(err)
			}
			<-time.After(1 * time.Second)
		}
		a, err = ownerACT.GetACT(actName)
		if err != nil {
			t.Fatal(err)
		}
		_, err = swarm.ParseHexAddress(a.GranteesRef)
		if err != nil {
			t.Fatal(err)
		}
		pubKeys, err = ownerACT.GetGrantees(actName)
		if err != nil {
			t.Fatal(err)
		}
		if len(pubKeys) != 0 {
			t.Fatal("pubkeys not matching")
		}
	})
	t.Run("create-second-act", func(t *testing.T) {
		acc := accounts[0]
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		ownerACT := act.NewACT(mockClient, fd, acc, tm, logger)
		for _, actName := range acts {
			for i := 1; i < 10; i++ {
				acc := accounts[i]
				_, err := ownerACT.CreateUpdateACT(actName, acc.GetUserAccountInfo().GetPublicKey(), nil)
				if err != nil {
					t.Fatal(err)
				}
				<-time.After(1 * time.Second)
			}
		}
		list, err := ownerACT.GetList()
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 2 {
			t.Fatal("acts not matching")
		}
	})

	t.Run("act-file-upload", func(t *testing.T) {
		ownerAcc := accounts[0]
		fd := feed.New(ownerAcc.GetUserAccountInfo(), mockClient, -1, 0, logger)
		pod1 := pod.NewPod(mockClient, fd, ownerAcc, tm, sm, -1, 0, logger)
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err := pod1.CreatePod(pods[0], "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s: %s", pods[0], err.Error())
		}

		p, err := pod1.OpenPod(pods[0])
		if err != nil {
			t.Fatalf("error opening pod %s: %s", pods[0], err.Error())
		}

		maxfiles := 1
		filePath := "/"
		for i := 1; i <= maxfiles; i++ {
			fileName, _ := utils.GetRandString(100)
			compression := ""
			fileSize := int64(1000)
			blockSize := file.MinBlockSize
			_, err = uploadFile(t, p.GetFile(), filePath, fileName, compression, p.GetPodPassword(), fileSize, blockSize)
			if err != nil {
				t.Fatal(err)
			}
			err = p.GetDirectory().AddEntryToDir("/", p.GetPodPassword(), fileName, true)
			if err != nil {
				t.Fatal(i, err)
			}
		}

		ref, err := pod1.PodShare(pods[0], "")
		if err != nil {
			t.Fatal(err)
		}
		reference, err := swarm.ParseHexAddress(ref)
		if err != nil {
			t.Fatal(err)
		}

		ownerACT := act.NewACT(mockClient, fd, ownerAcc, tm, logger)
		granteeAcc := accounts[1]
		_, err = ownerACT.CreateUpdateACT(acts[0], granteeAcc.GetUserAccountInfo().GetPublicKey(), nil)
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(1 * time.Second)

		respOne, err := ownerACT.GrantAccess(acts[0], reference)
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(1 * time.Second)

		granteeFeed := feed.New(granteeAcc.GetUserAccountInfo(), mockClient, -1, 0, logger)
		granteeACT := act.NewACT(mockClient, granteeFeed, granteeAcc, tm, logger)
		err = granteeACT.SaveGrantedPod(acts[0], respOne)
		if err != nil {
			t.Fatal(err)
		}
		addr, err := granteeACT.GetPodAccess(acts[0])
		if err != nil {
			t.Fatal(err)
		}

		granteePod := pod.NewPod(mockClient, granteeFeed, granteeAcc, tm, sm, -1, 0, logger)
		_, err = granteePod.ReceivePodInfo(utils.NewReference(addr.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(1 * time.Second)

		actAfterRevoke, err := ownerACT.CreateUpdateACT(acts[0], nil, granteeAcc.GetUserAccountInfo().GetPublicKey())
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(time.Second)
		addr, err = swarm.ParseHexAddress(actAfterRevoke.GranteesRef)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ownerACT.GetGrantees(acts[0])
		if err != nil {
			t.Fatal(err)
		}
		_, err = granteeACT.GetPodAccess(acts[0])
		if err == nil {
			t.Fatal("grantee should not have access")
		}
	})

	t.Run("act-content-list", func(t *testing.T) {
		ownerAcc := accounts[0]
		fd := feed.New(ownerAcc.GetUserAccountInfo(), mockClient, -1, 0, logger)
		pod1 := pod.NewPod(mockClient, fd, ownerAcc, tm, sm, -1, 0, logger)
		ownerACT := act.NewACT(mockClient, fd, ownerAcc, tm, logger)
		granteeAcc := accounts[1]
		_, err := ownerACT.CreateUpdateACT(acts[1], granteeAcc.GetUserAccountInfo().GetPublicKey(), nil)
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(1 * time.Second)
		for i := 1; i < 10; i++ {
			podPassword, _ := utils.GetRandString(pod.PasswordLength)
			podname, _ := utils.GetRandString(pod.PasswordLength)
			_, err := pod1.CreatePod(podname, "", podPassword)
			if err != nil {
				t.Fatalf("error creating pod %s: %s", podname, err.Error())
			}
			ref, err := pod1.PodShare(podname, "")
			if err != nil {
				t.Fatal(err)
			}
			reference, err := swarm.ParseHexAddress(ref)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ownerACT.GrantAccess(acts[1], reference)
			if err != nil {
				t.Fatal(err)
			}
			<-time.After(1 * time.Second)
		}
		contents, err := ownerACT.GetContentList(acts[1])
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 9 {
			t.Fatal("contents not matching", len(contents))
		}

	})
	//
	//t.Run("group-member-add", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	g, err := group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//	maxfiles := 10
	//	filePath := "/"
	//	for i := 1; i <= maxfiles; i++ {
	//		fileName, _ := utils.GetRandString(100)
	//		compression := ""
	//		fileSize := int64(1000)
	//		blockSize := file.MinBlockSize
	//		_, err = uploadFile(t, g.GetFile(), filePath, fileName, compression, g.GetPodPassword(), fileSize, blockSize)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		err = g.GetDirectory().AddEntryToDir("/", g.GetPodPassword(), fileName, true)
	//		if err != nil {
	//			t.Fatal(i, err)
	//		}
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	addr := acc2.GetUserAccountInfo().GetAddress()
	//	addrStr := addr.Hex()
	//	ref, err := group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionWrite)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl, logger)
	//	err = group2.AcceptGroupInvite(ref)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	gi, err := group2.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	dirInode, err := gi.GetDirectory().GetInode(gi.GetPodPassword(), filePath)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if len(dirInode.FileOrDirNames) != maxfiles {
	//		t.Fatal("files not present")
	//	}
	//})
	//
	//t.Run("group-member-check-no-permission", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl2 := acl.NewACL(mockClient, fd2, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl2, logger)
	//	groupName2, _ := utils.GetRandString(10)
	//	_, err = group2.CreateGroup(groupName2)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group2.OpenGroup(groupName2)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	perm, err := group2.GetPermission(groupName2)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if perm != acl.PermissionWrite {
	//		t.Fatal("permission does not match")
	//	}
	//
	//	perm1, err := group.GetPermission(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if perm1 != acl.PermissionWrite {
	//		t.Fatal("permission does not match")
	//	}
	//})
	//
	//t.Run("group-member-check-permission", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	addr := acc2.GetUserAccountInfo().GetAddress()
	//	addrStr := addr.Hex()
	//	ref, err := group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionRead)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl, logger)
	//	err = group2.AcceptGroupInvite(ref)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	_, err = group2.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	perm, err := group2.GetPermission(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if perm != acl.PermissionRead {
	//		t.Fatal("permission not read")
	//	}
	//
	//	err = group.UpdatePermission(groupName1, common.HexToAddress(addrStr), acl.PermissionWrite)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	perm, err = group2.GetPermission(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if perm != acl.PermissionWrite {
	//		t.Fatal("permission not write")
	//	}
	//})
	//
	//t.Run("group-member-upload-files", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	addr := acc2.GetUserAccountInfo().GetAddress()
	//	addrStr := addr.Hex()
	//	ref, err := group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionRead)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl, logger)
	//	err = group2.AcceptGroupInvite(ref)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	g, err := group2.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	fileName, _ := utils.GetRandString(100)
	//	compression := ""
	//	fileSize := int64(1000)
	//	blockSize := file.MinBlockSize
	//	_, err = uploadFile(t, g.GetFile(), "/", fileName, compression, g.GetPodPassword(), fileSize, blockSize)
	//	if !errors.Is(err, feed.ErrReadOnlyFeed) {
	//		t.Fatal(err)
	//	}
	//
	//	err = group.UpdatePermission(groupName1, common.HexToAddress(addrStr), acl.PermissionWrite)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	// reopen the group to reload feed with new permission
	//	err = group2.CloseGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	g, err = group2.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	_, err = uploadFile(t, g.GetFile(), "/", fileName, compression, g.GetPodPassword(), fileSize, blockSize)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	err = g.GetDirectory().AddEntryToDir("/", g.GetPodPassword(), fileName, true)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	dirInode, err := g.GetDirectory().GetInode(g.GetPodPassword(), "/")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if len(dirInode.FileOrDirNames) != 1 {
	//		t.Fatal("files not present")
	//	}
	//	if dirInode.FileOrDirNames[0] != "_F_"+fileName {
	//		t.Fatal("file name not correct")
	//	}
	//})
	//
	//t.Run("group-member-remove", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	addr := acc2.GetUserAccountInfo().GetAddress()
	//	addrStr := addr.Hex()
	//	ref, err := group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionRead)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl, logger)
	//	err = group2.AcceptGroupInvite(ref)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = group.RemoveMember(groupName1, common.HexToAddress(addrStr))
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	_, err = group2.OpenGroup(groupName1)
	//	if !errors.Is(err, pod.ErrPermissionDenied) {
	//		t.Fatal(err)
	//	}
	//})
	//
	//t.Run("group-remove", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//
	//	acc2 := account.New(logger)
	//	_, _, err = acc2.CreateUserAccount("")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	addr := acc2.GetUserAccountInfo().GetAddress()
	//	addrStr := addr.Hex()
	//	ref, err := group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionRead)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//
	//	group2 := pod.NewGroup(mockClient, fd2, acc2, mockAcl, logger)
	//	err = group2.AcceptGroupInvite(ref)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = group.RemoveGroup(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	_, err = group2.OpenGroup(groupName1)
	//	if !errors.Is(err, pod.ErrPermissionDenied) {
	//		t.Fatal(err)
	//	}
	//	_, err = group2.OpenGroup(groupName1)
	//	if !errors.Is(err, pod.ErrGroupDoesNotExist) {
	//		t.Fatal(err)
	//	}
	//})
	//
	//t.Run("group-add-multiple-member", func(t *testing.T) {
	//	t.Skip()
	//	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	//	mockAcl := acl.NewACL(mockClient, fd, logger)
	//	group := pod.NewGroup(mockClient, fd, acc, mockAcl, logger)
	//	groupName1, _ := utils.GetRandString(10)
	//	_, err = group.CreateGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error creating group %s: %s", groupName1, err.Error())
	//	}
	//
	//	_, err = group.ListGroup()
	//	if err != nil {
	//		t.Fatalf("error getting groups")
	//	}
	//	_, err = group.OpenGroup(groupName1)
	//	if err != nil {
	//		t.Fatalf("error opening group %s: %s", groupName1, err.Error())
	//	}
	//	userCount := 10
	//	for i := 0; i < userCount; i++ {
	//		acc2 := account.New(logger)
	//		_, _, err = acc2.CreateUserAccount("")
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		addr := acc2.GetUserAccountInfo().GetAddress()
	//		addrStr := addr.Hex()
	//		_, err = group.AddMember(groupName1, common.HexToAddress(addrStr), acc2.GetUserAccountInfo().GetPublicKey(), acl.PermissionWrite)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	users, err := group.GetGroupMembers(groupName1)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if len(users) != userCount+1 {
	//		t.Fatal("users not added")
	//	}
	//})

}
func uploadFile(t *testing.T, fileObject *file.File, filePath, fileName, compression, podPassword string, fileSize int64, blockSize uint32) ([]byte, error) {
	// create a temp file
	fd, err := os.CreateTemp("", fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fd.Name())

	// write contents to file
	content := make([]byte, fileSize)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fd.Write(content); err != nil {
		t.Fatal(err)
	}

	// close file
	uploadFileName := fd.Name()
	err = fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	// open file to upload
	f1, err := os.Open(uploadFileName)
	if err != nil {
		t.Fatal(err)
	}

	// upload  the temp file
	return content, fileObject.Upload(f1, fileName, fileSize, blockSize, 0, filePath, compression, podPassword)
}

func addFilesAndDirectories(t *testing.T, info *pod.Info, pod1 *pod.Pod, podName1, podPassword string) {
	t.Helper()
	dirObject := info.GetDirectory()
	err := dirObject.MkDir("/parentDir", podPassword, 0)
	if err != nil {
		t.Fatal(err)
	}

	node := dirObject.GetDirFromDirectoryMap("/parentDir")
	if pod1.GetName(node) != "parentDir" {
		t.Fatal("dir name mismatch in pod")
	}
	if pod1.GetPath(node) != "/" {
		t.Fatal("dir path mismatch in pod")
	}

	// populate the directory with few directory and files
	err = dirObject.MkDir("/parentDir/subDir1", podPassword, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.MkDir("/parentDir/subDir2", podPassword, 0)
	if err != nil {
		t.Fatal(err)
	}
	fileObject := info.GetFile()
	_, err = uploadFile(t, fileObject, "/parentDir", "file1", "", podPassword, 100, file.MinBlockSize)
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.AddEntryToDir("/parentDir", podPassword, "file1", true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = uploadFile(t, fileObject, "/parentDir", "file2", "", podPassword, 200, file.MinBlockSize)
	if err != nil {
		t.Fatal(err)
	}
	err = dirObject.AddEntryToDir("/parentDir", podPassword, "file2", true)
	if err != nil {
		t.Fatal(err)
	}

	// close the pod
	err = pod1.ClosePod(podName1)
	if err != nil {
		t.Fatal(err)
	}

	// close the pod
	err = pod1.ClosePod(podName1)
	if err == nil {
		t.Fatal("pod should not be open")
	}
}

func getPrivKey(keyNumber int) *ecdsa.PrivateKey {
	var keyHex string

	switch keyNumber {
	case 0:
		keyHex = "a786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa"
	case 1:
		keyHex = "b786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfb"
	case 2:
		keyHex = "c786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfc"
	default:
		panic("Invalid key number")
	}

	data, err := hex.DecodeString(keyHex)
	if err != nil {
		panic(err)
	}

	privKey, err := crypto.DecodeSecp256k1PrivateKey(data)
	if err != nil {
		panic(err)
	}

	return privKey
}
