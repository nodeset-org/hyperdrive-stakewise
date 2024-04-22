package swcommon

import (
	"fmt"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Creates a new Nodeset client
func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(status wallet.WalletStatus) error {
	err := services.CheckIfWalletReady(status)
	if err != nil {
		return err
	}
	w := sp.GetWallet()
	if w == nil {
		// TODO: Implement wallet init
		fmt.Printf("!!!Wallet not initialized\n")
		return nil
	}
	fmt.Printf("!!!Wallet is ready\n")
	return nil
}
