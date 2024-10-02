package testing

import (
	"github.com/ethereum/go-ethereum/common"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
)

const (
	// Address of a mock StakeWise vault for testing
	StakeWiseVaultString string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// Returns a new StakewiseResources instance with test network values
func getTestResources(hdResources *hdconfig.MergedResources, deploymentName string) *swconfig.MergedResources {
	return &swconfig.MergedResources{
		MergedResources: hdResources,
		StakeWiseResources: &swconfig.StakeWiseResources{
			DeploymentName: deploymentName,
			Vault:          common.HexToAddress(StakeWiseVaultString),
			FeeRecipient:   common.HexToAddress(""),
			SplitWarehouse: common.HexToAddress(""),
			PullSplit:      common.HexToAddress(""),
		},
	}
}

// Provisions a NetworkSettings instance with updated addresses
func provisionNetworkSettings(networkSettings *config.NetworkSettings) *config.NetworkSettings {
	return networkSettings
}
