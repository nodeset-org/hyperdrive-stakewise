package swcommon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/wallet"
)

type WalletData struct {
	Address string `json:"address"`
	Crypto  struct {
		Cipher       string `json:"cipher"`
		Ciphertext   string `json:"ciphertext"`
		Cipherparams struct {
			Iv string `json:"iv"`
		} `json:"cipherparams"`
		Kdf       string `json:"kdf"`
		Kdfparams struct {
			Dklen int    `json:"dklen"`
			Salt  string `json:"salt"`
			N     int    `json:"n"`
			R     int    `json:"r"`
			P     int    `json:"p"`
		} `json:"kdfparams"`
		Mac string `json:"mac"`
	} `json:"crypto"`
	Id      string `json:"id"`
	Version int    `json:"version"`
}

// Creates a new Nodeset client
func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(status wallet.WalletStatus) error {
	err := services.CheckIfWalletReady(status)
	// No wallet initialized for Hyperdrive
	if err != nil {
		return err
	}

	moduleDir := sp.GetModuleDir()
	walletPath := filepath.Join(moduleDir, swconfig.WalletFilename)

	fmt.Printf("!!! wallet status: %v\n", status)
	w := getWalletFromPath(walletPath)
	fmt.Printf("!!! wallet: %v\n", w)
	// If wallet is not initialized for SW, just initialize it
	if w == nil {
		client := sp.GetHyperdriveClient()
		ethkeyResponse, err := client.Wallet.ExportEthKey()
		if err != nil {
			return err
		}
		// Write the wallet to disk
		err = os.WriteFile(walletPath, ethkeyResponse.Data.EthKeyJson, 0600)
		if err != nil {
			return err
		}
		passwordPath := filepath.Join(moduleDir, swconfig.PasswordFilename)
		err = os.WriteFile(passwordPath, []byte(ethkeyResponse.Data.Password), 0600)
		if err != nil {
			return err
		}
		fmt.Printf("!!!Wallet not initialized\n")
	}
	return nil
}

func getWalletFromPath(walletPath string) *WalletData {
	// Read the file from the given path
	data, err := os.ReadFile(walletPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Handle the case where the wallet file does not exist
			fmt.Println("Wallet file not found:", walletPath)
			return nil
		}
		fmt.Printf("Error reading wallet file: %v\n", err)
		return nil
	}

	// Unmarshal the JSON data into the WalletData struct
	var wallet *WalletData
	if err := json.Unmarshal(data, &wallet); err != nil {
		fmt.Printf("Error parsing wallet JSON data: %v\n", err)
		return nil
	}

	return wallet
}
