package testing

import (
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
)

const (
	// Address of a mock StakeWise vault for testing
	StakeWiseVaultString string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// GetTestResources returns a new StakewiseResources instance with test network values
func GetTestResources(hdResources *hdconfig.HyperdriveResources) *swconfig.StakewiseResources {
	return &swconfig.StakewiseResources{
		HyperdriveResources: hdResources,
		Vault:               config.HexToAddressPtr(StakeWiseVaultString),
		FeeRecipient:        config.HexToAddressPtr(""),
		SplitWarehouse:      config.HexToAddressPtr(""),
		PullSplit:           config.HexToAddressPtr(""),
	}
}
