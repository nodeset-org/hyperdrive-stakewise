package swconfig

import (
	"fmt"
	"math/big"

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

	// The splitter proxy used by the NodeSet vault, which distributes rewards to each node op.
	// All Beacon + EL rewards from the Vault and FeeRecipient will end up here at some point, but they will be in the ERC-20 represented by the Vault.
	// Funds here can be sent to SplitMain by calling `Distribute` on SplitMain and passing this address in as part of it.
	// TODO: Find the "share" per node op address
	SplitWallet *common.Address

	// The address of the SplitMain contract, see https://docs.splits.org/core/split for details.
	// All node op rewards will live here; to claim them, call `Withdraw`.
	SplitMain *common.Address

	// The amount of ETH to claim
	ClaimEthAmount *big.Int

	// The list of token addresses to claim
	ClaimTokenList []common.Address
}

// Creates a new resource collection for the given network
func newStakewiseResources(network config.Network) *StakewiseResources {
	// Mainnet
	mainnetResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Mainnet),
		NodesetApiUrl:    "",
		Vault:            nil,
		FeeRecipient:     nil,
		SplitWallet:      nil,
		SplitMain:        nil,
		ClaimEthAmount:   big.NewInt(0),      // 0 => claim all
		ClaimTokenList:   []common.Address{}, // TODO: Get list from Wander
	}

	// Holesky
	holeskyResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Holesky),
		NodesetApiUrl:    "https://staging.nodeset.io/api",
		Vault:            config.HexToAddressPtr("0x646F5285D195e08E309cF9A5aDFDF68D6Fcc51C4"),
		FeeRecipient:     config.HexToAddressPtr("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWallet:      config.HexToAddressPtr("0x6fa066F4A6439B8a1537F2D300809f23bFF7d37D"),
		SplitMain:        config.HexToAddressPtr("0xfC8a305728051367797DADE6Aa0344E0987f5286"),
		ClaimEthAmount:   big.NewInt(0),      // 0 => claim all
		ClaimTokenList:   []common.Address{}, // TODO: Get list from Wander
	}

	// Holesky Dev
	holeskyDevResources := &StakewiseResources{
		NetworkResources: config.NewResources(config.Network_Holesky),
		NodesetApiUrl:    "https://staging.nodeset.io/api",
		Vault:            config.HexToAddressPtr("0xf8763855473ce978232bBa37ef90fcFc8aAE10d1"),
		FeeRecipient:     config.HexToAddressPtr("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWallet:      nil,
		SplitMain:        config.HexToAddressPtr("0xfC8a305728051367797DADE6Aa0344E0987f5286"),
		ClaimEthAmount:   big.NewInt(0),      // 0 => claim all
		ClaimTokenList:   []common.Address{}, // THIS NEEDS TO BE THE VAULT ADDRESS SINCE THAT'S THE ERC20
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
