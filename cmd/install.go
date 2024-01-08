/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/nodeset-org/hyperdrive-stakewise/hyperdrive"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		color.HiWhite("{::} Welcome to the NodeSet node installer for StakeWise {::}")

		c, err := hyperdrive.LoadConfig()
		if err != nil {
			log.Fatal("Can't read config:", errors.Join(err, hyperdrive.ErrorCanNotFindConfigFile))
		}
		if c.InternalClients {
			log.Infof("Starting %s at %s", c.ExceutionClientName, c.DataDir)
			log.Info("Generating jwtsecret...")
			text := fmt.Sprintf("docker compose -f compose.yaml -f compose.internal.yaml up -d %s", c.ExceutionClientName)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Fatal(err)
			}

			log.Info("Waiting for jwtsecret...")
			s := spinner.New(spinner.CharSets[11], 100*time.Millisecond) // Build our new spinner
			s.Start()                                                    // Start the spinner
			time.Sleep(5 * time.Second)                                  // Run for some time to simulate work
			s.Stop()

			// Check if the file exists
			if _, err := os.Stat(filepath.Join(c.DataDir, "jwtsecret", "jwtsecret")); err != nil {
				if os.IsNotExist(err) {
					log.Fatal("Could not generate jwtsecret before timeout!")
				} else {
					fmt.Println(err)
				}
			} else {
				log.Infof("jwtsecret exists: %s", filepath.Join(c.DataDir, "jwtsecret", "jwtsecret"))
			}
			u, _ := hyperdrive.CallingUser()
			hyperdrive.Chown(filepath.Join(c.DataDir, "jwtsecret", "jwtsecret"), u)

			if c.Checkpoint && c.Network != "mainnet" {
				log.Info("Performing checkpoint sync...")
				// docker compose -f "$DATA_DIR/compose.yaml" run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
				text = fmt.Sprintf("docker compose  -f compose.yaml -f compose.internal.yaml run -T nimbus trustedNodeSync -d=/home/user/data --network=%s --trusted-node-url=%s --backfill=false", c.Network, c.CheckpointURL)
				log.Infof(text)
				err = c.ExecCommand(text)
				if err != nil {
					log.Error(err)
				}
			}
		}

		// ### setup stakewise operator
		// echo "Pulling latest StakeWise operator binary..."
		// docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master
		log.Info("Pulling latest StakeWise operator binary...")
		log.Infof("docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master")
		err = c.ExecCommand("docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master")

		// 	if [ "$mnemonic" != "" ]; then
		//     echo "supplying a mnemonic is not yet supported, please check back later!"
		//     exit

		//     echo "Recreating StakeWise configuration using existing mnemonic..."
		//     # todo: recover setup using deposit data downloaded from NodeSet API
		//     #docker compose run stakewise src/main.py get-validators-root --deposit-data-file=<DEPOSIT DATA FILE>
		//     docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py recover --network="$NETWORK" --vault="$VAULT" --consensus-endpoints="http://$CCNAME:$CCAPIPORT" --execution-endpoints="http://$ECNAME:$ECAPIPORT" --mnemonic="$mnemonic"
		//     docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-wallet --vault="$VAULT" --mnemonic="$mnemonic"
		// else
		//     echo "Initializing new StakeWise configuration..."
		//     docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py init --network="$NETWORK" --vault="$VAULT" --language=english
		//     docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-keys --vault="$VAULT" --count="$NUMKEYS"
		//     docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-wallet --vault="$VAULT"
		// fi

		var composeFile string
		if c.InternalClients {
			composeFile = "-f compose.yaml -f compose.internal.yaml"
		} else {
			composeFile = "-f compose.yaml"
		}

		recover, _ := cmd.Flags().GetBool("recover")
		if recover {
			log.Fatal("supplying a mnemonic is not yet supported, please check back later!")
			prompt := promptui.Prompt{
				Label: "Provide a mnemonic seed phrase",
			}
			var err error
			mnemonic, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				log.Fatal(err)
			}
			log.Info("Recreating StakeWise configuration using existing mnemonic...")
			text := fmt.Sprintf("docker compose %s run stakewise src/main.py recover --network=%s --vault=%s --consensus-endpoints=http://%s:%s --execution-endpoints=http://%s:%s --mnemonic=%s", composeFile, c.Network, c.Vault, c.ConsensusClientName, c.ConsensusClientAPIPort, c.ExceutionClientName, c.ExceutionClientAPIPort, mnemonic)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Error(err)
			}
			text = fmt.Sprintf("docker compose %s run stakewise src/main.py create-wallet --vault=%s --mnemonic='%s'", composeFile, c.Vault, mnemonic)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Info("Initializing new StakeWise configuration...")
			text := fmt.Sprintf("docker compose %s run -T stakewise src/main.py init --network=%s --vault=%s --language=english", composeFile, c.Network, c.Vault)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Error(err)
			}
			text = fmt.Sprintf("docker compose %s run stakewise src/main.py create-wallet --vault=%s", composeFile, c.Vault)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Error(err)
			}
			text = fmt.Sprintf("docker compose %s run stakewise src/main.py create-keys --vault=%s --count=%s", composeFile, c.Vault, c.NumKeys)
			log.Infof(text)
			err = c.ExecCommand(text)
			if err != nil {
				log.Error(err)
			}
		}

		fmt.Println("Please note that you must have enough Ether in this node wallet to register validators.")
		fmt.Printf("Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n")
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Please send some ETH to the wallet address above (on the %s network), then type 'wallet is funded' to continue...", c.Network),
			Validate: func(input string) error {
				if input != strings.ToLower("wallet is funded") {
					return errors.New("Please type 'wallet is funded")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed with result: %s, %v\n", result, err)
			log.Fatal(err)
		}

		log.Info("Starting node...")

		// pull latest container images
		log.Info("Updating...")
		text := fmt.Sprintf("docker compose %s pull", composeFile)
		log.Infof(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Error(err)
		}

		// start containers
		text = fmt.Sprintf("docker compose %s up -d", composeFile)
		log.Infof(text)
		err = c.ExecCommand(text)
		if err != nil {
			log.Error(err)
		}

		color.HiWhite("{::} Node Started! {::}")

		installText := fmt.Sprintf(`
{::} Installation Complete! {::}

Your new node is started!

We recommend that you check two things from here:
1. Verify that your node is syncing correctly and watch its progress with \"nodeset logs\"
2. Verify the configuration file in your installation directory looks correct:
%s/nodeset.env
		`, c.DataDir)

		color.HiWhite(installText)

	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	installCmd.Flags().Bool("recover", false, "recover a wallet from a mnemonic seed phrase")
}
