package api_test

import (
	"strconv"
	"testing"

	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// NOTE: this uses a lot of SW v1 assumptions which may no longer hold for v2, investigate after the contracts are explored
func TestValidatorStatus_Active(t *testing.T) {
	// Take a snapshot, revert at the end
	err := testMgr.RevertSnapshot(initSnapshot)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer handle_panics()

	// Get some resources
	sp := mainNode.GetServiceProvider()
	res := sp.GetResources()
	wallet := sp.GetWallet()
	nsMock := testMgr.GetNodeSetMockServer().GetManager()
	nsDB := nsMock.GetDatabase()
	deployment := nsDB.StakeWise.GetDeployment(res.DeploymentName)
	vault := deployment.GetVault(res.Vault)
	nsNode, _ := nsDB.Core.GetNode(mainNodeAddress)

	// Generate a validator key
	key, err := wallet.GenerateNewValidatorKey()
	require.NoError(t, err)
	pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
	t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())

	// Generate deposit data
	depositData, err := swcommon.GenerateDepositData(logger, res, []*eth2types.BLSPrivateKey{key})
	require.NoError(t, err)
	t.Log("Deposit data generated")

	// Upload the deposit data to nodeset
	err = vault.HandleDepositDataUpload(nsNode, depositData)
	require.NoError(t, err)
	t.Log("Deposit data uploaded to nodeset")

	// Cut a new deposit data set
	depositDataSet := vault.CreateNewDepositDataSet(1)
	require.Equal(t, depositData, depositDataSet)
	t.Log("New deposit data set created")

	// Upload the deposit data to StakeWise
	vault.UploadDepositDataToStakeWise(depositDataSet)
	t.Log("Deposit data set uploaded to StakeWise")

	// Mark the deposit data set as uploaded
	vault.MarkDepositDataSetUploaded(depositDataSet)
	t.Log("Deposit data set marked as uploaded")

	// Add the validator to Beacon
	creds := validator.GetWithdrawalCredsFromAddress(res.Vault)
	bn := testMgr.GetBeaconMockManager()
	validator, err := bn.AddValidator(pubkey, creds)
	require.NoError(t, err)
	t.Log("Validator added to the beacon chain")

	// Mark the validator as active
	vault.MarkValidatorsRegistered(depositDataSet)
	t.Log("Deposit data set marked as registered")

	// Set the validator to active
	validator.Status = beacon.ValidatorState_ActiveOngoing
	validator.Index = 1

	// Run the status route
	client := mainNode.GetApiClient()
	response, err := client.Status.GetValidatorStatuses()
	require.NoError(t, err)
	t.Log("Ran validator status check")

	// Check the response
	require.Len(t, response.Data.States, 1)
	responseValidator := response.Data.States[0]
	require.Equal(t, pubkey, responseValidator.Pubkey)
	require.True(t, responseValidator.NodeSet.Registered)
	require.False(t, responseValidator.NodeSet.ExitMessageUploaded)
	require.Equal(t, beacon.ValidatorState_ActiveOngoing, responseValidator.BeaconStatus)
	require.Equal(t, strconv.FormatUint(validator.Index, 10), responseValidator.Index)
	t.Logf("Validator was active, index = %s", responseValidator.Index)
}
