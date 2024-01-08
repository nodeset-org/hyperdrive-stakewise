/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the node services",
	Long:  `If the node is configured to use internal clients, the execution and concensus layer clients will start as well`,
	Run: func(cmd *cobra.Command, args []string) {

		color.HiWhite("Starting node...")

		c, err := hyperdrive.LoadConfig()
		if err != nil {
			log.Fatal("Can't read config:", errors.Join(err, hyperdrive.ErrorCanNotFindConfigFile))
		}

		var composeFile string
		if c.InternalClients {
			composeFile = "-f compose.yaml -f compose.internal.yaml"
		} else {
			composeFile = "-f compose.yaml"
		}
		text := fmt.Sprintf("docker compose %s pull", composeFile)
		log.Info(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Fatal(err)
		}

		text = fmt.Sprintf("docker compose %s up -d", composeFile)
		log.Info(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
