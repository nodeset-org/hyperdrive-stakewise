package testing

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	hdservices "github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swclient "github.com/nodeset-org/hyperdrive-stakewise/client"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swserver "github.com/nodeset-org/hyperdrive-stakewise/server"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

// A complete StakeWise node instance
type StakeWiseNode struct {
	// The daemon's service provider
	sp swcommon.IStakeWiseServiceProvider

	// The daemon's HTTP API server
	serverMgr *swserver.ServerManager

	// An HTTP API client for the daemon
	client *swclient.ApiClient

	// The client logger
	logger *slog.Logger

	// The Hyperdrive node parent
	hdNode *hdtesting.HyperdriveNode

	// Wait group for graceful shutdown
	wg *sync.WaitGroup
}

// Create a new StakeWise node, including its folder structure, service provider, server manager, and API client.
func newStakeWiseNode(sp swcommon.IStakeWiseServiceProvider, address string, clientLogger *slog.Logger, hyperdriveNode *hdtesting.HyperdriveNode) (*StakeWiseNode, error) {
	// Create the server
	wg := &sync.WaitGroup{}
	cfg := sp.GetConfig()
	serverMgr, err := swserver.NewServerManager(sp, address, cfg.ApiPort.Value, wg)
	if err != nil {
		return nil, fmt.Errorf("error creating constellation server: %v", err)
	}

	// Create the client
	urlString := fmt.Sprintf("http://%s:%d/%s", address, serverMgr.GetPort(), swconfig.ApiClientRoute)
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("error parsing client URL [%s]: %v", urlString, err)
	}
	apiClient := swclient.NewApiClient(url, clientLogger, nil)

	return &StakeWiseNode{
		sp:        sp,
		serverMgr: serverMgr,
		client:    apiClient,
		logger:    clientLogger,
		hdNode:    hyperdriveNode,
		wg:        wg,
	}, nil
}

// Closes the StakeWise node. The caller is responsible for stopping the Hyperdrive daemon owning this module.
func (n *StakeWiseNode) Close() error {
	if n.serverMgr != nil {
		n.serverMgr.Stop()
		n.wg.Wait()
		n.serverMgr = nil
		n.logger.Info("Stopped StakeWise daemon API server")
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
