package testing

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	hdservices "github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/nodeset-org/hyperdrive-stakewise/relay"
	swserver "github.com/nodeset-org/hyperdrive-stakewise/server"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

const (
	apiAuthKey string = "sw-test-key"
)

// A complete StakeWise node instance
type StakeWiseNode struct {
	// The daemon's service provider
	sp swcommon.IStakeWiseServiceProvider

	// The daemon's HTTP API server
	serverMgr *swserver.ServerManager

	// The daemon's HTTP Relay server
	relayServer *relay.RelayServer

	// An HTTP API client for the daemon
	client *swclient.ApiClient

	// The client logger
	logger *slog.Logger

	// The Hyperdrive node parent
	hdNode *hdtesting.HyperdriveNode

	// Wait groups for graceful shutdown
	apiWg   *sync.WaitGroup
	relayWg *sync.WaitGroup
}

// Create a new StakeWise node, including its folder structure, service provider, server manager, and API client.
func newStakeWiseNode(sp swcommon.IStakeWiseServiceProvider, address string, clientLogger *slog.Logger, hyperdriveNode *hdtesting.HyperdriveNode) (*StakeWiseNode, error) {
	// Create the server
	apiWg := &sync.WaitGroup{}
	cfg := sp.GetConfig()
	serverAuthMgr := auth.NewAuthorizationManager("", "sw-server", auth.DefaultRequestLifespan)
	serverAuthMgr.SetKey([]byte(apiAuthKey))
	serverMgr, err := swserver.NewServerManager(sp, address, cfg.ApiPort.Value, apiWg, serverAuthMgr)
	if err != nil {
		return nil, fmt.Errorf("error creating constellation server: %v", err)
	}

	// Create the relay
	relayWg := &sync.WaitGroup{}
	relayServer, err := relay.NewRelayServer(sp, address, cfg.RelayPort.Value)
	if err != nil {
		return nil, fmt.Errorf("error creating relay server: %v", err)
	}
	err = relayServer.Start(relayWg)
	if err != nil {
		return nil, fmt.Errorf("error starting relay server: %v", err)
	}

	// Create the client
	urlString := fmt.Sprintf("http://%s:%d/%s", address, serverMgr.GetPort(), swconfig.ApiClientRoute)
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("error parsing client URL [%s]: %v", urlString, err)
	}
	clientAuthMgr := auth.NewAuthorizationManager("", "sw-client", auth.DefaultRequestLifespan)
	clientAuthMgr.SetKey([]byte(apiAuthKey))
	apiClient := swclient.NewApiClient(url, clientLogger, nil, clientAuthMgr)

	return &StakeWiseNode{
		sp:          sp,
		serverMgr:   serverMgr,
		relayServer: relayServer,
		client:      apiClient,
		logger:      clientLogger,
		hdNode:      hyperdriveNode,
		apiWg:       apiWg,
		relayWg:     relayWg,
	}, nil
}

// Closes the StakeWise node. The caller is responsible for stopping the Hyperdrive daemon owning this module.
func (n *StakeWiseNode) Close() error {
	if n.serverMgr != nil {
		n.serverMgr.Stop()
		n.apiWg.Wait()
		n.serverMgr = nil
		n.logger.Info("Stopped StakeWise daemon API server")
	}
	if n.relayServer != nil {
		err := n.relayServer.Stop()
		if err != nil {
			n.logger.Warn("relay server didn't shutdown cleanly", "error", err.Error())
		}
		n.relayWg.Wait()
		n.relayServer = nil
		n.logger.Info("Stopped StakeWise relay server")
	}
	return n.hdNode.Close()
}

// Get the daemon's service provider
func (n *StakeWiseNode) GetServiceProvider() swcommon.IStakeWiseServiceProvider {
	return n.sp
}

// Get the HTTP API server for the node's daemon
func (n *StakeWiseNode) GetServerManager() *swserver.ServerManager {
	return n.serverMgr
}

// Get the HTTP Relay server for the node's daemon
func (n *StakeWiseNode) GetRelayServer() *relay.RelayServer {
	return n.relayServer
}

// Get the HTTP API client for interacting with the node's daemon server
func (n *StakeWiseNode) GetApiClient() *swclient.ApiClient {
	return n.client
}

// Get the Hyperdrive node for this StakeWise module
func (n *StakeWiseNode) GetHyperdriveNode() *hdtesting.HyperdriveNode {
	return n.hdNode
}

// Create a new StakeWise node based on this one's configuration, but with a custom folder, address, and port.
func (n *StakeWiseNode) CreateSubNode(hdNode *hdtesting.HyperdriveNode, address string, port uint16) (*StakeWiseNode, error) {
	// Get the HD artifacts
	hdSp := hdNode.GetServiceProvider()
	hdCfg := hdSp.GetConfig()
	hdClient := hdNode.GetApiClient()

	// Make Constellation resources
	resources := getTestResources(hdSp.GetResources(), deploymentName)
	csCfg, err := swconfig.NewStakeWiseConfig(hdCfg, []*swconfig.StakeWiseSettings{})
	if err != nil {
		return nil, fmt.Errorf("error creating Constellation config: %v", err)
	}
	csCfg.ApiPort.Value = port

	// Make sure the module directory exists
	moduleDir := filepath.Join(hdCfg.UserDataPath.Value, hdconfig.ModulesName, swconfig.ModuleName)
	err = os.MkdirAll(moduleDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating data and modules directories [%s]: %v", moduleDir, err)
	}

	// Make a new service provider
	moduleSp, err := hdservices.NewModuleServiceProviderFromArtifacts(
		hdClient,
		hdCfg,
		csCfg,
		hdSp.GetResources(),
		moduleDir,
		swconfig.ModuleName,
		swconfig.ClientLogName,
		hdSp.GetEthClient(),
		hdSp.GetBeaconClient(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating service provider: %v", err)
	}
	csSp, err := swcommon.NewStakeWiseServiceProviderFromCustomServices(moduleSp, csCfg, resources)
	if err != nil {
		return nil, fmt.Errorf("error creating stakewise service provider: %v", err)
	}
	return newStakeWiseNode(csSp, address, n.logger, hdNode)
}
