package testing

import (
	"github.com/ethereum/go-ethereum/common"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

const (
	// Address of a mock StakeWise vault for testing
	StakeWiseVaultString string = "0x57ace215eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// GetTestResources returns a new StakewiseResources instance with test network values
func GetTestResources(hdResources *hdconfig.MergedResources) *swconfig.MergedResources {
	return &swconfig.MergedResources{
		MergedResources: hdResources,
		StakeWiseResources: &swconfig.StakeWiseResources{
			Vault:          common.HexToAddress(StakeWiseVaultString),
			FeeRecipient:   common.HexToAddress(""),
			SplitWarehouse: common.HexToAddress(""),
			PullSplit:      common.HexToAddress(""),
		},
	}
}
