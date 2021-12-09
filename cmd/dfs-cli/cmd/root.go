/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	dfs "github.com/fairdatasociety/fairOS-dfs"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	fdfsHost string
	fdfsPort string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fdfs-cli",
	Short: "Command line interface for fdfs",
	Long:  `This program interacts with a fdfs server with its API and displays the results`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version    : ", dfs.Version)
		fmt.Println("fdfsHost   : ", fdfsHost)
		fmt.Println("fdfsPort   : ", fdfsPort)
		NewPrompt()
		initPrompt()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fdfsCli := `
       /$$  /$$$$$$                           /$$ /$$
      | $$ /$$__  $$                         | $$|__/
  /$$$$$$$| $$  \__//$$$$$$$         /$$$$$$$| $$ /$$
 /$$__  $$| $$$$   /$$_____//$$$$$$ /$$_____/| $$| $$
| $$  | $$| $$_/  |  $$$$$$|______/| $$      | $$| $$
| $$  | $$| $$     \____  $$       | $$      | $$| $$
|  $$$$$$$| $$     /$$$$$$$/       |  $$$$$$$| $$| $$
 \_______/|__/    |_______/         \_______/|__/|__/
`
	fmt.Println(fdfsCli)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	defaultConfig := filepath.Join(home, ".fairOS/dfs-cli.yml")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfig, "config file")

	rootCmd.PersistentFlags().StringVar(&fdfsHost, "fdfsHost", "127.0.0.1", "fdfs host")
	rootCmd.PersistentFlags().StringVar(&fdfsPort, "fdfsPort", "9090", "fdfs port")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home dir.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home dir with name dfs-cli.yml
		viper.AddConfigPath(home)
		viper.SetConfigName(".fairOS/dfs-cli.yml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
