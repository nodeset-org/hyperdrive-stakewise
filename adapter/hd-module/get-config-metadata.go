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
	cfg := config.NewStakeWiseConfig()

	// Create the response
	cfgMap := hdconfig.MarshalConfigurationToMap(cfg)
	bytes, err := json.Marshal(cfgMap)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
