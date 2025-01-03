package config

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
)

// Handles `hd-module` commands
func RegisterCommands(app *cli.App) {
	commands := []*cli.Command{
		{
			Name:    "nodeset",
			Aliases: []string{"ns"},
			Usage:   "Commands for interacting with the module's configuration",
			Subcommands: []*cli.Command{
				{
					Name:    "upload-deposit-data",
					Aliases: []string{"u"},
					Flags: []cli.Flag{
						utils.YesFlag,
					},
					Usage: "Uploads the combined deposit data for all of your validator keys to NodeSet's Stakewise vault, so they can be assigned new deposits.",
					Action: func(c *cli.Context) error {
						utils.ValidateArgCount(c, 0)
						return uploadDepositData(c)
					},
				},
				{
					Name:    "generate-deposit-data",
					Aliases: []string{"g"},
					Flags: []cli.Flag{
						utils.GeneratePubkeyFlag,
						utils.GenerateIndentFlag,
					},
					Usage: "Generates and prints the deposit data for your validators without uploading it to NodeSet. Useful for debugging.",
					Action: func(c *cli.Context) error {
						utils.ValidateArgCount(c, 0)
						return generateDepositData(c)
					},
				},
			},
		},
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Commands for interacting with the module's configuration",
			Subcommands: []*cli.Command{
				{
					Name:    "status",
					Aliases: []string{"s"},
					Usage:   "Get active validators",
					Action: func(c *cli.Context) error {
						return getNodeStatus(c)
					},
				},
			},
		},
		{
			Name:    "validator",
			Aliases: []string{"v"},
			Usage:   "Commands for interacting with the module's configuration",
			Subcommands: []*cli.Command{
				{
					Name:    "exit",
					Aliases: []string{"e"},
					Usage:   "Exit a validator",
					Flags: []cli.Flag{
						utils.PubkeysFlag,
						utils.EpochFlag,
						utils.NoBroadcastFlag,
					},
					Action: func(c *cli.Context) error {
						// Validate args
						utils.ValidateArgCount(c, 0)

						// Run
						return exit(c)
					},
				},
			},
		},
		{
			Name:    "wallet",
			Aliases: []string{"s"},
			Usage:   "Commands for interacting with the module's configuration",
			Subcommands: []*cli.Command{
				{
					Name:    "init",
					Aliases: []string{"i"},
					Usage:   "Clone the node wallet file into a wallet that the Stakewise operator service can use.",
					Action: func(c *cli.Context) error {
						// Validate args
						utils.ValidateArgCount(c, 0)

						// Run
						return initialize(c)
					},
				},
				{
					Name:    "generate-keys",
					Aliases: []string{"g"},
					Usage:   "Generate new validator keys derived from your node wallet.",
					Flags: []cli.Flag{
						utils.YesFlag,
						utils.GenerateKeysCountFlag,
						utils.GenerateKeysNoRestartFlag,
					},
					Action: func(c *cli.Context) error {
						// Validate args
						utils.ValidateArgCount(c, 0)

						// Run
						return generateKeys(c)
					},
				},
				{
					Name:    "claim-rewards",
					Aliases: []string{"cr"},
					Usage:   "Claim rewards",
					Flags:   []cli.Flag{},
					Action: func(c *cli.Context) error {
						// Run
						return claimRewards(c)
					},
				},
			},
		},
	}

	app.Commands = append(app.Commands, commands...)
}
