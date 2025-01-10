package hdmodule

import (
	//"encoding/json"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	hdconfig "github.com/nodeset-org/hyperdrive/modules/config"
	"github.com/urfave/cli/v2"
)

func getConfigMetadata(c *cli.Context) error {
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

	// Handle no config file by using the default
	if cfg == nil {
		cfg = config.NewExampleConfig()
	}

	// Create the response
	cfgMap := hdconfig.MarshalConfigurationMetadataToMap(cfg)
	bytes, err := json.Marshal(cfgMap)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
