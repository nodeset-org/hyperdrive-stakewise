package swcommon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/nodeset-org/hyperdrive-stakewise/shared/keys"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/node-manager-core/wallet"
)

const (
	walletReadyCheckInterval = time.Duration(15) * time.Second
)

func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(ctx context.Context, status wallet.WalletStatus) error {
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

	return sp.initializeStakewiseWallet(ctx)
}

func (sp *StakewiseServiceProvider) WaitForStakewiseWallet(ctx context.Context) error {
	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	var hdWalletReady bool
	for {
		// Get the HD wallet if it's ready
		if !hdWalletReady {
			hdWalletStatus, err := sp.GetHyperdriveClient().Wallet.Status()
			if err != nil {
				return fmt.Errorf("error getting Hyperdrive wallet status: %w", err)
			}

			if services.CheckIfWalletReady(hdWalletStatus.Data.WalletStatus) == nil {
				hdWalletReady = true
			} else {
				logger.Info("Hyperdrive wallet not ready yet",
					slog.Duration(keys.RetryKey, walletReadyCheckInterval),
				)
				if utils.SleepWithCancel(ctx, walletReadyCheckInterval) {
					return nil
				}
				continue
			}
		}

		return sp.initializeStakewiseWallet(ctx)
	}
}

func (sp *StakewiseServiceProvider) initializeStakewiseWallet(ctx context.Context) error {
	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
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

	logger.Info("Stakewise wallet initialized successfully")
	return nil
}
