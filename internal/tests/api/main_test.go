package api_test

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	swtesting "github.com/nodeset-org/hyperdrive-stakewise/testing"
	"github.com/nodeset-org/osha/keys"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Various singleton variables used for testing
var (
	testMgr             *swtesting.StakeWiseTestManager = nil
	unregisteredTestMgr *swtesting.StakeWiseTestManager = nil
	wg                  *sync.WaitGroup                 = nil
	logger              *slog.Logger                    = nil
	nodeAddress         common.Address
	nsEmail             string = "test@nodeset.io"
)
var (
	unregisteredTestMgrIndex           = uint64(999999)
	unregisteredTestMgrDefaultMnemonic = "radar blur cabbage chef fix engine embark joy scheme fiction master release"
)

func TestMain(m *testing.M) {
	wg = &sync.WaitGroup{}

	// Initialize test managers
	unregisteredTestMgr, err := initializeTestManager()
	if err != nil {
		fail("error creating unregistered test manager: %v", err)
	}

	testMgr, err := initializeTestManager()
	if err != nil {
		fail("error creating test manager: %v", err)
	}

	logger = testMgr.GetLogger()

	// Generate wallets
	password := "test_password123"
	derivationPath := string(wallet.DerivationPath_Default)
	unregisteredRecoverResponse, err := recoverWallet(unregisteredTestMgr, unregisteredTestMgrDefaultMnemonic, &unregisteredTestMgrIndex, password, derivationPath)
	if err != nil {
		fail("error recovering unregistered wallet: %v", err)
	}

	index := uint64(0)
	recoverResponse, err := recoverWallet(testMgr, keys.DefaultMnemonic, &index, password, derivationPath)
	if err != nil {
		fail("error generating wallet: %v", err)
	}
	nodeAddress = recoverResponse.Data.AccountAddress

	// Setup NodeSet and add StakeWise vaults
	err = setupNodeSet(testMgr, nsEmail, nodeAddress, recoverResponse.Data.AccountAddress)
	if err != nil {
		fail("error setting up NodeSet: %v", err)
	}

	err = setupNodeSet(unregisteredTestMgr, nsEmail, unregisteredRecoverResponse.Data.AccountAddress, recoverResponse.Data.AccountAddress)
	if err != nil {
		fail("error setting up NodeSet for unregistered manager: %v", err)
	}

	// Register with NodeSet
	if err = registerNode(testMgr, nsEmail, nodeAddress); err != nil {
		fail("error registering node with nodeset: %v", err)
	}

	// Run tests
	code := m.Run()

	// Clean up and exit
	cleanup()
	os.Exit(code)
}

func initializeTestManager() (*swtesting.StakeWiseTestManager, error) {
	return swtesting.NewStakeWiseTestManager("localhost", "localhost", "localhost")
}

func recoverWallet(testMgr *swtesting.StakeWiseTestManager, mnemonic string, index *uint64, password, derivationPath string) (*wallet.RecoverResponse, error) {
	hdClient := testMgr.HyperdriveTestManager.GetApiClient()
	return hdClient.Wallet.Recover(&derivationPath, mnemonic, index, password, true)
}

func setupNodeSet(testMgr *swtesting.StakeWiseTestManager, email, accountAddress1, accountAddress2 string) error {
	sp := testMgr.GetStakeWiseServiceProvider()
	res := sp.GetResources()
	nsServer := testMgr.GetNodeSetMockServer().GetManager()

	if err := nsServer.AddStakeWiseVault(*res.Vault, res.EthNetworkName); err != nil {
		return fmt.Errorf("error adding stakewise vault to nodeset: %w", err)
	}

	if err := nsServer.AddUser(email); err != nil {
		return fmt.Errorf("error adding user to nodeset: %w", err)
	}

	if err := nsServer.WhitelistNodeAccount(email, accountAddress1); err != nil {
		return fmt.Errorf("error whitelisting node account: %w", err)
	}

	if err := nsServer.WhitelistNodeAccount(email, accountAddress2); err != nil {
		return fmt.Errorf("error whitelisting node account: %w", err)
	}

	return nil
}

func registerNode(testMgr *swtesting.StakeWiseTestManager, email, nodeAddress string) error {
	logger := log.NewDefaultLogger()
	ctx := logger.CreateContextWithLogger(testMgr.GetStakeWiseServiceProvider().GetBaseContext())
	nsClient := testMgr.GetStakeWiseServiceProvider().GetNodesetClient()
	return nsClient.RegisterNode(ctx, email, nodeAddress)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	cleanup()
	os.Exit(1)
}

func cleanup() {
	if testMgr == nil {
		return
	}
	err := testMgr.Close()
	if err != nil {
		logger.Error("Error closing test manager", log.Err(err))
	}
	testMgr = nil
}
