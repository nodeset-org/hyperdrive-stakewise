package utils

import (
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	sharedconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

const (
	contextMetadataName string = "hd-context"
)

// Context for global settings
type HyperdriveContext struct {
	*InstallationInfo

	// The path to the Hyperdrive user directory
	UserDirPath string

	// The max fee for transactions
	MaxFee float64

	// The max priority fee for transactions
	MaxPriorityFee float64

	// The nonce for the first transaction, if set
	Nonce *big.Int

	// True if debug mode is enabled
	DebugEnabled bool

	// True if this is a secure session
	SecureSession bool

	// The address and URL of the API server
	ApiUrl *url.URL

	// The HTTP trace file if tracing is enabled
	HttpTraceFile *os.File

	// The list of networks options and corresponding settings for Hyperdrive itself
	HyperdriveNetworkSettings []*hdconfig.HyperdriveSettings

	// The list of networks options and corresponding settings for the StakeWise module
	StakeWiseNetworkSettings []*sharedconfig.StakeWiseSettings
}

// Load the network settings
func (c *HyperdriveContext) LoadNetworkSettings() error {
	var err error
	installInfo := NewInstallationInfo()
	c.InstallationInfo = installInfo

	c.HyperdriveNetworkSettings, err = hdconfig.LoadSettingsFiles(installInfo.NetworksDir)
	if err != nil {
		return fmt.Errorf("error loading hyperdrive network settings from path [%s]: %s", installInfo.NetworksDir, err.Error())
	}

	swNetSettingsDir := filepath.Join(installInfo.NetworksDir, hdconfig.ModulesName, sharedconfig.ModuleName)
	c.StakeWiseNetworkSettings, err = sharedconfig.LoadSettingsFiles(swNetSettingsDir)
	if err != nil {
		return fmt.Errorf("error loading stakewise network settings from path [%s]: %s", swNetSettingsDir, err.Error())
	}

	return nil
}

// Get the Hyperdrive context from a CLI context
func GetHyperdriveContext(c *cli.Context) *HyperdriveContext {
	return c.App.Metadata[contextMetadataName].(*HyperdriveContext)
}
