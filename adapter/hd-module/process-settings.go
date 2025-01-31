package hdmodule

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
)

// Request format for `process-config`
type processConfigRequest struct {
	utils.KeyedRequest

	// The config instance to process
	Config map[string]any `json:"config"`
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
	// // Get the request
	// request, err := utils.HandleKeyedRequest[*processConfigRequest](c)
	// if err != nil {
	// 	return err
	// }

	// // Get the config
	// cfg := config.NewStakeWiseConfig()
	// err = hdconfig.UnmarshalConfigurationInstanceIntoMetadata(request.Config, cfg)
	// if err != nil {
	// 	return err
	// }

	// // This is where any examples of validation will go when added
	// errors := []string{}

	// // Get the open ports
	// ports := map[string]uint16{}
	// if cfg.ServerConfig.PortMode.Value != config.PortMode_Closed {
	// 	ports[ids.ServerConfigID+"/"+ids.PortModeID] = uint16(cfg.ServerConfig.Port.Value)
	// }

	// // Create the response
	// response := processConfigResponse{
	// 	Errors: errors,
	// 	Ports:  ports,
	// }

	// // Marshal it
	// bytes, err := json.Marshal(response)
	// if err != nil {
	// 	return fmt.Errorf("error marshalling process-config response: %w", err)
	// }

	// // Print it
	// fmt.Println(string(bytes))
	return nil
}
