/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Stops and removes any running containers and deletes the data directory including the chain history",
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

		remove := false
		if c.Network == "mainnet" {
			color.Red(`DANGER: You are attempting to reset your configuration for a mainnet vault!
This will require you to resync the chain completely before you can begin validating again, which may take several days.
Remember, if you're offline for too long, you may be kicked out of NodeSet!`)
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
				color.Red("THIS IS YOUR FINAL WARNING! Are you absolutely sure that you want to delete all of your data for this mainnet configuration?")
				prompt := promptui.Prompt{
					Label: "You must type 'DELETE EVERYTHING' to continue.",
				}
				var err error
				result, err := prompt.Run()
				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					log.Fatal(err)
				}
				if result == "DELETE EVERYTHING" {
					remove = true
				}
			}
		} else {
			color.Red("If you proceede you will have to resync your node. This is a testnet configuration, so this operation is safe.")
			prompt := promptui.Select{
				Label: "Are you sure you want to delete your previous configuration completely?",
				Items: []string{"n", "y"},
			}
			var err error
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			if result == "y" {
				remove = true
			}
		}

		if !remove {
			color.HiWhite("Remove cancelled...")
		}

		if remove {
			color.HiRed("Removing previous configuration...")
			color.Red("Shutting down containers...")

			text := fmt.Sprintf("docker compose %s down -v", composeFile)
			log.Info(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Fatal(err)
			}
			color.Red(fmt.Sprintf("Deleting data in %s...", c.DataDir))
			err := os.RemoveAll(c.DataDir)
			if err != nil {
				log.Fatal(err)
			}

			color.HiWhite("Finished removing previous configuration")
		}

	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
