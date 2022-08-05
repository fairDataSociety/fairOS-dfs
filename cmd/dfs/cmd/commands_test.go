package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
)

func Test_ExecuteCommand(t *testing.T) {
	t.Run("config", func(t *testing.T) {
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)
		rootCmd.SetArgs([]string{"config"})
		Execute()
		_, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("version", func(t *testing.T) {
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)
		rootCmd.SetArgs([]string{"version"})
		Execute()
		_, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("server", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", ".dfs")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)
		rootCmd.SetArgs([]string{"server", "--dataDir", tempDir})
		err = rootCmd.Execute()
		if !errors.Is(err, dfs.ErrBeeClient) {
			t.Fatal("server should fail")
		}
	})
}
