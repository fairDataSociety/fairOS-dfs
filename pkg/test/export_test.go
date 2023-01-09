package test_test

/*
func TestExport(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)

	t.Run("export-user", func(t *testing.T) {
		dataDir, err := ioutil.TempDir("", "new")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir)
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := NewUsers(dataDir, mockClient, ens, logger)
		_, _, ui, err := userObject.CreateNewUser("7e4567e7cb003804992eef11fd5c757275a4a", "password1", "", "")
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = userObject.ExportUser(&Info{name: "any_name"})
		if err == nil {
			t.Fatal("user should not be present")
		}
		username, address, err := userObject.ExportUser(ui)
		if err != nil {
			t.Fatal(err)
		}
		if ui.GetUserName() != username {
			t.Fatal("username mismatch")
		}
		if ui.GetAccount().GetAddress(account.UserAccountIndex).Hex() != address {
			t.Fatal("address mismatch")
		}
	})
}
*/
