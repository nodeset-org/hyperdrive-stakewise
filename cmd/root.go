/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	config "github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	// Used for flags.
	dataDir string
	cfgFile string
	remove  bool

	rootCmd = &cobra.Command{
		Use:   "hyperdrive-stakewise",
		Short: "A brief description of your application",
		Long: `{::} NodeSet Hyperdrive - StakeWise {::}

This script must be run with root privileges.
	`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	if !hyperdrive.IsRoot() {
		log.Fatal("Please run as root (or with sudo)")
	}

	dirname, err := hyperdrive.CallingUserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	d := filepath.Join(dirname, ".node-data")
	rootCmd.PersistentFlags().StringVarP(&dataDir, "directory", "d", d, "data directory")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "nodeset.env", "config file name")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	config.ConfigFile = cfgFile
	config.SetConfigPath(dataDir)

}
