package utils

import (
	"fmt"
	"os"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/terminal"
	"github.com/urfave/cli/v2"
)

// Validate command argument count
func ValidateArgCount(c *cli.Context, expectedCount int) {
	argCount := c.Args().Len()
	if argCount == expectedCount {
		return
	}

	// Handle invalid arg count
	fmt.Fprintf(os.Stderr, "%sIncorrect argument count - expected %d but have %d%s\n\n", terminal.ColorRed, expectedCount, argCount, terminal.ColorReset)
	cli.ShowSubcommandHelpAndExit(c, 1)
}
