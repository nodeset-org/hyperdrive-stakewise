package testing

import (
	"fmt"
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
	nsserver "github.com/nodeset-org/nodeset-svc-mock/server"
	"github.com/rocket-pool/node-manager-core/log"
)

// StakeWiseTestManager for managing testing resources and services
type StakeWiseTestManager struct {
	*hdtesting.HyperdriveTestManager

	// The service provider for the test environment
	sp *swcommon.StakeWiseServiceProvider

	// The mock for the nodeset.io service
	nodesetMock *nsserver.NodeSetMockServer

	// The StakeWise Daemon server
	serverMgr *swserver.ServerManager

	// The StakeWise Daemon client
	apiClient *swclient.ApiClient

	// Wait group for graceful shutdown
	swWg *sync.WaitGroup
	nsWg *sync.WaitGroup
}

// Creates a new TestManager instance
// `hdaddress` is the address to bind the Hyperdrive daemon to.
// `swaddress` is the address to bind the StakeWise daemon to.
// `nsddress` is the address to bind the nodeset.io mock server to.
func NewStakeWiseTestManager(hdaddress string, swaddress string, nsaddress string) (*StakeWiseTestManager, error) {
	tm, err := hdtesting.NewHyperdriveTestManagerWithDefaults(hdaddress)
	if err != nil {
		return nil, fmt.Errorf("error creating test manager: %w", err)
	}

	// Get the HD artifacts
	hdSp := tm.GetServiceProvider()
	hdCfg := hdSp.GetConfig()
	hdClient := tm.GetApiClient()

	// Make the nodeset.io mock server
	nsWg := &sync.WaitGroup{}
	nodesetMock, err := nsserver.NewNodeSetMockServer(tm.GetLogger(), nsaddress, 0)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating nodeset mock server: %v", err)
	}
	err = nodesetMock.Start(nsWg)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error starting nodeset mock server: %v", err)
	}

	// Make StakeWise resources
	resources := GetTestResources(hdCfg.GetNetworkResources(), fmt.Sprintf("http://%s:%d/api/", nsaddress, nodesetMock.GetPort()))
	swCfg := swconfig.NewStakeWiseConfigWithResources(hdCfg, resources)

	// Make the module directory
	moduleDir := filepath.Join(hdCfg.UserDataPath.Value, hdconfig.ModulesName, swconfig.ModuleName)
	err = os.MkdirAll(moduleDir, 0755)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating module directory [%s]: %v", moduleDir, err)
	}

	// Make a new service provider
	moduleSp, err := hdservices.NewServiceProviderFromArtifacts(hdClient, hdCfg, swCfg, resources.NetworkResources, moduleDir, swconfig.ModuleName, swconfig.ClientLogName, hdSp.GetEthClient(), hdSp.GetBeaconClient())
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating service provider: %v", err)
	}
	stakeWiseSP, err := swcommon.NewStakeWiseServiceProviderFromCustomServices(moduleSp, swCfg, resources)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating StakeWise service provider: %v", err)
	}

	// Create the server
	swWg := &sync.WaitGroup{}
	serverMgr, err := swserver.NewServerManager(stakeWiseSP, swaddress, 0, swWg)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating stakewise server: %v", err)
	}

	// Create the client
	urlString := fmt.Sprintf("http://%s:%d/%s", swaddress, serverMgr.GetPort(), swconfig.ApiClientRoute)
	url, err := url.Parse(urlString)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error parsing client URL [%s]: %v", urlString, err)
	}
	apiClient := swclient.NewApiClient(url, tm.GetLogger(), nil)

	// Return
	m := &StakeWiseTestManager{
		HyperdriveTestManager: tm,
		sp:                    stakeWiseSP,
		nodesetMock:           nodesetMock,
		serverMgr:             serverMgr,
		apiClient:             apiClient,
		swWg:                  swWg,
		nsWg:                  nsWg,
	}
	return m, nil
}

// Get the StakeWise service provider
func (m *StakeWiseTestManager) GetStakeWiseServiceProvider() *swcommon.StakeWiseServiceProvider {
	return m.sp
}

// Get the nodeset.io mock server
func (m *StakeWiseTestManager) GetNodeSetMockServer() *nsserver.NodeSetMockServer {
	return m.nodesetMock
}

// Get the StakeWise Daemon server manager
func (m *StakeWiseTestManager) GetServerManager() *swserver.ServerManager {
	return m.serverMgr
}

// Get the StakeWise Daemon client
func (m *StakeWiseTestManager) GetApiClient() *swclient.ApiClient {
	return m.apiClient
}

// Closes the test manager, shutting down the nodeset mock server and all other resources
func (m *StakeWiseTestManager) Close() error {
	if m.nodesetMock != nil {
		err := m.nodesetMock.Stop()
		if err != nil {
			m.GetLogger().Warn("WARNING: API server didn't shutdown cleanly", log.Err(err))
		}
		m.nsWg.Wait()
		m.TestManager.GetLogger().Info("Stopped nodeset.io mock server")
		m.nodesetMock = nil
	}
	if m.serverMgr != nil {
		m.serverMgr.Stop()
		m.swWg.Wait()
		m.TestManager.GetLogger().Info("Stopped daemon API server")
		m.serverMgr = nil
	}
	if m.HyperdriveTestManager != nil {
		err := m.HyperdriveTestManager.Close()
		if err != nil {
			return fmt.Errorf("error closing test manager: %w", err)
		}
		m.HyperdriveTestManager = nil
	}
	return nil
}

// ==========================
// === Internal Functions ===
// ==========================

// Closes the Hyperdrive test manager, logging any errors
func closeTestManager(tm *hdtesting.HyperdriveTestManager) {
	err := tm.Close()
	if err != nil {
		tm.GetLogger().Error("Error closing test manager", log.Err(err))
	}
}
