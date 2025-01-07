package utils

import (
	"fmt"

	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"

	"github.com/rocket-pool/node-manager-core/wallet"
)

func CheckIfWalletReady(hd *swclient.HyperdriveClient) (wallet.WalletStatus, bool, error) {
	// Get & check wallet status
	statusResponse, err := hd.Api.Wallet.Status()
	if err != nil {
		return wallet.WalletStatus{}, false, err
	}
	status := statusResponse.Data.WalletStatus

	// Check if it's already set properly and the wallet has been loaded
	if !wallet.IsWalletReady(status) {
		fmt.Println("The node wallet is not loaded or your node is in read-only mode. Please run `hyperdrive wallet status` for more details.")
		return status, false, nil
	}
	return status, true, nil
}
