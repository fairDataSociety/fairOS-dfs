package cmd

import (
	"bytes"
	"io"
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
		_, err := io.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("version", func(t *testing.T) {
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)
		rootCmd.SetArgs([]string{"version"})
		Execute()
		_, err := io.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("server-postageBlockId-required", func(t *testing.T) {
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

	t.Run("server-postageBlockId-invalid", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", ".dfs")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)

		rootCmd.SetArgs([]string{"server", "--postageBlockId", "postageBlockId is required to run serverpostageBlockId is required to run server", "--config", tempDir + string(os.PathSeparator) + ".dfs.yaml", "--dataDir", tempDir + string(os.PathSeparator) + ".fairOS/dfs"})
		err = rootCmd.Execute()
		if err.Error() != "postageBlockId is invalid" {
			t.Fatal("server should fail")
		}
	})

	t.Run("server-rpc-err", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", ".dfs")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)

		rootCmd.SetArgs([]string{"server", "--postageBlockId", "c108266827eb7ba357797de2707bea00446919346b51954f773560b79765d552", "--config", tempDir + string(os.PathSeparator) + ".dfs.yaml", "--dataDir", tempDir + string(os.PathSeparator) + ".fairOS/dfs"})
		err = rootCmd.Execute()
		if err.Error() != "rpc endpoint is missing" {
			t.Fatal("server should fail")
		}
	})

	t.Run("server-ens-err", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", ".dfs")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)

		rootCmd.SetArgs([]string{"server", "--rpc", "http://localhost:1633", "--postageBlockId", "c108266827eb7ba357797de2707bea00446919346b51954f773560b79765d552", "--config", tempDir + string(os.PathSeparator) + ".dfs.yaml", "--dataDir", tempDir + string(os.PathSeparator) + ".fairOS/dfs"})
		err = rootCmd.Execute()
		if err.Error() != "ens provider domain is missing" {
			t.Fatal("server should fail")
		}
	})

	t.Run("server-network-err", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", ".dfs")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		b := bytes.NewBufferString("")
		rootCmd.SetOut(b)

		rootCmd.SetArgs([]string{
			"server",
			"--network",
			"play",
			"--rpc",
			"http://localhost:1633",
			"--postageBlockId",
			"c108266827eb7ba357797de2707bea00446919346b51954f773560b79765d552",
			"--config",
			tempDir + string(os.PathSeparator) + ".dfs.yaml",
			"--dataDir",
			tempDir + string(os.PathSeparator) + ".fairOS/dfs",
		})
		err = rootCmd.Execute()
		if err.Error() != "could not connect to eth backend" {
			t.Fatal("server should fail")
		}
	})
}
