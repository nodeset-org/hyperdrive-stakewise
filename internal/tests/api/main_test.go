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
	testMgr     *swtesting.StakeWiseTestManager = nil
	wg          *sync.WaitGroup                 = nil
	logger      *slog.Logger                    = nil
	nodeAddress common.Address
	nsEmail     string = "test@nodeset.io"
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	wg = &sync.WaitGroup{}
	var err error

	// Create a new test manager
	testMgr, err = swtesting.NewStakeWiseTestManager("localhost", "localhost", "localhost")
	if err != nil {
		fail("error creating test manager: %v", err)
	}
	logger = testMgr.GetLogger()

	// Generate a new wallet
	derivationPath := string(wallet.DerivationPath_Default)
	index := uint64(0)
	password := "test_password123"
	hdClient := testMgr.HyperdriveTestManager.GetApiClient()
	recoverResponse, err := hdClient.Wallet.Recover(&derivationPath, keys.DefaultMnemonic, &index, password, true)
	if err != nil {
		fail("error generating wallet: %v", err)
	}
	nodeAddress = recoverResponse.Data.AccountAddress

	// Set up NodeSet with the StakeWise vault
	sp := testMgr.GetStakeWiseServiceProvider()
	res := sp.GetResources()
	nsServer := testMgr.GetNodeSetMockServer().GetManager()
	err = nsServer.AddStakeWiseVault(*res.Vault, res.EthNetworkName)
	if err != nil {
		fail("error adding stakewise vault to nodeset: %v", err)
	}

	// Make a NodeSet account
	err = nsServer.AddUser(nsEmail)
	if err != nil {
		fail("error adding user to nodeset: %v", err)
	}
	err = nsServer.WhitelistNodeAccount(nsEmail, nodeAddress)
	if err != nil {
		fail("error adding node account to nodeset: %v", err)
	}

	// Register with NodeSet
	response, err := hdClient.NodeSet.RegisterNode(nsEmail)
	if err != nil {
		fail("error registering node with nodeset: %v", err)
	}
	if response.Data.AlreadyRegistered {
		fail("node is already registered with nodeset")
	}
	if response.Data.NotWhitelisted {
		fail("node is not whitelisted with a nodeset user account")
	}

	// Run tests
	code := m.Run()

	// Clean up and exit
	cleanup()
	os.Exit(code)
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
