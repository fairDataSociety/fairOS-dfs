package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
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
		rootCmd.SetArgs([]string{"server", "--config", tempDir + string(os.PathSeparator) + ".dfs.yaml", "--dataDir", tempDir + string(os.PathSeparator) + ".fairOS/dfs"})
		err = rootCmd.Execute()
		if err.Error() != "postageBlockId is required to run server" {
			t.Fatal("server should fail")
		}
	})
}
