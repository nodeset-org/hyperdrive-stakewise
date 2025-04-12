package api_test

import (
	"context"
	"runtime/debug"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	hdtesting "github.com/nodeset-org/hyperdrive-daemon/testing"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

func TestRelay_Success(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(hdtesting.Service_EthClients | hdtesting.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer relay_cleanup(snapshotName)

	// Get some resources
	sp := mainNode.GetServiceProvider()
	res := sp.GetResources()
	wallet := sp.GetWallet()
	nsMock := testMgr.GetNodeSetMockServer().GetManager()
	nsDB := nsMock.GetDatabase()
	deployment := nsDB.StakeWise.GetDeployment(res.DeploymentName)
	vault := deployment.GetVault(res.Vault)
	op := testMgr.GetOperatorMock()
	keyMgr := sp.GetAvailableKeyManager()
	logger := testMgr.GetLogger()

	// Generate a validator key
	key, err := wallet.GenerateNewValidatorKey()
	require.NoError(t, err)
	pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
	t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())

	// Set the max validators per node to 1
	vault.MaxValidatorsPerUser = 1

	// Commit a block just so the latest block is fresh - otherwise the sync progress check will
	// error out because the block is too old and it thinks the client just can't find any peers
	err = testMgr.CommitBlock()
	if err != nil {
		t.Fatalf("Error committing block: %v", err)
	}

	// Initialize the key manager
	keyMgr.LoadPrivateKeys(logger)
	goodKeys, badKeys, err := keyMgr.GetAvailableKeys(context.Background(), logger, common.HexToHash("0x01"), swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: true,
	})
	require.NoError(t, err)
	require.Len(t, goodKeys, 1)
	require.Len(t, badKeys, 0)

	// Run the relay
	resp, err := op.SubmitValidatorsRequest()
	require.NoError(t, err)

	require.Len(t, resp.Validators, 1)
	require.Equal(t, resp.Validators[0].PublicKey, pubkey)
}

// Clean up after each test
func relay_cleanup(snapshotName string) {
	// Handle panics
	r := recover()
	if r != nil {
		debug.PrintStack()
		fail("Recovered from panic: %v", r)
	}

	// Revert to the snapshot taken at the start of the test
	err := testMgr.RevertToCustomSnapshot(snapshotName)
	if err != nil {
		fail("Error reverting to custom snapshot: %v", err)
	}

	// Reload the HD wallet to undo any changes made during the test
	err = mainNode.GetHyperdriveNode().GetServiceProvider().GetWallet().Reload(testMgr.GetLogger())
	if err != nil {
		fail("Error reloading hyperdrive wallet: %v", err)
	}

	// Reload the SW wallet to undo any changes made during the test
	err = mainNode.GetServiceProvider().GetWallet().Reload()
	if err != nil {
		fail("Error reloading stakewise wallet: %v", err)
	}

	// Reload the key manager to undo any changes made during the test
	err = mainNode.GetServiceProvider().GetAvailableKeyManager().Reload()
	if err != nil {
		fail("Error reloading available key manager: %v", err)
	}
}
