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
		},
	}
}

// Provisions a NetworkSettings instance with updated addresses
func provisionNetworkSettings(networkSettings *config.NetworkSettings) *config.NetworkSettings {
	networkSettings.NetworkResources.DepositContractAddress = common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	networkSettings.NetworkResources.MulticallAddress = common.HexToAddress("0x8464135c8F25Da09e49BC8782676a84730C318bC")
	networkSettings.NetworkResources.BalanceBatcherAddress = common.HexToAddress("0x71C95911E9a5D330f4D621842EC243EE1343292e")
	return networkSettings
}
