package hdmodule

import (
	"fmt"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	hdconfig "github.com/nodeset-org/hyperdrive/modules/config"
	"github.com/urfave/cli/v2"
)

// Request format for `set-config`
type setConfigRequest struct {
	utils.KeyedRequest

	// The config instance to process
	Config map[string]any `json:"config"`
}

// Handle the `set-config` command
func setConfig(c *cli.Context) error {
	// Get the request
	request, err := utils.HandleKeyedRequest[*setConfigRequest](c)
	if err != nil {
		return err
	}

	// Get the config
	cfg := config.NewExampleConfig()
	err = hdconfig.UnmarshalConfigurationInstanceIntoMetadata(request.Config, cfg)
	if err != nil {
		return err
	}

	// Make a config manager
	cfgMgr, err := config.NewAdapterConfigManager(c)
	if err != nil {
		return fmt.Errorf("error creating config manager: %w", err)
	}
	cfgMgr.AdapterConfig = cfg

	// Save it
	err = cfgMgr.SaveConfigToDisk()
	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}
	return nil
}
