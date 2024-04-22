package swcommon

import (
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Creates a new Nodeset client
func (sp *StakewiseServiceProvider) RequireStakewiseWalletReady(status wallet.WalletStatus) error {
	err := services.CheckIfWalletReady(status)
	if err != nil {
		return err
	}
	// Check if stakewise wallet is on disk (i.e. implement sp.GetWallet())
	// If wallet does not exist, imeplement wallet init
	return nil
}
