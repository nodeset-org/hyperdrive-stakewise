package utils

import (
	"fmt"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/terminal"
	"github.com/urfave/cli/v2"
)

var (
	YesFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "yes",
		Aliases: []string{"y"},
		Usage:   "Automatically confirm all interactive questions",
	}
	GeneratePubkeyFlag *cli.StringSliceFlag = &cli.StringSliceFlag{
		Name:    "pubkey",
		Aliases: []string{"p"},
		Usage:   "The pubkey of the validator to generate deposit data for. Can be specified multiple times for more than one pubkey. If not specified, deposit data for all validator keys will be generated.",
	}

	GenerateIndentFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "indent",
		Aliases: []string{"i"},
		Usage:   "Specify this to indent (pretty-print) the deposit data output.",
	}
	PubkeysFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "pubkeys",
		Aliases: []string{"p"},
		Usage:   "Comma-separated list of pubkeys (including 0x prefix) to get the exit message for",
	}
	EpochFlag *cli.Uint64Flag = &cli.Uint64Flag{
		Name:    "epoch",
		Aliases: []string{"e"},
		Usage:   "(Optional) the epoch to use when creating the signed exit messages. If not specified, the current chain head will be used.",
	}
	NoBroadcastFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "no-broadcast",
		Aliases: []string{"n"},
		Usage:   "(Optional) pass this flag to skip broadcasting the exit message(s) and print them instead",
	}
	GenerateKeysCountFlag *cli.Uint64Flag = &cli.Uint64Flag{
		Name:    "count",
		Aliases: []string{"c"},
		Usage:   "The number of keys to generate",
	}
	GenerateKeysNoRestartFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:  "no-restart",
		Usage: fmt.Sprintf("Don't automatically restart the Stakewise Operator or Validator Client containers after generating keys. %sOnly use this if you know what you're doing and can restart them manually.%s", terminal.ColorRed, terminal.ColorReset),
	}
	RegisterEmailFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "email",
		Aliases: []string{"e"},
		Usage:   "Email address to register with NodeSet.",
	}
	ForceGasLimitFlag *cli.Uint64Flag = &cli.Uint64Flag{
		Name:    "force-gas-limit",
		Aliases: []string{"fgl"},
		Usage:   fmt.Sprintf("Force a specific gas limit for all transactions produced by the command being run, overriding the gas limit estimated by transaction simulation. %sOnly use this if you know what you're doing.%s", terminal.ColorRed, terminal.ColorReset),
	}
	PrintTxDataFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "print-tx-data",
		Aliases: []string{"pt"},
		Usage:   "Print any TX data for transactions without signing or submitting them. Useful for masquerade mode or offline wallet operations.",
	}
	IgnoreTxSimFailureFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "ignore-tx-sim-failure",
		Aliases: []string{"itsf"},
		Usage:   fmt.Sprintf("Ignore any transaction simulation failures and sign / submit transactions even if they will revert. %sThis can result in failed transactions and loss of funds. Only use this if you absolutely know what you're doing.%s", terminal.ColorRed, terminal.ColorReset),
	}
	SignTxOnlyFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "sign-tx-only",
		Aliases: []string{"st"},
		Usage:   "Sign any TXs and print the results, but don't submit it to the network. Useful if you want to save a TX for later or bundle it up with a service like Flashbots.",
	}
	KeyFileFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "secret",
		Aliases: []string{"s"},
		Usage:   "The path to the secret key file for authentication",
		Value:   "/hd/secret",
	}
	ConfigDirFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "config-dir",
		Aliases: []string{"c"},
		Usage:   "The path to the directory for module configuration files",
		Value:   "/hd/config",
	}
	LogDirFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "log-dir",
		Aliases: []string{"l"},
		Usage:   "The path to the directory for module log files",
		Value:   "/hd/logs",
	}
)

func InstantiateFlag[FlagType cli.Flag](prototype FlagType, description string) cli.Flag {
	switch typedProto := any(prototype).(type) {
	case *cli.BoolFlag:
		return &cli.BoolFlag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.Uint64Flag:
		return &cli.Uint64Flag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.StringFlag:
		return &cli.StringFlag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.Float64Flag:
		return &cli.Float64Flag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	default:
		panic("unsupported flag type")
	}
}
