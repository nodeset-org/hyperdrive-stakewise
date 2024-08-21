package testing

import (
	"fmt"
	"os"
	"path/filepath"

	hdservices "github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/log"
)

// StakeWiseTestManager for managing testing resources and services
type StakeWiseTestManager struct {
	*hdtesting.HyperdriveTestManager

	// The complete StakeWise node
	node *StakeWiseNode
}

// Creates a new TestManager instance
// `hdAddress` is the address to bind the Hyperdrive daemon to.
// `swAddress` is the address to bind the StakeWise daemon to.
// `nsAddress` is the address to bind the nodeset.io mock server to.
func NewStakeWiseTestManager() (*StakeWiseTestManager, error) {
	tm, err := hdtesting.NewHyperdriveTestManagerWithDefaults(provisionNetworkSettings)
	if err != nil {
		return nil, fmt.Errorf("error creating test manager: %w", err)
	}

	// Get the HD artifacts
	hdNode := tm.GetNode()
	hdSp := hdNode.GetServiceProvider()
	hdCfg := hdSp.GetConfig()
	hdClient := hdNode.GetApiClient()

	// Make StakeWise resources
	resources := getTestResources(hdSp.GetResources())
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

	// Create the Constellation node
	node, err := newStakeWiseNode(stakeWiseSP, "localhost", tm.GetLogger(), hdNode)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating Constellation node: %v", err)
	}

	// Disable automining
	err = tm.ToggleAutoMine(false)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error disabling automining: %v", err)
	}

	// Return
	m := &StakeWiseTestManager{
		HyperdriveTestManager: tm,
		node:                  node,
	}
	return m, nil
}

// Get the node handle
func (m *StakeWiseTestManager) GetNode() *StakeWiseNode {
	return m.node
}

// Closes the test manager, shutting down the nodeset mock server and all other resources
func (m *StakeWiseTestManager) Close() error {
	err := m.node.Close()
	if err != nil {
		return fmt.Errorf("error closing StakeWise node: %w", err)
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
