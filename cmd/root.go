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
	first := false
	cfgf := getConfigFile()
	_, err := os.Stat(cfgf)
	if os.IsNotExist(err) && cfgFile == "" {
		f, err := os.Create(cfgf)
		if err != nil {
			log.Fatalf("%s", err)
		}
		f.Close()
		first = true
	}
	cfg.ReadUserConfig(getConfigFile())
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	if first {
		firstRun()
	}
}

func WriteUserConfig() {
	err := cfg.WriteUserConfig(getConfigFile())
	if err != nil {
		log.Fatalf("Cannot write config: %s", err)
	}
}
