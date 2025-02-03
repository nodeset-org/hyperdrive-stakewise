package swclient

import (
	"fmt"
	"log/slog"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/context"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"

	hdclient "github.com/nodeset-org/hyperdrive-daemon/client"
	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	adapterclient "github.com/nodeset-org/hyperdrive-stakewise/adapter/client"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/urfave/cli/v2"
)

const (
	SecretsDir        string = "secrets"
	DaemonKeyFilename string = "daemon.key"

	SettingsFile       string = "user-settings.yml"
	BackupSettingsFile string = "user-settings-backup.yml"
	metricsDir         string = "metrics"

	terminalLogColor      color.Attribute = color.FgHiYellow
	HyperdriveDaemonRoute string          = "hyperdrive"

	cliIssuer string = "hd-cli"
)

var hdApiKeyRelPath string = filepath.Join(SecretsDir, DaemonKeyFilename)
var moduleApiKeyRelPath string = filepath.Join(hdconfig.SecretsDir, hdconfig.ModulesName)
var swApiKeyRelPath string = filepath.Join(moduleApiKeyRelPath, swconfig.ModuleName, hdconfig.DaemonKeyFilename)

// TODO: Remove and reference from hyperdrive repo
// Hyperdrive client
type HyperdriveClient struct {
	Api      *hdclient.ApiClient
	Context  *context.HyperdriveContext
	Logger   *slog.Logger
	cfg      *GlobalConfig
	isNewCfg bool
}

// Binder for the StakeWise API server
type ApiClient struct {
	context   client.IRequesterContext
	Nodeset   *NodesetRequester
	Validator *ValidatorRequester
	Wallet    *WalletRequester
	Service   *ServiceRequester
	Status    *StatusRequester
}

// Stakewise client
type StakewiseClient struct {
	Api     *ApiClient
	Context *context.HyperdriveContext
	Logger  *slog.Logger
}

// Create new Hyperdrive client from CLI context
func NewHyperdriveClientFromCtx(c *cli.Context) (*HyperdriveClient, error) {
	hdCtx := context.GetHyperdriveContext(c)
	return NewHyperdriveClientFromHyperdriveCtx(hdCtx)
}

// Create new Hyperdrive client from a custom context
func NewHyperdriveClientFromHyperdriveCtx(hdCtx *context.HyperdriveContext) (*HyperdriveClient, error) {
	logger := log.NewTerminalLogger(hdCtx.DebugEnabled, terminalLogColor).With(slog.String(log.OriginKey, HyperdriveDaemonRoute))

	// Create the tracer if required
	var tracer *httptrace.ClientTrace
	if hdCtx.HttpTraceFile != nil {
		var err error
		tracer, err = adapterclient.CreateTracer(hdCtx.HttpTraceFile, logger)
		if err != nil {
			logger.Error("Error creating HTTP trace", log.Err(err))
		}
	}

	// Load the network settings from disk
	err := hdCtx.LoadNetworkSettings()
	if err != nil {
		return nil, fmt.Errorf("error loading network settings: %w", err)
	}

	// Make the client
	hdClient := &HyperdriveClient{
		Context: hdCtx,
		Logger:  logger,
	}

	// Get the API URL
	url := hdCtx.ApiUrl
	if url == nil {
		// Load the config to get the API port
		cfg, _, err := hdClient.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading config: %w", err)
		}

		url, err = url.Parse(fmt.Sprintf("http://localhost:%d/%s", cfg.Hyperdrive.ApiPort.Value, hdconfig.HyperdriveApiClientRoute))
		if err != nil {
			return nil, fmt.Errorf("error parsing Hyperdrive API URL: %w", err)
		}
	}

	// Create the auth manager
	authPath := filepath.Join(hdCtx.UserDirPath, hdApiKeyRelPath)
	err = auth.GenerateAuthKeyIfNotPresent(authPath, auth.DefaultKeyLength)
	if err != nil {
		return nil, fmt.Errorf("error generating Hyperdrive daemon API key: %w", err)
	}
	authMgr := auth.NewAuthorizationManager(authPath, cliIssuer, auth.DefaultRequestLifespan)

	// Create the API client
	hdClient.Api = hdclient.NewApiClient(url, logger, tracer, authMgr)
	return hdClient, nil
}

// Creates a new API client instance
func NewApiClient(apiUrl *url.URL, logger *slog.Logger, tracer *httptrace.ClientTrace, authMgr *auth.AuthorizationManager) *ApiClient {
	context := client.NewNetworkRequesterContext(apiUrl, logger, tracer, authMgr.AddAuthHeader)

	client := &ApiClient{
		context:   context,
		Nodeset:   NewNodesetRequester(context),
		Validator: NewValidatorRequester(context),
		Wallet:    NewWalletRequester(context),
		Service:   NewServiceRequester(context),
		Status:    NewStatusRequester(context),
	}
	return client
}

func (c *HyperdriveClient) LoadConfig() (*GlobalConfig, bool, error) {
	if c.cfg != nil {
		return c.cfg, c.isNewCfg, nil
	}

	settingsFilePath := filepath.Join(c.Context.UserDirPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, false, fmt.Errorf("error expanding settings file path: %w", err)
	}

	cfg, err := LoadConfigFromFile(expandedPath, c.Context.HyperdriveNetworkSettings, c.Context.StakeWiseNetworkSettings)
	if err != nil {
		return nil, false, err
	}

	if cfg != nil {
		// A config was loaded, return it now
		c.cfg = cfg
		return cfg, false, nil
	}

	// Config wasn't loaded, but there was no error - we should create one.
	hdCfg, err := hdconfig.NewHyperdriveConfig(c.Context.UserDirPath, c.Context.HyperdriveNetworkSettings)
	if err != nil {
		return nil, false, fmt.Errorf("error creating Hyperdrive config: %w", err)
	}

	c.cfg, err = NewGlobalConfig(hdCfg, c.Context.HyperdriveNetworkSettings)
	if err != nil {
		return nil, false, fmt.Errorf("error creating global config: %w", err)
	}
	c.isNewCfg = true
	return c.cfg, true, nil
}

// Load the backup config
func (c *HyperdriveClient) LoadBackupConfig() (*GlobalConfig, error) {
	settingsFilePath := filepath.Join(c.Context.UserDirPath, BackupSettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error expanding backup settings file path: %w", err)
	}

	return LoadConfigFromFile(expandedPath, c.Context.HyperdriveNetworkSettings, c.Context.StakeWiseNetworkSettings)
}

// Save the config
func (c *HyperdriveClient) SaveConfig(cfg *GlobalConfig) error {
	settingsFileDirectoryPath, err := homedir.Expand(c.Context.UserDirPath)
	if err != nil {
		return err
	}
	err = SaveConfig(cfg, settingsFileDirectoryPath, SettingsFile)
	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	// Update the client's config cache
	c.cfg = cfg
	c.isNewCfg = false
	return nil
}

// Loads a config without updating it if it exists
func LoadConfigFromFile(configPath string, hdSettings []*hdconfig.HyperdriveSettings, swSettings []*swconfig.StakeWiseSettings) (*GlobalConfig, error) {
	// Make sure the config file exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// Get the Hyperdrive config
	hdCfg, err := hdconfig.LoadFromFile(configPath, hdSettings)
	if err != nil {
		return nil, err
	}

	// Load the module configs
	cfg, err := NewGlobalConfig(hdCfg, hdSettings)
	if err != nil {
		return nil, fmt.Errorf("error creating global configuration: %w", err)
	}
	err = cfg.DeserializeModules()
	if err != nil {
		return nil, fmt.Errorf("error loading module configs from [%s]: %w", configPath, err)
	}

	return cfg, nil
}

// Create new Stakewise client from CLI context
// Only use this function from commands that may work if the Daemon service doesn't exist
func NewStakewiseClientFromCtx(c *cli.Context, hdClient *HyperdriveClient) (*StakewiseClient, error) {
	hdCtx := context.GetHyperdriveContext(c)
	return NewStakewiseClientFromHyperdriveCtx(hdCtx, hdClient)
}

// Create new Stakewise client from a custom context
// Only use this function from commands that may work if the Daemon service doesn't exist
func NewStakewiseClientFromHyperdriveCtx(hdCtx *context.HyperdriveContext, hdClient *HyperdriveClient) (*StakewiseClient, error) {
	return &StakewiseClient{}, nil
	// logger := log.NewTerminalLogger(hdCtx.DebugEnabled, terminalLogColor).With(slog.String(log.OriginKey, swconfig.ModuleName))

	// // Create the tracer if required
	// var tracer *httptrace.ClientTrace
	// if hdCtx.HttpTraceFile != nil {
	// 	var err error
	// 	tracer, err = adapterclient.CreateTracer(hdCtx.HttpTraceFile, logger)
	// 	if err != nil {
	// 		logger.Error("Error creating HTTP trace", log.Err(err))
	// 	}
	// }

	// // Make the client
	// swClient := &StakewiseClient{
	// 	Context: hdCtx,
	// 	Logger:  logger,
	// }

	// // Get the API URL
	// url := hdCtx.ApiUrl
	// if url == nil {
	// 	var err error
	// 	url, err = url.Parse(fmt.Sprintf("http://localhost:%d/%s", hdClient.cfg.StakeWise.ApiPort.Value, swconfig.ApiClientRoute))
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error parsing StakeWise API URL: %w", err)
	// 	}
	// } else {
	// 	host := fmt.Sprintf("%s://%s:%d/%s", url.Scheme, url.Hostname(), hdClient.cfg.StakeWise.ApiPort.Value, swconfig.ApiClientRoute)
	// 	var err error
	// 	url, err = url.Parse(host)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error parsing StakeWise API URL: %w", err)
	// 	}
	// }

	// // Create the auth manager
	// authPath := filepath.Join(hdCtx.UserDirPath, swApiKeyRelPath)
	// err := auth.GenerateAuthKeyIfNotPresent(authPath, auth.DefaultKeyLength)
	// if err != nil {
	// 	return nil, fmt.Errorf("error generating StakeWise module API key: %w", err)
	// }
	// authMgr := auth.NewAuthorizationManager(authPath, cliIssuer, auth.DefaultRequestLifespan)

	// // Create the API client
	// swClient.Api = NewApiClient(url, logger, tracer, authMgr)
	// return swClient, nil
}
