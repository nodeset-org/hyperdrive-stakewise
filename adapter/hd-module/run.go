package hdmodule

import (
	"fmt"
	"os"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/app"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
)

// Request format for `run`
type runRequest struct {
	utils.KeyedRequest

	// The command to run
	Command string `json:"command"`
}

// Handle the `run` command
func run(c *cli.Context) error {
	// Get the request
	request, err := utils.HandleKeyedRequest[*runRequest](c)
	if err != nil {
		return err
	}

	// Prevent recursive calls
	if strings.HasPrefix(request.Command, "hd-module") || strings.HasPrefix(request.Command, "hd") {
		return fmt.Errorf("recursive calls to `run` are not allowed")
	}

	//TODO: fix whatever's breaking the key flag in the sub command (do you need to make a new app?)
	// Run the app with the new command
	args, err := shellquote.Split(request.Command)
	args = append([]string{
		os.Args[0], // Adapter path
		fmt.Sprintf("--%s", utils.KeyFileFlag.Name),
		c.String(utils.KeyFileFlag.Name),
		fmt.Sprintf("--%s", utils.ConfigDirFlag.Name),
		c.String(utils.ConfigDirFlag.Name),
		fmt.Sprintf("--%s", utils.LogDirFlag.Name),
		c.String(utils.LogDirFlag.Name),
	}, args...)
	if err != nil {
		return fmt.Errorf("error parsing command: %w", err)
	}

	// Create and run a new app
	app := app.CreateApp()
	return app.Run(args)
}
