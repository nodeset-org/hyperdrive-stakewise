package config

import "github.com/rocket-pool/node-manager-core/config"

// The master configuration struct
type HyperdriveConfig struct {
	// General settings
	Network                  config.Parameter[config.Network]
	ClientMode               config.Parameter[config.ClientMode]
	EnableIPv6               config.Parameter[bool]
	ProjectName              config.Parameter[string]
	ApiPort                  config.Parameter[uint16]
	UserDataPath             config.Parameter[string]
	AutoTxMaxFee             config.Parameter[float64]
	MaxPriorityFee           config.Parameter[float64]
	AutoTxGasThreshold       config.Parameter[float64]
	AdditionalDockerNetworks config.Parameter[string]
	ClientTimeout            config.Parameter[uint16]

	// The Docker Hub tag for the daemon container
	ContainerTag config.Parameter[string]

	// Logging
	Logging *config.LoggerConfig

	// Execution client settings
	LocalExecutionClient    *config.LocalExecutionConfig
	ExternalExecutionClient *config.ExternalExecutionConfig

	// Beacon node settings
	LocalBeaconClient    *config.LocalBeaconConfig
	ExternalBeaconClient *config.ExternalBeaconConfig

	// Fallback clients
	Fallback *config.FallbackConfig

	// Metrics
	Metrics *config.MetricsConfig

	// MEV-Boost
	MevBoost *MevBoostConfig

	// Modules
	Modules map[string]any

	// Internal fields
	Version                 string
	hyperdriveUserDirectory string
	networkSettings         []*HyperdriveSettings
}
