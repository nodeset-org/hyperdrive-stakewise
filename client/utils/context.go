package utils

import (
	"math/big"
	"net/url"
	"os"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/urfave/cli/v2"
)

const (
	contextMetadataName string = "hd-context"
)

// Holds information about Hyperdrive's installation on the system
type InstallationInfo struct {
	// The system path for Hyperdrive scripts used in the Docker containers
	ScriptsDir string

	// The system path for Hyperdrive templates
	TemplatesDir string

	// The system path for the source files to put in the user's override directory
	OverrideSourceDir string

	// The system path for built-in network settings and resource definitions
	NetworksDir string
}

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
	HyperdriveNetworkSettings []*config.HyperdriveSettings
}

// Get the Hyperdrive context from a CLI context
func GetHyperdriveContext(c *cli.Context) *HyperdriveContext {
	return c.App.Metadata[contextMetadataName].(*HyperdriveContext)
}
