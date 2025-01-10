package hdmodule

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	hdconfig "github.com/nodeset-org/hyperdrive/modules/config"
	"github.com/urfave/cli/v2"
)

func getConfigInstance(c *cli.Context) error {
	// Get the request
	_, err := utils.HandleKeyedRequest[*utils.KeyedRequest](c)
	if err != nil {
		return err
	}

	// Get the config
	cfgMgr, err := config.NewAdapterConfigManager(c)
	if err != nil {
		return fmt.Errorf("error creating config manager: %w", err)
	}
	cfg, err := cfgMgr.LoadConfigFromDisk()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Handle no saved config
	if cfg == nil {
		return fmt.Errorf("config has not been initialized yet")
	}

	// Create the response
	instance := hdconfig.CreateInstanceFromMetadata(cfg)
	bytes, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
