package hdmodule

import (
	"fmt"

	hdconfig "github.com/nodeset-org/hyperdrive/shared/config"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
)

// Request format for `process-config`
type processConfigRequest struct {
	utils.KeyedRequest

	// The config instance to process
	Settings *hdconfig.HyperdriveSettings `json:"settings"`
}

// Response format for `process-config`
type processConfigResponse struct {
	// A list of errors that occurred during processing, if any
	Errors []string `json:"errors"`

	// A list of ports that will be exposed
	Ports map[string]uint16 `json:"ports"`
}

// Handle the `process-config` command
func processSetting(c *cli.Context) error {
	// Get the request
	request, err := utils.HandleKeyedRequest[*processConfigRequest](c)
	if err != nil {
		return err
	}

	// Construct the module settings from the Hyperdrive config
	modInstance, exists := request.Settings.Modules[utils.FullyQualifiedModuleName]
	if !exists {
		return fmt.Errorf("could not find settings for %s", utils.FullyQualifiedModuleName)
	}
	var settings config.StakeWiseConfigSettings
	err = modInstance.DeserializeSettingsIntoKnownType(&settings)
	if err != nil {
		return fmt.Errorf("error loading settings: %w", err)
	}

	// This is where any examples of validation will go when added
	errors := []string{}

	// // Get the open ports
	ports := map[string]uint16{}

	// if cfg.ServerConfig.PortMode.Value != config.PortMode_Closed {
	// 	ports[ids.ServerConfigID+"/"+ids.PortModeID] = uint16(cfg.ServerConfig.Port.Value)
	// }

	// Create the response
	response := processConfigResponse{
		Errors: errors,
		Ports:  ports,
	}

	// Marshal it
	bytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling process-config response: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
