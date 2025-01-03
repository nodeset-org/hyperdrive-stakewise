package swclient

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

// Wrapper for global configuration
type GlobalConfig struct {
	ExternalIP string

	// Hyperdrive
	Hyperdrive          *config.HyperdriveConfig
	HyperdriveResources *config.MergedResources

	// StakeWise
	StakeWise          *swconfig.StakeWiseConfig
	StakeWiseResources *swconfig.StakeWiseResources
}
