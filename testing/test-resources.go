package testing

import (
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
)

const (
	// Address of a mock StakeWise vault for testing
	StakeWiseVaultString string = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	SplitWarehouseString string = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
)

// GetTestResources returns a new StakewiseResources instance with test network values
func GetTestResources(networkResources *config.NetworkResources, nodesetUrl string) *swconfig.StakewiseResources {
	return &swconfig.StakewiseResources{
		NetworkResources: networkResources,
		NodesetApiUrl:    nodesetUrl,
		Vault:            config.HexToAddressPtr(StakeWiseVaultString),
		FeeRecipient:     config.HexToAddressPtr(""),
		SplitWarehouse:   config.HexToAddressPtr(SplitWarehouseString),
		PullSplit:        config.HexToAddressPtr(""),
	}
}
