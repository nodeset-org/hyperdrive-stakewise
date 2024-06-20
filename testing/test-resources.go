package testing

import (
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
)

const (
	// Address of a mock StakeWise vault for testing
	StakeWiseVaultString string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// GetTestResources returns a new StakewiseResources instance with test network values
func GetTestResources(networkResources *config.NetworkResources, nodesetUrl string) *swconfig.StakewiseResources {
	return &swconfig.StakewiseResources{
		NetworkResources: networkResources,
		NodesetApiUrl:    nodesetUrl,
		Vault:            config.HexToAddressPtr(StakeWiseVaultString),
		FeeRecipient:     config.HexToAddressPtr(""),
		SplitWarehouse:   config.HexToAddressPtr(""),
		PullSplit:        config.HexToAddressPtr(""),
	}
}
