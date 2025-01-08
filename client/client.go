package swclient

import (
	"fmt"
	"log/slog"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"

	clitemplate "github.com/nodeset-org/hyperdrive-stakewise/adapter/client/template"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils/config"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"

	hdclient "github.com/nodeset-org/hyperdrive-daemon/client"
	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swclient "github.com/nodeset-org/hyperdrive-stakewise/adapter/client"
	"github.com/nodeset-org/hyperdrive-stakewise/client/utils"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/urfave/cli/v2"
)

const (
	metricsDirMode           os.FileMode = 0755
	prometheusConfigTemplate string      = "prometheus-cfg.tmpl"
	prometheusConfigTarget   string      = "prometheus.yml"
	grafanaConfigTemplate    string      = "grafana-prometheus-datasource.tmpl"
	grafanaConfigTarget      string      = "grafana-prometheus-datasource.yml"

	SettingsFile       string = "user-settings.yml"
	BackupSettingsFile string = "user-settings-backup.yml"
	metricsDir         string = "metrics"

	terminalLogColor color.Attribute = color.FgHiYellow

	cliIssuer string = "hd-cli"
)

var hdApiKeyRelPath string = filepath.Join(config.SecretsDir, config.DaemonKeyFilename)
var moduleApiKeyRelPath string = filepath.Join(hdconfig.SecretsDir, hdconfig.ModulesName)
var swApiKeyRelPath string = filepath.Join(moduleApiKeyRelPath, swconfig.ModuleName, hdconfig.DaemonKeyFilename)

// TODO: Remove and reference from hyperdrive repo
// Hyperdrive client
type HyperdriveClient struct {
	Api     *hdclient.ApiClient
	Context *utils.HyperdriveContext
	Logger  *slog.Logger
	// docker   *docker.Client
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
	Context *utils.HyperdriveContext
	Logger  *slog.Logger
}

// Create new Hyperdrive client from CLI context
func NewHyperdriveClientFromCtx(c *cli.Context) (*HyperdriveClient, error) {
	hdCtx := utils.GetHyperdriveContext(c)
	return NewHyperdriveClientFromHyperdriveCtx(hdCtx)
}

// Create new Hyperdrive client from a custom context
func NewHyperdriveClientFromHyperdriveCtx(hdCtx *utils.HyperdriveContext) (*HyperdriveClient, error) {
	logger := log.NewTerminalLogger(hdCtx.DebugEnabled, terminalLogColor).With(slog.String(log.OriginKey, hdconfig.HyperdriveDaemonRoute))

	// Create the tracer if required
	var tracer *httptrace.ClientTrace
	if hdCtx.HttpTraceFile != nil {
		var err error
		tracer, err = swclient.CreateTracer(hdCtx.HttpTraceFile, logger)
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
	swCfg, err := swconfig.NewStakeWiseConfig(hdCfg, c.Context.StakeWiseNetworkSettings)
	if err != nil {
		return nil, false, fmt.Errorf("error creating StakeWise config: %w", err)
	}
	c.cfg, err = NewGlobalConfig(hdCfg, c.Context.HyperdriveNetworkSettings, swCfg, c.Context.StakeWiseNetworkSettings)
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

// Create the metrics and modules folders, and deploy the config templates for Prometheus and Grafana
func (c *HyperdriveClient) DeployMetricsConfigurations(config *GlobalConfig) error {
	// Make sure the metrics path exists
	metricsDirPath := filepath.Join(c.Context.UserDirPath, metricsDir)
	modulesDirPath := filepath.Join(metricsDirPath, hdconfig.ModulesName)
	err := os.MkdirAll(modulesDirPath, metricsDirMode)
	if err != nil {
		return fmt.Errorf("error creating metrics and modules directories [%s]: %w", modulesDirPath, err)
	}

	err = updatePrometheusConfiguration(c.Context, config, metricsDirPath)
	if err != nil {
		return fmt.Errorf("error updating Prometheus configuration: %w", err)
	}
	err = updateGrafanaDatabaseConfiguration(c.Context, config, metricsDirPath)
	if err != nil {
		return fmt.Errorf("error updating Grafana configuration: %w", err)
	}
	return nil
}

// Load the Prometheus config template, do a template variable substitution, and save it
func updatePrometheusConfiguration(ctx *utils.HyperdriveContext, config *GlobalConfig, metricsDirPath string) error {
	prometheusConfigTemplatePath, err := homedir.Expand(filepath.Join(ctx.TemplatesDir, prometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("error expanding Prometheus config template path: %w", err)
	}

	prometheusConfigTargetPath, err := homedir.Expand(filepath.Join(metricsDirPath, prometheusConfigTarget))
	if err != nil {
		return fmt.Errorf("error expanding Prometheus config target path: %w", err)
	}

	t := clitemplate.Template{
		Src: prometheusConfigTemplatePath,
		Dst: prometheusConfigTargetPath,
	}

	return t.Write(config)
}

// Load the Grafana config template, do a template variable substitution, and save it
func updateGrafanaDatabaseConfiguration(ctx *utils.HyperdriveContext, config *GlobalConfig, metricsDirPath string) error {
	grafanaConfigTemplatePath, err := homedir.Expand(filepath.Join(ctx.TemplatesDir, grafanaConfigTemplate))
	if err != nil {
		return fmt.Errorf("error expanding Grafana config template path: %w", err)
	}

	grafanaConfigTargetPath, err := homedir.Expand(filepath.Join(metricsDirPath, grafanaConfigTarget))
	if err != nil {
		return fmt.Errorf("error expanding Grafana config target path: %w", err)
	}

	t := clitemplate.Template{
		Src: grafanaConfigTemplatePath,
		Dst: grafanaConfigTargetPath,
	}

	return t.Write(config)
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

	// Get the StakeWise config
	swCfg, err := swconfig.NewStakeWiseConfig(hdCfg, swSettings)
	if err != nil {
		return nil, err
	}

	// Load the module configs
	cfg, err := NewGlobalConfig(hdCfg, hdSettings, swCfg, swSettings)
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
	hdCtx := utils.GetHyperdriveContext(c)
	return NewStakewiseClientFromHyperdriveCtx(hdCtx, hdClient)
}

// Create new Stakewise client from a custom context
// Only use this function from commands that may work if the Daemon service doesn't exist
func NewStakewiseClientFromHyperdriveCtx(hdCtx *utils.HyperdriveContext, hdClient *HyperdriveClient) (*StakewiseClient, error) {
	logger := log.NewTerminalLogger(hdCtx.DebugEnabled, terminalLogColor).With(slog.String(log.OriginKey, swconfig.ModuleName))

	// Create the tracer if required
	var tracer *httptrace.ClientTrace
	if hdCtx.HttpTraceFile != nil {
		var err error
		tracer, err = swclient.CreateTracer(hdCtx.HttpTraceFile, logger)
		if err != nil {
			logger.Error("Error creating HTTP trace", log.Err(err))
		}
	}

	// Make the client
	swClient := &StakewiseClient{
		Context: hdCtx,
		Logger:  logger,
	}

	// Get the API URL
	url := hdCtx.ApiUrl
	if url == nil {
		var err error
		url, err = url.Parse(fmt.Sprintf("http://localhost:%d/%s", hdClient.cfg.StakeWise.ApiPort.Value, swconfig.ApiClientRoute))
		if err != nil {
			return nil, fmt.Errorf("error parsing StakeWise API URL: %w", err)
		}
	} else {
		host := fmt.Sprintf("%s://%s:%d/%s", url.Scheme, url.Hostname(), hdClient.cfg.StakeWise.ApiPort.Value, swconfig.ApiClientRoute)
		var err error
		url, err = url.Parse(host)
		if err != nil {
			return nil, fmt.Errorf("error parsing StakeWise API URL: %w", err)
		}
	}

	// Create the auth manager
	authPath := filepath.Join(hdCtx.UserDirPath, swApiKeyRelPath)
	err := auth.GenerateAuthKeyIfNotPresent(authPath, auth.DefaultKeyLength)
	if err != nil {
		return nil, fmt.Errorf("error generating StakeWise module API key: %w", err)
	}
	authMgr := auth.NewAuthorizationManager(authPath, cliIssuer, auth.DefaultRequestLifespan)

	// Create the API client
	swClient.Api = NewApiClient(url, logger, tracer, authMgr)
	return swClient, nil
}
