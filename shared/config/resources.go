package swconfig

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
)

// A collection of network-specific resources and getters for them
type StakewiseResources struct {
	*config.NetworkResources

	// The URL for the NodeSet API server
	NodesetApiUrl string

	// The address of the Stakewise v3 vault contract, the withdrawal address for all Stakewise validators on Beacon.
	// It's also the address of an ERC-20 token on the EL, which is an LST for the NodeSet partner administrating the vault.
	// See https://app.stakewise.io/vault/holesky/<vault_address> for details.
	Vault *common.Address

	// The address of the NodeSet fee recipient, the Stakewise "smoothing pool".
	// See https://github.com/stakewise/v3-core/blob/main/contracts/vaults/ethereum/mev/SharedMevEscrow.sol
	FeeRecipient *common.Address

	// The address of the SplitWarehouse contract used to hold user funds.
	// All node op rewards will live here; to claim them, call `Withdraw`.
	// See https://docs.splits.org/core/warehouse
	SplitWarehouse *common.Address

	// The address of the PullSplit contract used to manage recipients/allocations and distributions
	// See https://docs.splits.org/core/split-v2
	PullSplit *common.Address
}

// Creates a new resource collection for the given network
func newStakewiseResources(network config.Network) *StakewiseResources {
	// Mainnet
	mainnetResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Mainnet),
		NodesetApiUrl:    "https://nodeset.io/api",
		Vault:            config.HexToAddressPtr("0xE2AEECC76839692AEa35a8D119181b14ebf411c9"),
		FeeRecipient:     config.HexToAddressPtr("0x48319f97E5Da1233c21c48b80097c0FB7a20Ff86"),
		SplitWarehouse:   config.HexToAddressPtr("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:        config.HexToAddressPtr("0x6Cc15f76F76326aCe299Ad7b8fdf4693a96E05C1"),
	}

	// Holesky
	holeskyResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Holesky),
		NodesetApiUrl:    "https://nodeset.io/api",
		Vault:            config.HexToAddressPtr("0x646F5285D195e08E309cF9A5aDFDF68D6Fcc51C4"),
		FeeRecipient:     config.HexToAddressPtr("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWarehouse:   config.HexToAddressPtr("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:        config.HexToAddressPtr("0xAefad0Baa37e1BAF14404bcc2c5E91e4B41c929B"),
	}

	// Holesky Dev
	holeskyDevResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Holesky),
		NodesetApiUrl:    "https://staging.nodeset.io/api",
		Vault:            config.HexToAddressPtr("0xf8763855473ce978232bBa37ef90fcFc8aAE10d1"),
		FeeRecipient:     config.HexToAddressPtr("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWarehouse:   config.HexToAddressPtr("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:        config.HexToAddressPtr("0xAefad0Baa37e1BAF14404bcc2c5E91e4B41c929B"),
	}
	holeskyDevResources.Network = hdconfig.Network_HoleskyDev

	switch network {
	case config.Network_Mainnet:
		return mainnetResources
	case config.Network_Holesky:
		return holeskyResources
	case hdconfig.Network_HoleskyDev:
		return holeskyDevResources
	}

	panic(fmt.Sprintf("network %s is not supported", network))
}
