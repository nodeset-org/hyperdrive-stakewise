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

const (
	deploymentName string = "localtest"
)

// StakeWiseTestManager for managing testing resources and services
type StakeWiseTestManager struct {
	*hdtesting.HyperdriveTestManager

	// The complete StakeWise node
	node *StakeWiseNode

	// The StakeWise operator mock
	operatorMock *OperatorMock

	// The snapshot ID of the baseline snapshot
	baselineSnapshotID string
}

// Creates a new TestManager instance
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
	resources := getTestResources(hdSp.GetResources(), deploymentName)
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

	// Create the StakeWise node
	nodeIP := "localhost"
	node, err := newStakeWiseNode(stakeWiseSP, nodeIP, tm.GetLogger(), hdNode)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating StakeWise node: %v", err)
	}

	// Create the Operator mock
	relayUrl := fmt.Sprintf("http://%s:%d", nodeIP, node.relayServer.GetPort())
	operatorMock, err := NewOperatorMock(relayUrl, resources, 0)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error creating operator mock: %v", err)
	}

	// Disable automining
	err = tm.ToggleAutoMine(false)
	if err != nil {
		closeTestManager(tm)
		return nil, fmt.Errorf("error disabling automining: %v", err)
	}

	// Return
	module := &StakeWiseTestManager{
		HyperdriveTestManager: tm,
		node:                  node,
		operatorMock:          operatorMock,
	}
	tm.RegisterModule(module)

	baselineSnapshot, err := tm.CreateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	module.baselineSnapshotID = baselineSnapshot
	return module, nil
}

// ===============
// === Getters ===
// ===============

func (m *StakeWiseTestManager) GetModuleName() string {
	return "hyperdrive-stakewise"
}

// Get the node handle
func (m *StakeWiseTestManager) GetNode() *StakeWiseNode {
	return m.node
}

// Get the operator mock handle
func (m *StakeWiseTestManager) GetOperatorMock() *OperatorMock {
	return m.operatorMock
}

// ====================
// === Snapshotting ===
// ====================

// Reverts the service states to the baseline snapshot
func (m *StakeWiseTestManager) DependsOnStakeWiseBaseline() error {
	err := m.RevertSnapshot(m.baselineSnapshotID)
	if err != nil {
		return fmt.Errorf("error reverting to baseline snapshot: %w", err)
	}
	return nil
}

// Takes a snapshot of the service states
func (m *StakeWiseTestManager) TakeModuleSnapshot() (any, error) {
	snapshotName, err := m.HyperdriveTestManager.TakeModuleSnapshot()
	if err != nil {
		return nil, fmt.Errorf("error taking snapshot: %w", err)
	}
	// The OperatorMock doesn't have any state so it doesn't need snapshotting
	return snapshotName, nil
}

func (m *StakeWiseTestManager) RevertModuleToSnapshot(moduleState any) error {
	err := m.HyperdriveTestManager.RevertModuleToSnapshot(moduleState)
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %w", err)
	}

	// Reload the SW wallet to undo any changes made during the test
	wallet := m.node.sp.GetWallet()
	err = wallet.Reload()
	if err != nil {
		return fmt.Errorf("error reloading stakewise wallet: %v", err)
	}

	// Reload the available key manager
	err = m.node.sp.GetAvailableKeyManager().Reload()
	if err != nil {
		return fmt.Errorf("error reloading available key manager: %v", err)
	}
	return nil
}

// Closes the test manager, shutting down the nodeset mock server and all other resources
func (m *StakeWiseTestManager) CloseModule() error {
	err := m.node.Close()
	if err != nil {
		return fmt.Errorf("error closing StakeWise node: %w", err)
	}
	if m.HyperdriveTestManager != nil {
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
