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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultDir    = filepath.Join(".fairOS", "dfs")
	defaultConfig = ".dfs.yaml"

	cfgFile     string
	beeApi      string
	beeDebugApi string
	verbosity   string
	dataDir     string

	dataDirPath string
	config      = viper.New()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dfs",
	Short: "Decentralised file system over Swarm(https://ethswarm.org/)",
	Long: `dfs is the file system layer of fairOS. It is a thin layer over Swarm.  
It adds features to Swarm that is required by the fairOS to parallelize computation of data. 
It manages the metadata of directories and files created and expose them to higher layers.
It can also be used as a standalone personal, decentralised drive over the internet`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.BindPFlag(optionDFSDataDir, cmd.Flags().Lookup("dataDir"))
		config.BindPFlag(optionBeeApi, cmd.Flags().Lookup("beeApi"))
		config.BindPFlag(optionBeeDebugApi, cmd.Flags().Lookup("beeDebugApi"))
		config.BindPFlag(optionVerbosity, cmd.Flags().Lookup("verbosity"))

		dataDir = config.GetString(optionDFSDataDir)
		beeApi = config.GetString(optionBeeApi)
		beeDebugApi = config.GetString(optionBeeDebugApi)
		verbosity = config.GetString(optionVerbosity)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fairOSdfs := `
  /$$$$$$          /$$            /$$$$$$   /$$$$$$                /$$  /$$$$$$         
 /$$__  $$        |__/           /$$__  $$ /$$__  $$              | $$ /$$__  $$        
| $$  \__//$$$$$$  /$$  /$$$$$$ | $$  \ $$| $$  \__/          /$$$$$$$| $$  \__//$$$$$$$
| $$$$   |____  $$| $$ /$$__  $$| $$  | $$|  $$$$$$  /$$$$$$ /$$__  $$| $$$$   /$$_____/
| $$_/    /$$$$$$$| $$| $$  \__/| $$  | $$ \____  $$|______/| $$  | $$| $$_/  |  $$$$$$ 
| $$     /$$__  $$| $$| $$      | $$  | $$ /$$  \ $$        | $$  | $$| $$     \____  $$
| $$    |  $$$$$$$| $$| $$      |  $$$$$$/|  $$$$$$/        |  $$$$$$$| $$     /$$$$$$$/
|__/     \_______/|__/|__/       \______/  \______/          \_______/|__/    |_______/

`
	fmt.Println(fairOSdfs)
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
	configPath := filepath.Join(home, defaultConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configPath, "config file")

	dataDirPath = filepath.Join(home, defaultDir)
	rootCmd.PersistentFlags().String("dataDir", dataDirPath, "store data in this dir")
	rootCmd.PersistentFlags().String("beeApi", "localhost:1633", "bee host")
	rootCmd.PersistentFlags().String("beeDebugApi", "localhost:1635", "bee port")
	rootCmd.PersistentFlags().String("verbosity", "5", "verbosity level")

	rootCmd.PersistentFlags().MarkDeprecated("beeHost", "run help to check new flags")
	rootCmd.PersistentFlags().MarkDeprecated("beePort", "run help to check new flags")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// check file stat
		if _, err := os.Stat(cfgFile); err != nil {
			// if there is no configFile, write it
			writeConfig(config)
		}
		// Use config file from the flag.
		config.SetConfigFile(cfgFile)
	} else {
		// Find home dir.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// check file stat
		cfgFile = filepath.Join(home, defaultConfig)
		if _, err := os.Stat(cfgFile); err != nil {
			// if there is no configFile, write it
			writeConfig(config)
		}

		config.SetConfigFile(cfgFile)
	}

	config.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", config.ConfigFileUsed())
	}
}

func writeConfig(c *viper.Viper) {
	c.Set(optionCORSAllowedOrigins, []string{})
	c.Set(optionDFSDataDir, dataDirPath)
	c.Set(optionDFSHttpPort, ":9090")
	c.Set(optionDFSPprofPort, ":9091")
	c.Set(optionVerbosity, "info")
	c.Set(optionBeeApi, "http://localhost:1633")
	c.Set(optionBeeDebugApi, "http://localhost:1635")
	c.Set(optionBeePostageBatchId, "")
	c.Set(optionCookieDomain, "api.fairos.io")

	if err := c.WriteConfigAs(cfgFile); err != nil {
		fmt.Println("failed to write config file")
		os.Exit(1)
	}
}
