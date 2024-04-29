package swcommon

import (
	"os"
	"path/filepath"

	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/wallet"
)

func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(status wallet.WalletStatus) error {
	err := services.CheckIfWalletReady(status)
	// No wallet initialized for Hyperdrive
	if err != nil {
		return err
	}

	moduleDir := sp.GetModuleDir()
	walletPath := filepath.Join(moduleDir, swconfig.WalletFilename)

	_, err = os.ReadFile(walletPath)
	// If wallet is not initialized for SW, just initialize it
	if err != nil {
		client := sp.GetHyperdriveClient()
		ethkeyResponse, err := client.Wallet.ExportEthKey()
		if err != nil {
			return err
		}

		// Write the wallet to disk
		ethKey := ethkeyResponse.Data.EthKeyJson
		err = os.WriteFile(walletPath, ethKey, 0600)
		if err != nil {
			return err
		}

		// Write the password to disk
		password := ethkeyResponse.Data.Password
		passwordPath := filepath.Join(moduleDir, swconfig.PasswordFilename)
		err = os.WriteFile(passwordPath, []byte(password), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}
