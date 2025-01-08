package command

import (
	"fmt"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/terminal"
	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"
	"github.com/urfave/cli/v2"
)

func initialize(c *cli.Context) error {
	// Get client

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

	// Check wallet status
	_, ready, err := utils.CheckIfWalletReady(hd)
	if err != nil {
		return err
	}
	if !ready {
		return nil
	}

	// Initialize the Stakewise wallet
	swResponse, err := sw.Api.Wallet.Initialize()
	if err != nil {
		return err
	}

	fmt.Printf("Your node wallet has been successfully copied to the Stakewise module with address %s%s%s.", terminal.ColorBlue, swResponse.Data.AccountAddress.Hex(), terminal.ColorReset)
	return nil
}
