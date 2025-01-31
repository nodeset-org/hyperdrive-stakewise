package hdmodule

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
)

// Request format for `set-config`
type setConfigRequest struct {
	utils.KeyedRequest

	// The config instance to process
	Config map[string]any `json:"config"`
}

// Handle the `set-config` command
func setSettings(c *cli.Context) error {
	// // Get the request
	// request, err := utils.HandleKeyedRequest[*setConfigRequest](c)
	// if err != nil {
	// 	return err
	// }

	// // Get the config
	// cfg := config.NewStakeWiseConfig()
	// err = hdconfig.UnmarshalConfigurationInstanceIntoMetadata(request.Config, cfg)
	// if err != nil {
	// 	return err
	// }

	// // Make a config manager
	// cfgMgr, err := config.NewAdapterConfigManager(c)
	// if err != nil {
	// 	return fmt.Errorf("error creating config manager: %w", err)
	// }
	// cfgMgr.AdapterConfig = cfg

	// // Save it
	// err = cfgMgr.SaveConfigToDisk()
	// if err != nil {
	// 	return fmt.Errorf("error saving config: %w", err)
	// }
	return nil
}
