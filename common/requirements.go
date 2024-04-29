package swcommon

import (
	"context"
	"fmt"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"
)

func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(ctx context.Context, status wallet.WalletStatus) error {
	err := services.CheckIfWalletReady(status)
	// No wallet initialized for Hyperdrive
	if err != nil {
		return err
	}

	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Check if the wallet files exist
	exists, err = sp.wallet.CheckIfStakewiseWalletExists()
	if exists && err == nil {
		return nil
	}
	if err != nil {
		logger.Debug("Error checking if Stakewise wallet exists", log.Err(err))
	}

	// If wallet is not initialized for SW, just initialize it
	logger.Warn("Stakewise wallet not found, initializing now")
	ethkeyResponse, err := sp.GetHyperdriveClient().Wallet.ExportEthKey()
	if err != nil {
		return fmt.Errorf("error getting geth-style keystore from Hyperdrive client: %w", err)
	}
	ethKey := ethkeyResponse.Data.EthKeyJson
	password := ethkeyResponse.Data.Password

	// Write the Stakewise wallet files to disk
	err = sp.wallet.SaveStakewiseWallet(ethKey, password)
	if err != nil {
		return err
	}
	return nil
}
