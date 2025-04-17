package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/nodeset-org/hyperdrive-daemon/shared"
	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/nodeset-org/hyperdrive-stakewise/relay"
	"github.com/nodeset-org/hyperdrive-stakewise/server"
	swshared "github.com/nodeset-org/hyperdrive-stakewise/shared"
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
	app.Version = swshared.StakewiseVersion
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
		Value:   "http://127.0.0.1:" + strconv.FormatUint(uint64(hdconfig.DefaultApiPort), 10),
	}
	settingsFolderFlag := &cli.StringFlag{
		Name:     "settings-folder",
		Aliases:  []string{"s"},
		Usage:    "The path to the folder containing the network settings files",
		Required: true,
	}
	ipFlag := &cli.StringFlag{
		Name:    "ip",
		Aliases: []string{"i"},
		Usage:   "The IP address to bind the API server to",
		Value:   "127.0.0.1",
	}
	apiPortFlag := &cli.UintFlag{
		Name:    "api-port",
		Aliases: []string{"p"},
		Usage:   "The port to bind the API server to",
		Value:   uint(swconfig.DefaultApiPort),
	}
	apiKeyFlag := &cli.StringFlag{
		Name:     "api-key",
		Aliases:  []string{"k"},
		Usage:    "Path of the key to use for authenticating incoming API requests",
		Required: true,
	}
	hyperdriveApiKeyFlag := &cli.StringFlag{
		Name:     "hd-api-key",
		Aliases:  []string{"hk"},
		Usage:    "Path of the key to use when sending requests to the Hyperdrive API",
		Required: true,
	}
	relayPortFlag := &cli.UintFlag{
		Name:    "relay-port",
		Aliases: []string{"rp"},
		Usage:   "The port to bind the relay server to, for StakeWise Operator to connect to",
		Value:   uint(swconfig.DefaultRelayPort),
	}

	app.Flags = []cli.Flag{
		moduleDirFlag,
		hyperdriveUrlFlag,
		settingsFolderFlag,
		ipFlag,
		apiPortFlag,
		relayPortFlag,
		apiKeyFlag,
		hyperdriveApiKeyFlag,
	}
	app.Action = func(c *cli.Context) error {
		// Get the env vars
		moduleDir := c.String(moduleDirFlag.Name)
		hdUrlString := c.String(hyperdriveUrlFlag.Name)
		hyperdriveUrl, err := url.Parse(hdUrlString)
		if err != nil {
			return fmt.Errorf("error parsing Hyperdrive URL [%s]: %w", hdUrlString, err)
		}

		// Get the settings file path
		settingsFolder := c.String(settingsFolderFlag.Name)
		if settingsFolder == "" {
			fmt.Println("No settings folder provided.")
			os.Exit(1)
		}
		_, err = os.Stat(settingsFolder)
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("Settings folder not found at [%s].", settingsFolder)
			os.Exit(1)
		}

		// Load the network settings
		settingsList, err := swconfig.LoadSettingsFiles(settingsFolder)
		if err != nil {
			fmt.Printf("Error loading network settings: %s", err)
			os.Exit(1)
		}

		// Make an incoming API auth manager
		apiKeyPath := c.String(apiKeyFlag.Name)
		moduleAuthMgr := auth.NewAuthorizationManager(apiKeyPath, "sw-daemon-svr", auth.DefaultRequestLifespan)
		err = moduleAuthMgr.LoadAuthKey()
		if err != nil {
			return fmt.Errorf("error loading module API key: %w", err)
		}

		// Make an HD API auth manager
		hdApiKeyPath := c.String(hyperdriveApiKeyFlag.Name)
		hdAuthMgr := auth.NewAuthorizationManager(hdApiKeyPath, "sw-daemon-client", auth.DefaultRequestLifespan)
		err = hdAuthMgr.LoadAuthKey()
		if err != nil {
			return fmt.Errorf("error loading Hyperdrive API key: %w", err)
		}

		// Wait group to handle the API server (separate because of error handling)
		stopWg := new(sync.WaitGroup)

		// Create the service provider
		configFactory := func(hdCfg *hdconfig.HyperdriveConfig) (*swconfig.StakeWiseConfig, error) {
			return swconfig.NewStakeWiseConfig(hdCfg, settingsList)
		}
		sp, err := services.NewModuleServiceProvider(hyperdriveUrl, moduleDir, swconfig.ModuleName, swconfig.ClientLogName, configFactory, hdAuthMgr)
		if err != nil {
			return fmt.Errorf("error creating service provider: %w", err)
		}
		stakewiseSp, err := swcommon.NewStakeWiseServiceProvider(sp, settingsList)
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

		// Start the API server after the task loop so it can log into NodeSet before this starts serving registration status checks
		ip := c.String(ipFlag.Name)
		port := c.Uint64(apiPortFlag.Name)
		serverMgr, err := server.NewServerManager(stakewiseSp, ip, uint16(port), stopWg, moduleAuthMgr)
		if err != nil {
			return fmt.Errorf("error creating API server: %w", err)
		}

		// Start the relay server
		relayPort := c.Uint64(relayPortFlag.Name)
		relayServer, err := relay.NewRelayServer(stakewiseSp, ip, uint16(relayPort))
		if err != nil {
			return fmt.Errorf("error creating relay server: %w", err)
		}
		err = relayServer.Start(stopWg)
		if err != nil {
			return fmt.Errorf("error starting relay server: %w", err)
		}
		fmt.Printf("Relay server started on %s:%d\n", ip, relayPort)

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			fmt.Println("Shutting down daemon...")
			stakewiseSp.CancelContextOnShutdown()
			serverMgr.Stop()
			err := relayServer.Stop()
			if err != nil {
				fmt.Printf("WARNING: relay server didn't shutdown cleanly: %s\n", err.Error())
			}
		}()

		// Run the daemon until closed
		fmt.Println("Daemon online.")
		fmt.Printf("HD client calls are being logged to: %s\n", sp.GetClientLogger().GetFilePath())
		fmt.Printf("API calls are being logged to:       %s\n", sp.GetApiLogger().GetFilePath())
		fmt.Printf("Tasks are being logged to:           %s\n", sp.GetTasksLogger().GetFilePath())
		fmt.Printf("Relay calls are being logged to:     %s\n", relayServer.GetLogPath())
		fmt.Println("To view them, use `hyperdrive service daemon-logs [sw-hd | sw-api | sw-tasks | sw-relay].") // TODO: don't hardcode
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
