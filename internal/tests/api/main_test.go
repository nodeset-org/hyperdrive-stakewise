package api_test

import (
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swtesting "github.com/nodeset-org/hyperdrive-stakewise/testing"
	"github.com/nodeset-org/osha/keys"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Various singleton variables used for testing
var (
	testMgr *swtesting.StakeWiseTestManager = nil
	logger  *slog.Logger                    = nil
	nsEmail string                          = "test@nodeset.io"

	// CS nodes
	mainNode        *swtesting.StakeWiseNode
	mainNodeAddress common.Address
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	// Create a new test manager
	var err error
	testMgr, err = swtesting.NewStakeWiseTestManager()
	if err != nil {
		fail("error creating test manager: %v", err)
	}
	logger = testMgr.GetLogger()
	mainNode = testMgr.GetNode()

	// Generate a new wallet
	derivationPath := string(wallet.DerivationPath_Default)
	index := uint64(0)
	password := "test_password123"
	hdNode := mainNode.GetHyperdriveNode()
	hd := hdNode.GetApiClient()
	recoverResponse, err := hd.Wallet.Recover(&derivationPath, keys.DefaultMnemonic, &index, password, true)
	if err != nil {
		fail("error generating wallet: %v", err)
	}
	mainNodeAddress = recoverResponse.Data.AccountAddress

	// Set up NodeSet with the StakeWise vault
	sp := mainNode.GetServiceProvider()
	res := sp.GetResources()
	nsMgr := testMgr.GetNodeSetMockServer().GetManager()
	nsDB := nsMgr.GetDatabase()
	deployment := nsDB.StakeWise.AddDeployment(res.DeploymentName, big.NewInt(int64(res.ChainID)))
	_ = deployment.AddVault(res.Vault)
	nsDB.SetSecretEncryptionIdentity(hdtesting.EncryptionIdentity)

	// Make a NodeSet account
	_, err = nsDB.Core.AddUser(nsEmail)
	if err != nil {
		fail("error adding user to nodeset: %v", err)
	}

	// Register the primary
	err = registerWithNodeset(mainNode, mainNodeAddress)
	if err != nil {
		fail("error registering with nodeset: %v", err)
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

// Register a node with nodeset
func registerWithNodeset(node *swtesting.StakeWiseNode, address common.Address) error {
	// whitelist the node with the nodeset.io account
	nsServer := testMgr.GetNodeSetMockServer().GetManager()
	nsDB := nsServer.GetDatabase()
	user := nsDB.Core.GetUser(nsEmail)
	_ = user.WhitelistNode(address)

	// Register with NodeSet
	hd := node.GetHyperdriveNode().GetApiClient()
	response, err := hd.NodeSet.RegisterNode(nsEmail)
	if err != nil {
		fail("error registering node with nodeset: %v", err)
	}
	if response.Data.AlreadyRegistered {
		fail("node is already registered with nodeset")
	}
	if response.Data.NotWhitelisted {
		fail("node is not whitelisted with a nodeset user account")
	}
	return nil
}
