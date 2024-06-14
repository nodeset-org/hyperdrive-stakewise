package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/nodeset-org/hyperdrive-daemon/shared"
	"github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/nodeset-org/hyperdrive-stakewise/server"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	swtasks "github.com/nodeset-org/hyperdrive-stakewise/tasks"
	"github.com/urfave/cli/v2"
)

// Run
func main() {
	// Add logo and attribution to application help template
	attribution := "ATTRIBUTION:\n   Adapted from the Rocket Pool Smart Node (https://github.com/rocketpool/smartnode) with love."
	cli.AppHelpTemplate = fmt.Sprintf("\n%s\n\n%s\n%s\n", shared.Logo, cli.AppHelpTemplate, attribution)
	cli.CommandHelpTemplate = fmt.Sprintf("%s\n%s\n", cli.CommandHelpTemplate, attribution)
	cli.SubcommandHelpTemplate = fmt.Sprintf("%s\n%s\n", cli.SubcommandHelpTemplate, attribution)

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "stakewise-daemon"
	app.Usage = "Hyperdrive Daemon for NodeSet StakeWise Module Management"
	app.Version = shared.HyperdriveVersion
	app.Authors = []*cli.Author{
		{
			Name:  "Nodeset",
			Email: "info@nodeset.io",
		},
	}
	app.Copyright = "(C) 2024 NodeSet LLC"

	moduleDirFlag := &cli.StringFlag{
		Name:     "module-dir",
		Aliases:  []string{"d"},
		Usage:    "The path to the StakeWise module data directory",
		Required: true,
	}
	hyperdriveUrlFlag := &cli.StringFlag{
		Name:    "hyperdrive-url",
		Aliases: []string{"hd"},
		Usage:   "The URL of the Hyperdrive API",
		Value:   "http://127.0.0.1:" + strconv.FormatUint(uint64(config.DefaultApiPort), 10),
	}
	ipFlag := &cli.StringFlag{
		Name:    "ip",
		Aliases: []string{"i"},
		Usage:   "The IP address to bind the API server to",
		Value:   "127.0.0.1",
	}
	portFlag := &cli.UintFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Usage:   "The port to bind the API server to",
		Value:   uint(swconfig.DefaultApiPort),
	}

	app.Flags = []cli.Flag{
		moduleDirFlag,
		hyperdriveUrlFlag,
		ipFlag,
		portFlag,
	}
	app.Action = func(c *cli.Context) error {
		// Get the env vars
		moduleDir := c.String(moduleDirFlag.Name)
		hdUrlString := c.String(hyperdriveUrlFlag.Name)
		hyperdriveUrl, err := url.Parse(hdUrlString)
		if err != nil {
			return fmt.Errorf("error parsing Hyperdrive URL [%s]: %w", hdUrlString, err)
		}

		// Wait group to handle the API server (separate because of error handling)
		stopWg := new(sync.WaitGroup)

		// Create the service provider
		sp, err := services.NewServiceProvider(hyperdriveUrl, moduleDir, swconfig.ModuleName, swconfig.ClientLogName, swconfig.NewStakeWiseConfig, config.ClientTimeout)
		if err != nil {
			return fmt.Errorf("error creating service provider: %w", err)
		}
		stakewiseSp, err := swcommon.NewStakeWiseServiceProvider(sp)
		if err != nil {
			return fmt.Errorf("error creating StakeWise service provider: %w", err)
		}

		// Start the task loop
		fmt.Println("Starting task loop...")
		taskLoop := swtasks.NewTaskLoop(stakewiseSp, stopWg)
		err = taskLoop.Run()
		if err != nil {
			return fmt.Errorf("error starting task loop: %w", err)
		}

		// Start the server after the task loop so it can log into NodeSet before this starts serving registration status checks
		ip := c.String(ipFlag.Name)
		port := c.Uint64(portFlag.Name)
		serverMgr, err := server.NewServerManager(stakewiseSp, ip, uint16(port), stopWg)
		if err != nil {
			return fmt.Errorf("error creating StakeWise server: %w", err)
		}

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			fmt.Println("Shutting down daemon...")
			stakewiseSp.CancelContextOnShutdown()
			serverMgr.Stop()
		}()

		// Run the daemon until closed
		fmt.Println("Daemon online.")
		fmt.Printf("HD client calls are being logged to: %s\n", sp.GetClientLogger().GetFilePath())
		fmt.Printf("API calls are being logged to: %s\n", sp.GetApiLogger().GetFilePath())
		fmt.Printf("Tasks are being logged to:     %s\n", sp.GetTasksLogger().GetFilePath())
		fmt.Println("To view them, use `hyperdrive service daemon-logs [sw-hd | sw-api | sw-tasks].") // TODO: don't hardcode
		stopWg.Wait()
		sp.Close()
		fmt.Println("Daemon stopped.")
		return nil
	}

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
