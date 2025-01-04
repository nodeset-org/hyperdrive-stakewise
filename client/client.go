package swclient

import (
	"fmt"
	"log/slog"
	"net/http/httptrace"
	"net/url"
	"path/filepath"

	docker "github.com/docker/docker/client"
	"github.com/fatih/color"

	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	hdclient "github.com/nodeset-org/hyperdrive-stakewise/adapter/client"
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config"
	"github.com/nodeset-org/hyperdrive-stakewise/client/utils"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/urfave/cli/v2"
)

const terminalLogColor color.Attribute = color.FgHiYellow

var hdApiKeyRelPath string = filepath.Join(config.SecretsDir, config.DaemonKeyFilename)

// Binder for the StakeWise API server
type ApiClient struct {
	context   client.IRequesterContext
	Nodeset   *NodesetRequester
	Validator *ValidatorRequester
	Wallet    *WalletRequester
	Service   *ServiceRequester
	Status    *StatusRequester
}

// Hyperdrive client
type HyperdriveClient struct {
	Api      *ApiClient
	Context  *utils.HyperdriveContext
	Logger   *slog.Logger
	docker   *docker.Client
	cfg      *GlobalConfig
	isNewCfg bool
}

// Create new Hyperdrive client from CLI context
func NewHyperdriveClientFromCtx(c *cli.Context) (*HyperdriveClient, error) {
	hdCtx := utils.GetHyperdriveContext(c)
	return NewHyperdriveClientFromHyperdriveCtx(hdCtx)
}

// Create new Hyperdrive client from a custom context
func NewHyperdriveClientFromHyperdriveCtx(hdCtx *utils.HyperdriveContext) (*HyperdriveClient, error) {
	logger := log.NewTerminalLogger(hdCtx.DebugEnabled, terminalLogColor).With(slog.String(log.OriginKey, config.HyperdriveDaemonRoute))

	// Create the tracer if required
	var tracer *httptrace.ClientTrace
	if hdCtx.HttpTraceFile != nil {
		var err error
		tracer, err = hdclient.CreateTracer(hdCtx.HttpTraceFile, logger)
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
	hdClient.Api = client.NewApiClient(url, logger, tracer, authMgr)
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
