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

// exitCmd represents the exit command
var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Exit all validators for the node.",
	Long:  ``,
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

		exit := false
		if c.Network == "mainnet" {
			//TODO: also check if there are any active validators before giving this warning
			//i.e. docker compose up nimbus "check if validators exist"
			color.Red(`DANGER: You are attempting to exit your mainnet validators!
You should ONLY do this if you are sure that you don't want to run these validators anymore.
Once you do this, you must pay the initialization gas fees again if you want to run more validators for this vault.`)
			prompt := promptui.Prompt{
				Label: "Are you sure you want to continue? You must type 'I UNDERSTAND' to continue.",
			}
			var err error
			result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}

			if result == "I UNDERSTAND" {
				color.Red("THIS IS YOUR FINAL WARNING! Are you absolutely sure that you want to exit all of your validators for this mainnet vault configuration (%s)?", c.Vault)
				prompt := promptui.Prompt{
					Label: "You must type 'EXIT EVERYTHING' to continue.",
				}
				var err error
				result, err := prompt.Run()
				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					log.Fatal(err)
				}
				if result == "EXIT EVERYTHING" {
					exit = true
				}
			}
		} else {
			prompt := promptui.Select{
				Label: "Are you sure you want to exit all of your validators for this testnet vault configuration?",
				Items: []string{"n", "y"},
			}
			var err error
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			if result == "y" {
				exit = true
			}
		}

		if !exit {
			color.HiWhite("exit cancelled...")
		}

		if exit {
			color.HiRed("Exiting Validators...")

			text := fmt.Sprintf("docker compose %s run stakewise src/main.py validators-exit --vault %s --consensus-endpoints=%s:%s", composeFile, c.Vault, c.ConsensusClientURL, c.ConsensusClientAPIPort)
			log.Info(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Fatal(err)
			}
			color.HiWhite("Finished exiting validators")
		}

	},
}

func init() {
	rootCmd.AddCommand(exitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
