/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "docker compose logs",
	Long:  `Available args: geth, nimbus, stakewise, ethdo`,
	Run: func(cmd *cobra.Command, args []string) {

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

		var s string
		if len(args) > 0 {
			s = strings.Join(args, " ")
		}
		text := fmt.Sprintf("docker compose %s logs %s -f", composeFile, s)
		log.Info(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Errorf("Avaliable arguments to log [%s %s stakewise ethdo]", c.ExceutionClientName, c.ConsensusClientName)
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
