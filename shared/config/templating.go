package swconfig

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
)

func (c *StakeWiseConfig) WalletFilename() string {
	return WalletFilename
}

func (c *StakeWiseConfig) PasswordFilename() string {
	return PasswordFilename
}

func (c *StakeWiseConfig) KeystorePasswordFile() string {
	return KeystorePasswordFile
}

func (c *StakeWiseConfig) DaemonContainerName() string {
	return string(ContainerID_StakewiseDaemon)
}

func (c *StakeWiseConfig) OperatorContainerName() string {
	return string(ContainerID_StakewiseOperator)
}

func (c *StakeWiseConfig) VcContainerName() string {
	return string(ContainerID_StakewiseValidator)
}

func (c *StakeWiseConfig) DepositDataFile() string {
	return DepositDataFile
}

// The tag for the daemon container
func (cfg *StakeWiseConfig) GetDaemonContainerTag() string {
	return cfg.DaemonContainerTag.Value
}

// Get the container tag of the selected VC
func (cfg *StakeWiseConfig) GetVcContainerTag() string {
	bn := cfg.hdCfg.GetSelectedBeaconNode()
	switch bn {
	case config.BeaconNode_Lighthouse:
		return cfg.Lighthouse.ContainerTag.Value
	case config.BeaconNode_Lodestar:
		return cfg.Lodestar.ContainerTag.Value
	case config.BeaconNode_Nimbus:
		return cfg.Nimbus.ContainerTag.Value
	case config.BeaconNode_Prysm:
		return cfg.Prysm.ContainerTag.Value
	case config.BeaconNode_Teku:
		return cfg.Teku.ContainerTag.Value
	default:
		panic(fmt.Sprintf("Unknown Beacon Node %s", bn))
	}
}

// Gets the additional flags of the selected VC
func (cfg *StakeWiseConfig) GetVcAdditionalFlags() string {
	bn := cfg.hdCfg.GetSelectedBeaconNode()
	switch bn {
	case config.BeaconNode_Lighthouse:
		return cfg.Lighthouse.AdditionalFlags.Value
	case config.BeaconNode_Lodestar:
		return cfg.Lodestar.AdditionalFlags.Value
	case config.BeaconNode_Nimbus:
		return cfg.Nimbus.AdditionalFlags.Value
	case config.BeaconNode_Prysm:
		return cfg.Prysm.AdditionalFlags.Value
	case config.BeaconNode_Teku:
		return cfg.Teku.AdditionalFlags.Value
	default:
		panic(fmt.Sprintf("Unknown Beacon Node %s", bn))
	}
}

// Check if any of the services have doppelganger detection enabled
// NOTE: update this with each new service that runs a VC!
func (cfg *StakeWiseConfig) IsDoppelgangerEnabled() bool {
	return cfg.VcCommon.DoppelgangerDetection.Value
}

// Used by text/template to format validator.yml
func (cfg *StakeWiseConfig) Graffiti() (string, error) {
	prefix := cfg.hdCfg.GraffitiPrefix()
	customGraffiti := cfg.VcCommon.Graffiti.Value
	if customGraffiti == "" {
		return prefix, nil
	}
	return fmt.Sprintf("%s (%s)", prefix, customGraffiti), nil
}

func (cfg *StakeWiseConfig) FeeRecipient() string {
	return cfg.resources.FeeRecipient.Hex()
}

func (cfg *StakeWiseConfig) Vault() string {
	return cfg.resources.Vault.Hex()
}

func (cfg *StakeWiseConfig) Network() string {
	return cfg.resources.EthNetworkName
}

func (cfg *StakeWiseConfig) IsEnabled() bool {
	return cfg.Enabled.Value
}
