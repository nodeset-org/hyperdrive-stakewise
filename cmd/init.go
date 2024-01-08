/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/manifoldco/promptui"
	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	"github.com/nodeset-org/hyperdrive-stakewise/local"
)

// var Local embed.FS

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initalizes the ~/.node-data/ directory with nodeset.env, compose.yaml and the ec and cc docker files.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("{::} Welcome to the NodeSet config script for StakeWise {::}")

		remove, _ := cmd.Flags().GetBool("remove")
		network, _ := cmd.Flags().GetString("network")
		internalFlag, _ := cmd.Flags().GetString("internal")
		var useInternalClients bool
		ecName, _ := cmd.Flags().GetString("ecname")
		ccName, _ := cmd.Flags().GetString("ccname")
		checkpoint, _ := cmd.Flags().GetBool("checkpoint")

		var err error
		var c hyperdrive.Config
		if network == "" {
			prompt := promptui.Select{
				Label: "Select Network",
				Items: []string{"NodeSet Test Vault (holesky)", "Gravita (mainnet)", "NodeSet Dev Vault (holskey-dev)"},
			}
			var err error
			_, network, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			switch network {
			case "NodeSet Test Vault (holesky)":
				c = hyperdrive.Holskey
			case "NodeSet Dev Vault (holskey-dev)":
				c = hyperdrive.HoleskyDev
			case "Gravita (mainnet)":
				c = hyperdrive.Gravita

			default:
				log.Fatalf("network %s is not avaliable, please choose holskey, holskey-dev or Gravita", network)
			}
		}
		//if not external, provide options to user
		if internalFlag == "" {
			prompt := promptui.Select{
				Label: "How do you want to manage your client configuration?",
				Items: []string{"External (recommended)", "Internal"},
			}
			var err error
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}

			switch result {
			case "External (recommended)":
				useInternalClients = false
			case "Internal":
				useInternalClients = true
			default:
				useInternalClients = true
			}

		} else if strings.ToLower(internalFlag) == "true" {
			useInternalClients = true
		} else if strings.ToLower(internalFlag) == "false" {
			useInternalClients = false
		}

		if ecName == "" && useInternalClients {
			prompt := promptui.Select{
				Label: "Select Execution Client",
				Items: []string{"geth", "nethermind"},
			}
			var err error
			_, ecName, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
		}

		if ccName == "" && useInternalClients {
			prompt := promptui.Select{
				Label: "Select Concensus Client",
				Items: []string{"nimbus", "teku"},
			}
			var err error
			_, ccName, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
		}

		if remove {
			if c.Network == "mainnet" {
				log.Error("remove=true, init script can not remove data dirctory on mainnet, try using the remove command")
			} else {
				err := os.RemoveAll(dataDir)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
		//set the viper config with defaults before overwriting the values below
		c.SetViper()

		if useInternalClients {
			//Interactive prompt for setting ports
			prompt := promptui.Prompt{
				Label:   "Execution Client Port",
				Default: "30303",
			}

			ecPort, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			viper.Set("ECPORT", ecPort)

			prompt = promptui.Prompt{
				Label:   "Concensus Client Port",
				Default: "9000",
			}

			ccPort, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			viper.Set("CCPORT", ccPort)
		}

		if checkpoint && useInternalClients {
			prompt := promptui.Prompt{
				Label:   "Provide Checkpoint Sync URL",
				Default: "https://checkpoint-sync.holesky.ethpandaops.io",
			}
			var err error
			checkpointURL, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			viper.Set("CHECKPOINT", checkpoint)
			viper.Set("CHECKPOINT_URL", checkpointURL)
		}

		dataDir := viper.GetString("DATA_DIR")
		c.DataDir = dataDir
		log.Infof("Writing config to data path: %s", dataDir)
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			log.Error(err)
		}

		//Ensure that nodeset.env contains the correct ECNAME and CCNAME
		if useInternalClients {
			// c.InternalClients = true
			viper.Set("INTERNAL_CLIENTS", true)
			viper.Set("ECNAME", ecName)
			viper.Set("CCNAME", ccName)
			viper.Set("ECURL", fmt.Sprintf("http://%s", ecName))
			viper.Set("CCURL", fmt.Sprintf("http://%s", ccName))
		} else {
			c.InternalClients = false
			viper.Set("ECNAME", "external")
			viper.Set("CCNAME", "external")
			prompt := promptui.Prompt{
				Label:   "Please enter your eth1 (execution) client URL, excluding ports.",
				Default: "http://127.0.0.1",
			}
			var err error
			ecURL, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			viper.Set("ECURL", ecURL)

			prompt = promptui.Prompt{
				Label:   "Please enter your eth2 (consensus) client URL, excluding ports.",
				Default: "http://127.0.0.1",
			}

			ccURL, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			viper.Set("CCURL", ccURL)
		}

		err = c.WriteConfig()
		if err != nil {
			log.Fatal(err)
		}

		//Write the compose file
		err = os.WriteFile(filepath.Join(dataDir, "compose.yaml"), local.Compose, 0766)
		if err != nil {
			log.Fatal(err)
		}

		//Write the compose internal file
		err = os.WriteFile(filepath.Join(dataDir, "compose.internal.yaml"), local.ComposeInternal, 0766)
		if err != nil {
			log.Fatal(err)
		}

		if useInternalClients {
			//Select EL client
			ecCompose, err := local.Clients.ReadFile(fmt.Sprintf("clients/%s.yaml", ecName))
			if err != nil {
				log.Error(err)
			}
			err = os.WriteFile(filepath.Join(dataDir, fmt.Sprintf("%s.yaml", ecName)), ecCompose, 0766)
			if err != nil {
				log.Fatal(err)
			}
			//Select CC client
			ccCompose, err := local.Clients.ReadFile(fmt.Sprintf("clients/%s.yaml", ccName))
			if err != nil {
				log.Error(err)
			}
			err = os.WriteFile(filepath.Join(dataDir, fmt.Sprintf("%s.yaml", ccName)), ccCompose, 0766)
			if err != nil {
				log.Fatal(err)
			}

			//from install.sh
			// prep data directory
			// mkdir $DATA_DIR/$CCNAME-data
			// mkdir $DATA_DIR/stakewise-data
			// chown $callinguser $DATA_DIR/$CCNAME-data
			// chmod 700 $DATA_DIR/$CCNAME-data
			// # v3-operator user is "nobody" for safety since keys are stored there
			// # you will need to use root to access this directory
			// chown nobody $DATA_DIR/stakewise-data
			u, err := hyperdrive.CallingUser()
			if err != nil {
				log.Errorf("error looking up calling user user info: %e", err)
			}
			os.MkdirAll(filepath.Join(dataDir, fmt.Sprintf("%s-data", ccName)), 0700)
			hyperdrive.Chown(filepath.Join(dataDir, fmt.Sprintf("%s-data", ccName)), u)
			os.MkdirAll(filepath.Join(dataDir, fmt.Sprintf("%s-data", ecName)), 0700)
			hyperdrive.Chown(filepath.Join(dataDir, fmt.Sprintf("%s-data", ccName)), u)
		}
		os.MkdirAll(filepath.Join(dataDir, "stakewise-data"), 0700)
		nobody, err := user.Lookup("nobody")
		hyperdrive.Chown(filepath.Join(dataDir, "stakewise-data"), nobody)

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initCmd.Flags().StringP("network", "n", "", "Select the network")
	initCmd.Flags().String("ecname", "", "Select the execution client [geth, nethermind]")
	initCmd.Flags().String("ccname", "", "Select the consensus client [nimbus, teku]")
	initCmd.Flags().String("internal", "", "Manage the client configuration with internal clients")
	initCmd.Flags().BoolP("remove", "r", false, "Remove the existing installation (if any) in the specified data directory before proceeding with the installation.")
	initCmd.Flags().Bool("checkpoint", true, "Sync the consensus client from latest")

}
