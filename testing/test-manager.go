package testing

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	hdservices "github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/nodeset-org/osha/keys"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"
)

const (
	deploymentName string = "localtest"
)

// StakeWiseTestManager for managing testing resources and services
type StakeWiseTestManager struct {
	*hdtesting.HyperdriveTestManager

	// The complete StakeWise node
	node *StakeWiseNode

	// The ID of the baseline snapshot
	baselineSnapshotID string

	mainNodeAddress common.Address
	nsEmail         string
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
	module := &StakeWiseTestManager{
		HyperdriveTestManager: tm,
		node:                  node,
	}

	err = module.SetupTest()
	if err != nil {
		return nil, fmt.Errorf("error setting up test: %w", err)
	}

	tm.RegisterModule(module)
	baselineSnapshot, err := tm.CreateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("error creating baseline snapshot: %w", err)
	}
	module.baselineSnapshotID = baselineSnapshot

	return module, nil
}

// Initialize test manager by generating a new wallet, starting the StakeWise node, and registering with NodeSet
func (m *StakeWiseTestManager) SetupTest() error {
	mainNode := m.GetNode()
	m.nsEmail = "test@nodeset.io"
	// Generate a new wallet
	derivationPath := string(wallet.DerivationPath_Default)
	index := uint64(0)
	password := "test_password123"
	hdNode := mainNode.GetHyperdriveNode()
	hd := hdNode.GetApiClient()
	recoverResponse, err := hd.Wallet.Recover(&derivationPath, keys.DefaultMnemonic, &index, password, true)
	if err != nil {
		return fmt.Errorf("error generating wallet: %v", err)
	}
	m.mainNodeAddress = recoverResponse.Data.AccountAddress

	// Set up NodeSet with the StakeWise vault
	sp := mainNode.GetServiceProvider()
	res := sp.GetResources()
	nsMgr := m.GetNodeSetMockServer().GetManager()
	nsDB := nsMgr.GetDatabase()
	deployment := nsDB.StakeWise.AddDeployment(res.DeploymentName, big.NewInt(int64(res.ChainID)))
	_ = deployment.AddVault(res.Vault)
	nsDB.SetSecretEncryptionIdentity(hdtesting.EncryptionIdentity)

	// Make a NodeSet account
	_, err = nsDB.Core.AddUser(m.nsEmail)
	if err != nil {
		return fmt.Errorf("error adding user to nodeset: %v", err)
	}

	// Register the primary
	err = m.registerWithNodeset(mainNode, m.mainNodeAddress)
	if err != nil {
		return fmt.Errorf("error registering with nodeset: %v", err)
	}
	return nil
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

func (m *StakeWiseTestManager) GetMainNodeAddress() common.Address {
	return m.mainNodeAddress
}

// ====================
// === Snapshotting ===
// ====================

// Reverts the service states to the baseline snapshot
func (m *StakeWiseTestManager) DependsOnStakewiseBaseline() error {
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
	return snapshotName, nil
}

func (m *StakeWiseTestManager) RevertModuleToSnapshot(moduleState any) error {
	err := m.HyperdriveTestManager.RevertModuleToSnapshot(moduleState)
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %w", err)
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

// Register a node with nodeset
func (m *StakeWiseTestManager) registerWithNodeset(node *StakeWiseNode, address common.Address) error {
	// whitelist the node with the nodeset.io account
	nsServer := m.GetNodeSetMockServer().GetManager()
	nsDB := nsServer.GetDatabase()
	user := nsDB.Core.GetUser(m.nsEmail)
	_ = user.WhitelistNode(address)

	// Register with NodeSet
	hd := node.GetHyperdriveNode().GetApiClient()
	response, err := hd.NodeSet.RegisterNode(m.nsEmail)
	if err != nil {
		return fmt.Errorf("error registering node with nodeset: %v", err)
	}
	if response.Data.AlreadyRegistered {
		return fmt.Errorf("node is already registered with nodeset")
	}
	if response.Data.NotWhitelisted {
		return fmt.Errorf("node is not whitelisted with a nodeset user account")
	}
	return nil
}
