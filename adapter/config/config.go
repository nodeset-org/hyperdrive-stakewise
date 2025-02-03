package config

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config/ids"
	"github.com/nodeset-org/hyperdrive-stakewise/shared"
	sharedconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	hdconfig "github.com/nodeset-org/hyperdrive/modules/config"
	"github.com/rocket-pool/node-manager-core/config"
)

type PortMode string

const (
	PortMode_Closed    PortMode = "closed"
	PortMode_Localhost PortMode = "localhost"
	PortMode_External  PortMode = "external"
)

type StakeWiseConfig struct {
	Enabled              hdconfig.BoolParameter
	ApiPort              hdconfig.UintParameter
	VerifyDepositsRoot   hdconfig.BoolParameter
	DaemonContainerTag   hdconfig.StringParameter
	OperatorContainerTag hdconfig.StringParameter
	AdditionalOpFlags    hdconfig.StringParameter

	// TODO: (HN) - Recreate all this stuff
	VcCommon   *config.ValidatorClientCommonConfig
	Lighthouse *config.LighthouseVcConfig
	Lodestar   *config.LodestarVcConfig
	Nimbus     *config.NimbusVcConfig
	Prysm      *config.PrysmVcConfig
	Teku       *config.TekuVcConfig

	// Internal fields
	Version string
	// hdCfg           *hdconfig.HyperdriveConfig
	// networkSettings []*StakeWiseSettings
}

type StakeWiseConfigSettings struct {
	Enabled              bool   `json:"enabled"`
	ApiPort              uint16 `json:"apiPort"`
	VerifyDepositsRoot   bool   `json:"verifyDepositsRoot"`
	DaemonContainerTag   string `json:"daemonContainerTag"`
	OperatorContainerTag string `json:"operatorContainerTag"`
	AdditionalOpFlags    string `json:"additionalOpFlags"`

	// VcCommon   *config.ValidatorClientCommonConfig `json:"vcCommon"`
	// Lighthouse *config.LighthouseVcConfig          `json:"lighthouse"`
	// Lodestar   *config.LodestarVcConfig            `json:"lodestar"`
	// Nimbus     *config.NimbusVcConfig              `json:"nimbus"`
	// Prysm      *config.PrysmVcConfig               `json:"prysm"`
	// Teku       *config.TekuVcConfig                `json:"teku"`

	Version string `json:"Version"`
}

func NewStakeWiseConfig() *StakeWiseConfig {
	cfg := &StakeWiseConfig{}

	cfg.Enabled.ID = hdconfig.Identifier(ids.EnabledID)
	cfg.Enabled.Name = "Enabled"
	cfg.Enabled.Description.Default = "Toggle for enabling the module"
	cfg.Enabled.AffectedContainers = []string{shared.ServiceContainerName}

	cfg.ApiPort.ID = hdconfig.Identifier(ids.ApiPortID)
	cfg.ApiPort.Name = "API Port"
	cfg.ApiPort.Description.Default = "Port to run the Stakewise API server on"
	cfg.ApiPort.AffectedContainers = []string{shared.ServiceContainerName}

	cfg.VerifyDepositsRoot.ID = hdconfig.Identifier(ids.VerifyDepositsRootID)
	cfg.VerifyDepositsRoot.Name = "Verify Deposits Root"
	cfg.VerifyDepositsRoot.Description.Default = "Toggle for verifying deposit data Merkle roots before saving"
	cfg.VerifyDepositsRoot.AffectedContainers = []string{shared.ServiceContainerName}

	cfg.DaemonContainerTag.ID = hdconfig.Identifier(ids.DaemonContainerTagID)
	cfg.DaemonContainerTag.Name = "Daemon Container Tag"
	cfg.DaemonContainerTag.Description.Default = "The Docker Hub tag for the Stakewise daemon"
	cfg.DaemonContainerTag.AffectedContainers = []string{shared.ServiceContainerName}

	cfg.OperatorContainerTag.ID = hdconfig.Identifier(ids.OperatorContainerTagID)
	cfg.OperatorContainerTag.Name = "Operator Container Tag"
	cfg.OperatorContainerTag.Description.Default = "The Docker Hub tag for the Stakewise operator"
	cfg.OperatorContainerTag.AffectedContainers = []string{shared.ServiceContainerName}

	cfg.AdditionalOpFlags.ID = hdconfig.Identifier(ids.AdditionalOpFlagsID)
	cfg.AdditionalOpFlags.Name = "Additional Operator Flags"
	cfg.AdditionalOpFlags.Description.Default = "Custom command line flags"
	cfg.AdditionalOpFlags.AffectedContainers = []string{shared.ServiceContainerName}

	return cfg
}

func (cfg StakeWiseConfig) GetParameters() []hdconfig.IParameter {
	return []hdconfig.IParameter{
		&cfg.Enabled,
		&cfg.ApiPort,
		&cfg.VerifyDepositsRoot,
		&cfg.DaemonContainerTag,
		&cfg.OperatorContainerTag,
		&cfg.AdditionalOpFlags,
	}
}

func (cfg StakeWiseConfig) GetSections() []hdconfig.ISection {
	return []hdconfig.ISection{
		// TODO: Clients need to get added here after porting
	}
}

func CreateInstanceFromNativeConfig(native *sharedconfig.NativeStakeWiseConfigSettings) *StakeWiseConfigSettings {
	instance := &StakeWiseConfigSettings{
		Enabled:              native.Enabled,
		ApiPort:              native.ApiPort,
		VerifyDepositsRoot:   native.VerifyDepositsRoot,
		DaemonContainerTag:   native.DaemonContainerTag,
		OperatorContainerTag: native.OperatorContainerTag,
		AdditionalOpFlags:    native.AdditionalOpFlags,
	}
	return instance
}

func ConvertInstanceToNativeConfig(instance *StakeWiseConfigSettings) *sharedconfig.NativeStakeWiseConfigSettings {
	native := &sharedconfig.NativeStakeWiseConfigSettings{
		Enabled:              instance.Enabled,
		ApiPort:              instance.ApiPort,
		VerifyDepositsRoot:   instance.VerifyDepositsRoot,
		DaemonContainerTag:   instance.DaemonContainerTag,
		OperatorContainerTag: instance.OperatorContainerTag,
		AdditionalOpFlags:    instance.AdditionalOpFlags,
	}
	return native
}
