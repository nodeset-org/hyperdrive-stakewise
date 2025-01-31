package hdmodule

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/nodeset-org/hyperdrive-stakewise/shared"
	"github.com/urfave/cli/v2"
)

// Response format for `get-containers`
type getContainersResponse struct {
	// The list of containers owned by this module
	Containers []string `json:"containers"`
}

// Handle the `get-containers` command
func getContainers(c *cli.Context) error {
	// Get the request
	_, err := utils.HandleKeyedRequest[*utils.KeyedRequest](c)
	if err != nil {
		return err
	}

	// Create the response
	response := getContainersResponse{
		Containers: []string{
			shared.ServiceContainerName,
		},
	}

	// Marshal it
	bytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling get-containers response: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
