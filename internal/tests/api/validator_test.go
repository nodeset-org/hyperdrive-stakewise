package api_test

import (
	"testing"

	"github.com/nodeset-org/osha"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
)

func TestExit_EmptyPubkeys(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)
	client := testMgr.GetApiClient()
	pubkeys := []beacon.ValidatorPubkey{}

	_, err = client.Validator.Exit(pubkeys, nil, false)
	require.Error(t, err)
}

func TestExit(t *testing.T) {
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	sp := testMgr.GetStakeWiseServiceProvider()
	wallet := sp.GetWallet()
	vault := *sp.GetResources().Vault

	// Generate a validator key
	key, err := wallet.GenerateNewValidatorKey()
	require.NoError(t, err)
	pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
	t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())

	// Add the validator to Beacon
	creds := validator.GetWithdrawalCredsFromAddress(vault)
	bn := testMgr.GetBeaconMockManager()
	_, err = bn.AddValidator(pubkey, creds)
	require.NoError(t, err)
	t.Log("Validator added to the beacon chain")

	client := testMgr.GetApiClient()
	// var exitEpoch uint64 = 1
	response, err := client.Validator.Exit([]beacon.ValidatorPubkey{pubkey}, nil, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(response.Data.ExitInfos))
}
