package user

//
//func TestNew(t *testing.T) {
//	mockClient := mock.NewMockBeeClient()
//	logger := logging.New(io.Discard, 0)
//
//	t.Run("new-user-migrate-invalid-user", func(t *testing.T) {
//		dataDir, err := ioutil.TempDir("", "new")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer os.RemoveAll(dataDir)
//
//		ens := mock2.NewMockNamespaceManager()
//
//		//create user
//		userObject := NewUsers(dataDir, mockClient, ens, logger)
//		username := "user12"
//		password := "password1"
//		_, _, ui, err := userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//		err = userObject.MigrateUser("username_not_present", "", dataDir, password, ui.sessionId, mockClient, ui)
//		if !errors.Is(err, ErrInvalidUserName) {
//			t.Fatal(err)
//		}
//	})
//
//	t.Run("new-user-migrate-invalid-session", func(t *testing.T) {
//		dataDir, err := ioutil.TempDir("", "new")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer os.RemoveAll(dataDir)
//
//		ens := mock2.NewMockNamespaceManager()
//
//		//create user
//		userObject := NewUsers(dataDir, mockClient, ens, logger)
//		username := "user12"
//		password := "password1"
//		_, _, _, err = userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		err = userObject.MigrateUser(username, "", dataDir, password, "asd8989", mockClient, nil)
//		if err == nil {
//			t.Fatalf("invalid sessionId")
//		}
//	})
//
//	t.Run("new-user-migrate", func(t *testing.T) {
//		dataDir, err := ioutil.TempDir("", "new")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer os.RemoveAll(dataDir)
//
//		ens := mock2.NewMockNamespaceManager()
//
//		//create user
//		userObject := NewUsers(dataDir, mockClient, ens, logger)
//		username := "user1"
//		password := "password1"
//		_, mnemonic, ui, err := userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//		pod1 := ui.GetPod()
//		podName1 := "test1"
//
//		pi1, err := pod1.CreatePod(podName1, password, "")
//		if err != nil {
//			t.Fatalf("error creating pod %s : %s", podName1, err.Error())
//		}
//
//		if ui.GetUserName() != "user1" {
//			t.Fatalf("invalid user name")
//		}
//		if ui.GetFeed() == nil || ui.GetAccount() == nil {
//			t.Fatalf("invalid feed or account")
//		}
//		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
//		if err != nil {
//			t.Fatalf("invalid mnemonic")
//		}
//
//		err = userObject.MigrateUser(username, "", dataDir, password, ui.sessionId, mockClient, ui)
//		if err != nil {
//			t.Fatalf("migrate user: %s", err.Error())
//		}
//
//		ui2, _, _, err := userObject.LoginUserV2(username, password, mockClient, "")
//		if err != nil {
//			t.Fatalf("v2 login: %s", err.Error())
//		}
//		pod2 := ui2.GetPod()
//		pi2, err := pod2.OpenPod(podName1, password)
//		if err != nil {
//			t.Fatalf("open pod after migration: %s", err.Error())
//		}
//		if pi1.GetPodAddress() != pi2.GetPodAddress() {
//			t.Fatalf("pod accounts do not match")
//		}
//	})
//
//	t.Run("new-user-migrate-already-migrated", func(t *testing.T) {
//		dataDir, err := ioutil.TempDir("", "new")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer os.RemoveAll(dataDir)
//
//		ens := mock2.NewMockNamespaceManager()
//
//		//create user
//		userObject := NewUsers(dataDir, mockClient, ens, logger)
//		username := "user14"
//		password := "password1"
//		_, _, ui, err := userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		err = userObject.MigrateUser(username, "", dataDir, password, ui.sessionId, mockClient, ui)
//		if err != nil {
//			t.Fatalf("migrate user: %s", err.Error())
//		}
//
//		_, _, ui, err = userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		err = userObject.MigrateUser(username, "", dataDir, password, ui.sessionId, mockClient, ui)
//		if !errors.Is(err, ErrUserAlreadyPresent) {
//			t.Fatal("user already migrated")
//		}
//	})
//
//	t.Run("new-user-migrate-with-pods", func(t *testing.T) {
//		dataDir, err := ioutil.TempDir("", "new")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer os.RemoveAll(dataDir)
//
//		ens := mock2.NewMockNamespaceManager()
//
//		//create user
//		userObject := NewUsers(dataDir, mockClient, ens, logger)
//		username := "user1"
//		password := "password1"
//		_, mnemonic, ui, err := userObject.CreateNewUser(username, password, "", "")
//		if err != nil {
//			t.Fatal(err)
//		}
//		pod1 := ui.GetPod()
//		podName1 := "test1"
//		podName2 := "test2"
//
//		pi1, err := pod1.CreatePod(podName1, password, "")
//		if err != nil {
//			t.Fatalf("error creating pod %s : %s", podName1, err.Error())
//		}
//		pi2, err := pod1.CreatePod(podName2, password, "")
//		if err != nil {
//			t.Fatalf("error creating pod %s : %s", podName1, err.Error())
//		}
//
//		if ui.GetUserName() != "user1" {
//			t.Fatalf("invalid user name")
//		}
//		if ui.GetFeed() == nil || ui.GetAccount() == nil {
//			t.Fatalf("invalid feed or account")
//		}
//		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
//		if err != nil {
//			t.Fatalf("invalid mnemonic")
//		}
//
//		err = userObject.Logout(ui.GetSessionId())
//		if err != nil {
//			t.Fatalf("logout failed: %s", err)
//		}
//
//		loggedIn := userObject.IsUserLoggedIn(ui.sessionId)
//		if loggedIn {
//			t.Fatalf("user logout failed")
//		}
//
//		ui, err = userObject.LoginUser(username, password, dataDir, mockClient, "")
//		if err != nil {
//			t.Fatal("v1 login failed")
//		}
//		err = userObject.MigrateUser(username, "", dataDir, password, ui.sessionId, mockClient, ui)
//		if err != nil {
//			t.Fatalf("migrate user: %s", err.Error())
//		}
//
//		ui2, _, _, err := userObject.LoginUserV2(username, password, mockClient, "")
//		if err != nil {
//			t.Fatalf("v2 login: %s", err.Error())
//		}
//		pod2 := ui2.GetPod()
//		pi3, err := pod2.OpenPod(podName1, password)
//		if err != nil {
//			t.Fatalf("open pod after migration: %s", err.Error())
//		}
//		if pi1.GetPodAddress() != pi3.GetPodAddress() {
//			t.Fatalf("pod accounts do not match")
//		}
//		pi4, err := pod2.OpenPod(podName2, password)
//		if err != nil {
//			t.Fatalf("open pod after migration: %s", err.Error())
//		}
//		if pi2.GetPodAddress() != pi4.GetPodAddress() {
//			t.Fatalf("pod accounts do not match")
//		}
//	})
//
//}
