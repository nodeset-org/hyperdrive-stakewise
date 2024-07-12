package swcommon

import (
	"context"
	"fmt"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"
)

func (sp *stakeWiseServiceProvider) RequireStakewiseWalletReady(ctx context.Context, status wallet.WalletStatus) error {
	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Check if the wallet files exist
	exists, err := sp.wallet.CheckIfStakewiseWalletExists()
	if exists {
		return nil
	}
	if err != nil {
		logger.Error("Error checking if Stakewise wallet exists", log.Err(err))
	}

	// Check if the Hyperdrive wallet is ready
	err = services.CheckIfWalletReady(status)
	if err != nil {
		return fmt.Errorf("hyperdrive wallet not initialized, can't initialize stakewise wallet yet")
	}

	return sp.initializeStakewiseWallet(logger)
}

func (sp *stakeWiseServiceProvider) WaitForStakewiseWallet(ctx context.Context) error {
	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Check if the wallet files exist
	exists, err := sp.wallet.CheckIfStakewiseWalletExists()
	if exists {
		return nil
	}
	if err != nil {
		logger.Error("Error checking if Stakewise wallet exists", log.Err(err))
	}

	// Wait for the Hyperdrive wallet first, then initialize the Stakewise one
	err = sp.WaitForWallet(ctx)
	if err != nil {
		return err
	}
	return sp.initializeStakewiseWallet(logger)
}

func (sp *stakeWiseServiceProvider) initializeStakewiseWallet(logger *log.Logger) error {
	// Get the wallet from Hyperdrive
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

	logger.Info("Stakewise wallet initialized successfully")
	return nil
}
