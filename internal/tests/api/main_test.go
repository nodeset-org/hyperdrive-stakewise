package api_test

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/hyperdrive-daemon/shared/types/api"
	swtesting "github.com/nodeset-org/hyperdrive-stakewise/testing"
	"github.com/nodeset-org/osha/keys"
	"github.com/rocket-pool/node-manager-core/api/types"
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
var ()

func TestMain(m *testing.M) {
	wg = &sync.WaitGroup{}
	var err error

	// Create the new test managers
	unregisteredTestMgr, err = initializeTestManager()
	if err != nil {
		fail("error creating test manager: %v", err)
	}
	testMgr, err = initializeTestManager()
	if err != nil {
		fail("error creating test manager: %v", err)
	}

	logger = testMgr.GetLogger()

	// Generate a new wallet
	password := "test_password123"

	derivationPath := string(wallet.DerivationPath_Default)
	// Unregistered client
	unregisteredIndex := uint64(9999999)
	unregisteredRecoverResponse, err := recoverWallet(unregisteredTestMgr, "radar blur cabbage chef fix engine embark joy scheme fiction master release", &unregisteredIndex, password, derivationPath)
	if err != nil {
		fail("error recovering unregistered wallet: %v", err)
	}

	// Fully registered client
	index := uint64(0)
	recoverResponse, err := recoverWallet(testMgr, keys.DefaultMnemonic, &index, password, derivationPath)
	if err != nil {
		fail("error generating wallet: %v", err)
	}
	nodeAddress = recoverResponse.Data.AccountAddress

	// Set up NodeSet with the StakeWise vault
	err = setupNodeSet(testMgr, nsEmail, recoverResponse.Data.AccountAddress)
	if err != nil {
		fail("error setup NodeSet: %v", err)
	}

	err = setupNodeSet(unregisteredTestMgr, nsEmail, unregisteredRecoverResponse.Data.AccountAddress)
	if err != nil {
		fail("error setup NodeSet: %v", err)
	}

	// Register with NodeSet
	err = registerNodeSet(testMgr, nsEmail, nodeAddress)
	if err != nil {
		fail("error registering NodeSet: %v", err)
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

func recoverWallet(_testMgr *swtesting.StakeWiseTestManager, mnemonic string, index *uint64, password, derivationPath string) (*types.ApiResponse[api.WalletRecoverData], error) {
	hdClient := _testMgr.HyperdriveTestManager.GetApiClient()
	return hdClient.Wallet.Recover(&derivationPath, mnemonic, index, password, true)
}

func setupNodeSet(_testMgr *swtesting.StakeWiseTestManager, email string, account common.Address) error {
	sp := _testMgr.GetStakeWiseServiceProvider()
	res := sp.GetResources()
	nsServer := _testMgr.GetNodeSetMockServer().GetManager()

	if err := nsServer.AddStakeWiseVault(*res.Vault, res.EthNetworkName); err != nil {
		return fmt.Errorf("error adding stakewise vault to nodeset: %w", err)
	}

	if err := nsServer.AddUser(email); err != nil {
		return fmt.Errorf("error adding user to nodeset: %w", err)
	}

	if err := nsServer.WhitelistNodeAccount(email, account); err != nil {
		return fmt.Errorf("error whitelisting node account: %w", err)
	}

	return nil
}

func registerNodeSet(_testMgr *swtesting.StakeWiseTestManager, email string, nodeAddress common.Address) error {
	logger := log.NewDefaultLogger()
	ctx := logger.CreateContextWithLogger(_testMgr.GetStakeWiseServiceProvider().GetBaseContext())
	nsClient := _testMgr.GetStakeWiseServiceProvider().GetNodesetClient()
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
