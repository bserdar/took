// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/bserdar/took/cfg"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string
var verbose bool

var UserCfg cfg.Configuration
var CommonCfg cfg.Configuration

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "took",
	Short: "Authentication token manager",
	Long:  `Manage tokens.`}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.took.yaml)")

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getConfigFile() string {
	if cfgFile == "" {
		f, err := homedir.Expand("~/.took.yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return f
	}
	return cfgFile
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	CommonCfg = cfg.ReadCommonConfig()
	cfgf := getConfigFile()
	_, err := os.Stat(cfgf)
	if os.IsNotExist(err) && cfgFile == "" {
		f, err := os.Create(cfgf)
		if err != nil {
			log.Fatalf("%s", err)
		}
		f.Close()
	}
	UserCfg = cfg.ReadConfig(getConfigFile())
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
}

func WriteUserConfig() {
	err := cfg.WriteConfig(getConfigFile(), UserCfg)
	if err != nil {
		log.Fatalf("Cannot write config: %s", err)
	}
}
