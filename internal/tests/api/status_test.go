package api_test

import (
	"runtime/debug"
	"strconv"
	"testing"

	swtypes "github.com/nodeset-org/hyperdrive-stakewise/shared/types"
	"github.com/nodeset-org/osha"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestValidatorStatus_Active(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	// Get some resources
	sp := testMgr.GetStakeWiseServiceProvider()
	vault := *sp.GetResources().Vault
	network := sp.GetResources().EthNetworkName
	wallet := sp.GetWallet()
	ddMgr := sp.GetDepositDataManager()
	nsMock := testMgr.GetNodeSetMockServer().GetManager()

	// Generate a validator key
	key, err := wallet.GenerateNewValidatorKey()
	require.NoError(t, err)
	pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
	t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())

	// Generate deposit data
	depositDataPtrs, err := ddMgr.GenerateDepositData([]*eth2types.BLSPrivateKey{key})
	require.NoError(t, err)
	depositData := make([]beacon.ExtendedDepositData, len(depositDataPtrs))
	for i, ptr := range depositDataPtrs {
		depositData[i] = (*ptr).ExtendedDepositData
	}
	t.Log("Deposit data generated")

	// Upload the deposit data to nodeset
	err = nsMock.HandleDepositDataUpload(nodeAddress, depositData)
	require.NoError(t, err)
	t.Log("Deposit data uploaded to nodeset")

	// Cut a new deposit data set
	depositDataSet := nsMock.CreateNewDepositDataSet(network, 1)
	require.Equal(t, depositData, depositDataSet)
	t.Log("New deposit data set created")

	// Upload the deposit data to StakeWise
	err = nsMock.UploadDepositDataToStakeWise(vault, network, depositDataSet)
	require.NoError(t, err)
	t.Log("Deposit data set uploaded to StakeWise")

	// Mark the deposit data set as uploaded
	err = nsMock.MarkDepositDataSetUploaded(vault, network, depositDataSet)
	require.NoError(t, err)
	t.Log("Deposit data set marked as uploaded")

	// Add the validator to Beacon
	creds := validator.GetWithdrawalCredsFromAddress(vault)
	bn := testMgr.GetBeaconMockManager()
	validator, err := bn.AddValidator(pubkey, creds)
	require.NoError(t, err)
	t.Log("Validator added to the beacon chain")

	// Mark the validator as active
	err = nsMock.MarkValidatorsRegistered(vault, network, depositDataSet)
	require.NoError(t, err)
	t.Log("Deposit data set marked as registered")

	// Set the validator to active
	validator.Status = beacon.ValidatorState_ActiveOngoing
	validator.Index = 1

	// Run the status route
	client := testMgr.GetApiClient()
	response, err := client.Status.GetValidatorStatuses()
	require.NoError(t, err)
	t.Log("Ran validator status check")

	// Check the response
	require.Len(t, response.Data.States, 1)
	responseValidator := response.Data.States[0]
	require.Equal(t, swtypes.NodesetStatus_RegisteredToStakewise, responseValidator.NodesetStatus)
	require.Equal(t, beacon.ValidatorState_ActiveOngoing, responseValidator.BeaconStatus)
	require.Equal(t, strconv.FormatUint(validator.Index, 10), responseValidator.Index)
	t.Logf("Validator was active, index = %s", responseValidator.Index)
}

// Clean up after each test
func status_cleanup(snapshotName string) {
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
	err = testMgr.GetServiceProvider().GetWallet().Reload(testMgr.GetLogger())
	if err != nil {
		fail("Error reloading hyperdrive wallet: %v", err)
	}

	// Reload the SW wallet to undo any changes made during the test
	err = testMgr.GetStakeWiseServiceProvider().GetWallet().Reload()
	if err != nil {
		fail("Error reloading stakewise wallet: %v", err)
	}

	// Log out of the NS mock server
	testMgr.GetStakeWiseServiceProvider().GetNodesetClient().Logout()
}
