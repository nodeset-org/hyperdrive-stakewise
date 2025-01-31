package hdmodule

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/nodeset-org/hyperdrive-stakewise/shared"
	"github.com/urfave/cli/v2"
)

// Request format for `get-log-file`
type getLogFileRequest struct {
	utils.KeyedRequest

	// The log file source to retrieve
	Source string `json:"source"`
}

// Response format for `get-log-file`
type getLogFileResponse struct {
	// The path to the log file
	Path string `json:"path"`
}

// Handle the `get-log-file` command
func getLogFile(c *cli.Context) error {
	// Get the request
	request, err := utils.HandleKeyedRequest[*getLogFileRequest](c)
	if err != nil {
		return err
	}

	// Get the path
	path := ""
	switch request.Source {
	case "adapter":
		path = utils.AdapterLogFile
	case shared.ServiceContainerName:
		path = shared.ServiceLogFile
	}

	// Create the response
	response := getLogFileResponse{
		Path: path,
	}

	// Marshal it
	bytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling get-log-file response: %w", err)
	}

	// Print it
	fmt.Println(string(bytes))
	return nil
}
