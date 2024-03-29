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
	defaultConfig = ".dfs.yaml"

	cfgFile   string
	beeApi    string
	verbosity string

	config = viper.New()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dfs",
	Short: "Decentralised file system over Swarm(https://ethswarm.org/)",
	Long: `dfs is the file system layer of fairOS. It is a thin layer over Swarm.  
It adds features to Swarm that is required by the fairOS to parallelize computation of data. 
It manages the metadata of directories and files created and expose them to higher layers.
It can also be used as a standalone personal, decentralised drive over the internet`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.BindPFlag(optionBeeApi, cmd.Flags().Lookup("beeApi")); err != nil {
			return err
		}
		if err := config.BindPFlag(optionVerbosity, cmd.Flags().Lookup("verbosity")); err != nil {
			return err
		}

		beeApi = config.GetString(optionBeeApi)
		verbosity = config.GetString(optionVerbosity)
		return nil
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
|__/     \_______/|__/|__/       \______/  \______/          \_______/|__/    |_______/`
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

	rootCmd.PersistentFlags().String("beeApi", "http://localhost:1633", "full bee api endpoint")
	rootCmd.PersistentFlags().String("verbosity", "trace", "verbosity level")

	rootCmd.PersistentFlags().String("beeDebugApi", "localhost:1635", "full bee debug api endpoint")
	rootCmd.PersistentFlags().String("beeHost", "127.0.0.1", "bee host")
	rootCmd.PersistentFlags().String("beePort", "1633", "bee port")
	rootCmd.PersistentFlags().String("dataDir", "dataDirPath", "store data in this dir")

	if err := rootCmd.PersistentFlags().MarkDeprecated("beeDebugApi", "using debugAPI is not supported in fairOS-dfs server anymore"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.PersistentFlags().MarkDeprecated("beeHost", "run --beeApi, full bee api endpoint"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.PersistentFlags().MarkDeprecated("beePort", "run --beeApi, full bee api endpoint"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.PersistentFlags().MarkDeprecated("dataDir", "dataDir is no longer required"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// check file stat
		if _, err := os.Stat(cfgFile); err != nil {
			// if there is no configFile, write it
			writeConfig()
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
			writeConfig()
		}

		config.SetConfigFile(cfgFile)
	}

	config.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := config.ReadInConfig(); err != nil {
		fmt.Println("config file not found")
		os.Exit(1)
	}
}

func writeConfig() {
	c := viper.New()
	c.Set(optionCORSAllowedOrigins, defaultCORSAllowedOrigins)
	c.Set(optionDFSHttpPort, defaultDFSHttpPort)
	c.Set(optionDFSPprofPort, defaultDFSPprofPort)
	c.Set(optionVerbosity, defaultVerbosity)
	c.Set(optionBeeApi, defaultBeeApi)
	c.Set(optionBeePostageBatchId, "")
	c.Set(optionBeeRedundancyLevel, 0)
	c.Set(optionCookieDomain, defaultCookieDomain)

	if err := c.WriteConfigAs(cfgFile); err != nil {
		fmt.Println("failed to write config file", err.Error())
		os.Exit(1)
	}
}
