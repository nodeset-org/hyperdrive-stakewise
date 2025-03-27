package swconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	"github.com/rocket-pool/node-manager-core/config"
	"gopkg.in/yaml.v2"
)

var (
	// Mainnet resources for reference in testing
	MainnetResourcesReference *StakeWiseResources = &StakeWiseResources{
		Vault:          common.HexToAddress("0xE2AEECC76839692AEa35a8D119181b14ebf411c9"),
		FeeRecipient:   common.HexToAddress("0x48319f97E5Da1233c21c48b80097c0FB7a20Ff86"),
		SplitWarehouse: common.HexToAddress("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:      common.HexToAddress("0x6Cc15f76F76326aCe299Ad7b8fdf4693a96E05C1"),

		// https://github.com/stakewise/sw-utils/blob/2d79588a64858d657f7b8a9a520b149727df0359/sw_utils/networks.py
		Keeper:             common.HexToAddress("0x6B5815467da09DaA7DC83Db21c9239d98Bb487b5"),
		KeeperGenesisBlock: big.NewInt(18470089),
	}

	// Holesky resources for reference in testing
	HoleskyResourcesReference *StakeWiseResources = &StakeWiseResources{
		Vault:          common.HexToAddress("0x646F5285D195e08E309cF9A5aDFDF68D6Fcc51C4"),
		FeeRecipient:   common.HexToAddress("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWarehouse: common.HexToAddress("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:      common.HexToAddress("0xAefad0Baa37e1BAF14404bcc2c5E91e4B41c929B"),

		// https://github.com/stakewise/sw-utils/blob/e07c24879082a44cf1d2c3a5de4a6f8db951e717/sw_utils/networks.py
		Keeper:             common.HexToAddress("0x0xB580799Bf7d62721D1a523f0FDF2f5Ed7BA4e259"),
		KeeperGenesisBlock: big.NewInt(215379),
	}

	// Devnet resources for reference in testing
	DevnetResourcesReference *StakeWiseResources = &StakeWiseResources{
		Vault:          common.HexToAddress("0xf8763855473ce978232bBa37ef90fcFc8aAE10d1"),
		FeeRecipient:   common.HexToAddress("0xc98F25BcAA6B812a07460f18da77AF8385be7b56"),
		SplitWarehouse: common.HexToAddress("0x8fb66F38cF86A3d5e8768f8F1754A24A6c661Fb8"),
		PullSplit:      common.HexToAddress("0xAefad0Baa37e1BAF14404bcc2c5E91e4B41c929B"),

		// https://github.com/stakewise/sw-utils/blob/e07c24879082a44cf1d2c3a5de4a6f8db951e717/sw_utils/networks.py
		Keeper:             common.HexToAddress("0x0xB580799Bf7d62721D1a523f0FDF2f5Ed7BA4e259"),
		KeeperGenesisBlock: big.NewInt(215379),
	}

	// Hoodi resources for reference in testing
	HoodiResourcesReference *StakeWiseResources = &StakeWiseResources{
		Vault:          common.HexToAddress("0x0000000000000000000000000000000000000000"),
		FeeRecipient:   common.HexToAddress("0x0000000000000000000000000000000000000000"),
		SplitWarehouse: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		PullSplit:      common.HexToAddress("0x0000000000000000000000000000000000000000"),

		// https://github.com/stakewise/sw-utils/blob/2d79588a64858d657f7b8a9a520b149727df0359/sw_utils/networks.py
		Keeper:             common.HexToAddress("0x0000000000000000000000000000000000000000"),
		KeeperGenesisBlock: big.NewInt(0),
	}
)

// Details for a StakeWise vault
type StakeWiseVault struct {
	// Whether or not the vault is enabled on the node
	Enabled bool `yaml:"enabled" json:"enabled"`

	// The address of the vault contract
	Address common.Address `yaml:"address" json:"address"`

	// The fee recipient to use for the vault
	FeeRecipient common.Address `yaml:"feeRecipient" json:"feeRecipient"`

	// The address of the SplitWarehouse contract used for node operator fund distribution
	SplitWarehouse common.Address `yaml:"splitWarehouse" json:"splitWarehouse"`

	// The address of the PullSplit contract used for node operator fund distribution
	PullSplit common.Address `yaml:"pullSplit" json:"pullSplit"`
}

// Network settings with a field for StakeWise-specific settings
type StakeWiseSettings struct {
	// The unique key used to identify the network in the configuration
	Key config.Network `yaml:"key" json:"key"`

	// Hyperdrive resources for the network
	StakeWiseResources *StakeWiseResources `yaml:"stakeWiseResources" json:"stakeWiseResources"`

	// A collection of default configuration settings to use for the network, which will override
	// the standard "general-purpose" default value for the setting
	DefaultConfigSettings map[string]any `yaml:"defaultConfigSettings,omitempty" json:"defaultConfigSettings,omitempty"`
}

// A collection of network-specific resources and getters for them
type StakeWiseResources struct {
	// The name of the deployment used by nodeset.io
	DeploymentName string `yaml:"deploymentName" json:"deploymentName"`

	// The address of the Stakewise v3 vault contract, the withdrawal address for all Stakewise validators on Beacon.
	// It's also the address of an ERC-20 token on the EL, which is an LST for the NodeSet partner administrating the vault.
	// See https://app.stakewise.io/vault/holesky/<vault_address> for details.
	Vault common.Address `yaml:"vault" json:"vault"`

	// The address of the NodeSet fee recipient, the Stakewise "smoothing pool".
	// See https://github.com/stakewise/v3-core/blob/main/contracts/vaults/ethereum/mev/SharedMevEscrow.sol
	FeeRecipient common.Address `yaml:"feeRecipient" json:"feeRecipient"`

	// The address of the SplitWarehouse contract used to hold user funds.
	// All node op rewards will live here; to claim them, call `Withdraw`.
	// See https://docs.splits.org/core/warehouse
	SplitWarehouse common.Address `yaml:"splitWarehouse" json:"splitWarehouse"`

	// The address of the PullSplit contract used to manage recipients/allocations and distributions
	// See https://docs.splits.org/core/split-v2
	PullSplit common.Address `yaml:"pullSplit" json:"pullSplit"`

	// The address of the Stakewise keeper contract
	Keeper common.Address `yaml:"keeper" json:"keeper"`

	// The block on the network that StakeWise's keeper contract was deployed
	KeeperGenesisBlock *big.Int `yaml:"keeperGenesisBlock" json:"keeperGenesisBlock"`
}

// A merged set of general resources and StakeWise-specific resources for the selected network
type MergedResources struct {
	// General resources
	*hdconfig.MergedResources

	// StakeWise-specific resources
	*StakeWiseResources
}

// Load network settings from a folder
func LoadSettingsFiles(sourceDir string) ([]*StakeWiseSettings, error) {
	// Make sure the folder exists
	_, err := os.Stat(sourceDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("network settings folder [%s] does not exist", sourceDir)
	}

	// Enumerate the dir
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating network settings source folder: %w", err)
	}

	settingsList := []*StakeWiseSettings{}
	for _, file := range files {
		// Ignore dirs and nonstandard files
		if file.IsDir() || !file.Type().IsRegular() {
			continue
		}

		// Load the file
		filename := file.Name()
		ext := filepath.Ext(filename)
		if ext != ".yaml" && ext != ".yml" {
			// Only load YAML files
			continue
		}
		settingsFilePath := filepath.Join(sourceDir, filename)
		bytes, err := os.ReadFile(settingsFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading network settings file [%s]: %w", settingsFilePath, err)
		}

		// Unmarshal the settings
		settings := new(StakeWiseSettings)
		err = yaml.Unmarshal(bytes, settings)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling network settings file [%s]: %w", settingsFilePath, err)
		}
		settingsList = append(settingsList, settings)
	}
	return settingsList, nil
}
