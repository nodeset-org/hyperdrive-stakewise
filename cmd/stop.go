/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the running containers",
	Long: `Runs a "docker compose down" command to stop any running containers associated with the compose files.
	
The --clean option will run execute "docker compose down --remove-orphans" which will remove any containers not associated with your NodeSet-StakeWise configuration.
`,
	Run: func(cmd *cobra.Command, args []string) {
		clean, _ := cmd.Flags().GetBool("clean")
		color.HiWhite("Shutting down...")

		c, err := hyperdrive.LoadConfig()
		if err != nil {
			log.Fatal("Can't read config:", errors.Join(err, hyperdrive.ErrorCanNotFindConfigFile))
		}

		var removeOrphans string
		if clean {
			prompt := promptui.Select{
				Label: "This will remove all containers, even those not associated with your NodeSet-StakeWise configuration, Are you sure you want to continue?",
				Items: []string{"n", "y"},
			}
			var err error
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			if result == "n" {
				color.HiWhite("Quitting stop")
				return
			}
			if result == "y" {
				removeOrphans = "--remove-orphans"
			}
		}

		var composeFile string
		if c.InternalClients {
			composeFile = "-f compose.yaml -f compose.internal.yaml"
		} else {
			composeFile = "-f compose.yaml"
		}
		text := fmt.Sprintf("docker compose %s down %s", composeFile, removeOrphans)
		log.Info(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	stopCmd.Flags().Bool("clean", false, "WARNING: Using the --clean option for stop will remove any containers not associated with your NodeSet-StakeWise configuration.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
