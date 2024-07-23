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
	"github.com/rocket-pool/node-manager-core/log"
)

// StakeWiseTestManager for managing testing resources and services
type StakeWiseTestManager struct {
	*hdtesting.HyperdriveTestManager

	// The service provider for the test environment
	sp swcommon.IStakeWiseServiceProvider

	// The StakeWise Daemon server
	serverMgr *swserver.ServerManager

	// The StakeWise Daemon client
	apiClient *swclient.ApiClient

	// Wait group for graceful shutdown
	swWg *sync.WaitGroup
}

// Creates a new TestManager instance
// `hdAddress` is the address to bind the Hyperdrive daemon to.
// `swAddress` is the address to bind the StakeWise daemon to.
// `nsAddress` is the address to bind the nodeset.io mock server to.
func NewStakeWiseTestManager(hdAddress string, swAddress string, nsAddress string) (*StakeWiseTestManager, error) {
	tm, err := hdtesting.NewHyperdriveTestManagerWithDefaults(hdAddress, nsAddress, provisionNetworkSettings)
	if err != nil {
		return nil, fmt.Errorf("error creating test manager: %w", err)
	}

	// Get the HD artifacts
	hdSp := tm.GetServiceProvider()
	hdCfg := hdSp.GetConfig()
	hdClient := tm.GetApiClient()

	// Make StakeWise resources
	resources := GetTestResources(hdSp.GetResources())
	swCfg, err := swconfig.NewStakeWiseConfig(hdCfg, []*swconfig.StakeWiseSettings{})
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating StakeWise config: %v", err)
	}

	// Make the module directory
	moduleDir := filepath.Join(hdCfg.UserDataPath.Value, hdconfig.ModulesName, swconfig.ModuleName)
	err = os.MkdirAll(moduleDir, 0755)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating module directory [%s]: %v", moduleDir, err)
	}

	// Make a new service provider
	moduleSp, err := hdservices.NewModuleServiceProviderFromArtifacts(hdClient, hdCfg, swCfg, hdSp.GetResources(), moduleDir, swconfig.ModuleName, swconfig.ClientLogName, hdSp.GetEthClient(), hdSp.GetBeaconClient())
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
	serverMgr, err := swserver.NewServerManager(stakeWiseSP, swAddress, 0, swWg)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating stakewise server: %v", err)
	}

	// Create the client
	urlString := fmt.Sprintf("http://%s:%d/%s", swAddress, serverMgr.GetPort(), swconfig.ApiClientRoute)
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
		serverMgr:             serverMgr,
		apiClient:             apiClient,
		swWg:                  swWg,
	}
	return m, nil
}

// Get the StakeWise service provider
func (m *StakeWiseTestManager) GetStakeWiseServiceProvider() swcommon.IStakeWiseServiceProvider {
	return m.sp
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
