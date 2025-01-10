package config

import (
	"encoding/json"
	"fmt"

	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/urfave/cli/v2"
)

func generateDepositData(c *cli.Context) error {
	// Get the client
	hd, err := swclient.NewHyperdriveClientFromCtx(c)
	if err != nil {
		return err
	}
	sw, err := swclient.NewStakewiseClientFromCtx(c, hd)
	if err != nil {
		return err
	}
	cfg, _, err := hd.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading Hyperdrive config: %w", err)
	}
	if !cfg.StakeWise.Enabled.Value {
		fmt.Println("The StakeWise module is not enabled in your Hyperdrive configuration.")
		return nil
	}

	// Parse the pubkeys
	pubkeyStrings := c.StringSlice(generatePubkeyFlag.Name)
	pubkeys := make([]beacon.ValidatorPubkey, len(pubkeyStrings))
	for i, pubkeyString := range pubkeyStrings {
		pubkey, err := beacon.HexToValidatorPubkey(pubkeyString)
		if err != nil {
			return fmt.Errorf("error parsing pubkey [%s]: %w", pubkeyString, err)
		}
		pubkeys[i] = pubkey
	}

	// Generate the deposit data
	fmt.Println("Generating deposit data...")
	response, err := sw.Api.Nodeset.GenerateDepositData(pubkeys)
	if err != nil {
		return err
	}

	// Serialize the deposit data
	var bytes []byte
	shouldIndent := c.Bool(generateIndentFlag.Name)
	if shouldIndent {
		bytes, err = json.MarshalIndent(response.Data.Deposits, "", "  ")
	} else {
		bytes, err = json.Marshal(response.Data.Deposits)
	}
	if err != nil {
		return fmt.Errorf("error serializing deposit data: %w", err)
	}

	// Print the deposit data
	fmt.Println("Deposit data:")
	fmt.Println()
	fmt.Println(string(bytes))
	return nil
}
