package api_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/osha"
	"github.com/stretchr/testify/require"
)

func TestWalletInit_Success(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	client := testMgr.GetApiClient()

	response, err := client.Wallet.Initialize()
	require.NoError(t, err)
	t.Logf("Wallet initialized: %v", response)
	require.Equal(t, common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), response.Data.AccountAddress)
}

func TestWalletGenerateKeys_SingleKey(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	client := testMgr.GetApiClient()

	response, err := client.Wallet.GenerateKeys(1, false)
	require.NoError(t, err)
	t.Logf("Generated key: %v", response)
	require.Equal(t, 1, len(response.Data.Pubkeys))
}

func TestWalletGenerateKeys_MultipleKeys(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	client := testMgr.GetApiClient()

	response, err := client.Wallet.GenerateKeys(5, false)
	require.NoError(t, err)
	t.Logf("Generated keys: %v", response)
	require.Equal(t, 5, len(response.Data.Pubkeys))
}

func TestWalletClaimRewards(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(snapshotName)

	client := testMgr.GetApiClient()

	response, err := client.Wallet.ClaimRewards()
	require.NoError(t, err)
	t.Logf("Claimed rewards: %v", response)
	require.Equal(t, "0", response.Data.TokenName)
}
