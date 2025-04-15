package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	batchquery "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/stretchr/testify/require"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestRelay_One(t *testing.T) {
	// Revert to the initial setup
	err := testMgr.RevertSnapshot(initSnapshot)
	if err != nil {
		fail("Error reverting to initial snapshot: %v", err)
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
	_, _, err = keyMgr.GetAvailableKeys(context.Background(), logger, common.HexToHash("0x01"), swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: true,
	})
	require.NoError(t, err)

	// Run the relay
	resp, err := op.SubmitValidatorsRequest()
	require.NoError(t, err)

	require.Len(t, resp.Validators, 1)
	require.Equal(t, resp.Validators[0].PublicKey, pubkey)
}

func TestRelay_Three(t *testing.T) {
	// Revert to the initial setup
	err := testMgr.RevertSnapshot(initSnapshot)
	if err != nil {
		fail("Error reverting to initial snapshot: %v", err)
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
	op := testMgr.GetOperatorMock()
	keyMgr := sp.GetAvailableKeyManager()
	logger := testMgr.GetLogger()

	// Generate three validator keys
	pubkeys := make([]beacon.ValidatorPubkey, 3)
	for i := 0; i < 3; i++ {
		key, err := wallet.GenerateNewValidatorKey()
		require.NoError(t, err)
		pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
		pubkeys[i] = pubkey
		t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())
	}

	// Set the max validators per node to 1
	vault.MaxValidatorsPerUser = 3

	// Commit a block just so the latest block is fresh - otherwise the sync progress check will
	// error out because the block is too old and it thinks the client just can't find any peers
	err = testMgr.CommitBlock()
	if err != nil {
		t.Fatalf("Error committing block: %v", err)
	}

	// Initialize the key manager
	keyMgr.LoadPrivateKeys(logger)
	_, _, err = keyMgr.GetAvailableKeys(context.Background(), logger, common.HexToHash("0x01"), swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: true,
	})
	require.NoError(t, err)

	// Run the relay
	resp, err := op.SubmitValidatorsRequest()
	require.NoError(t, err)

	require.Len(t, resp.Validators, 3)
	for i, pubkey := range pubkeys {
		require.Equal(t, resp.Validators[i].PublicKey, pubkey)
	}
}

func TestRelay_Staggered(t *testing.T) {
	// Revert to the initial setup
	err := testMgr.RevertSnapshot(initSnapshot)
	if err != nil {
		fail("Error reverting to initial snapshot: %v", err)
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
	op := testMgr.GetOperatorMock()
	keyMgr := sp.GetAvailableKeyManager()
	logger := testMgr.GetLogger()
	qMgr := sp.GetQueryManager()
	bdc := sp.GetBeaconDepositContract()

	// Generate three validator keys
	pubkeys := make([]beacon.ValidatorPubkey, 3)
	for i := 0; i < 3; i++ {
		key, err := wallet.GenerateNewValidatorKey()
		require.NoError(t, err)
		pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
		pubkeys[i] = pubkey
		t.Logf("Validator key generated, pubkey = %s", pubkey.HexWithPrefix())
	}

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
	_, _, err = keyMgr.GetAvailableKeys(context.Background(), logger, common.HexToHash("0x01"), swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: true,
	})
	require.NoError(t, err)

	// Run the relay
	resp, err := op.SubmitValidatorsRequest()
	require.NoError(t, err)
	require.Equal(t, resp.Validators[0].PublicKey, pubkeys[0])
	t.Log("First key was submitted for deposits as expected")

	// Try again and make sure the first key reports as unavailable
	resp, err = op.SubmitValidatorsRequest()
	require.NoError(t, err)
	require.Empty(t, resp.Validators)
	t.Log("No keys were submitted for deposits as expected")

	// Change the deposit root
	key, err := keygen.GetBlsPrivateKey(1000)
	require.NoError(t, err)
	err = deposit(key, mainNodeOpts)
	require.NoError(t, err)
	var depositRoot common.Hash
	err = qMgr.Query(func(mc *batchquery.MultiCaller) error {
		bdc.GetDepositRoot(mc, &depositRoot)
		return nil
	}, nil)
	require.NoError(t, err)
	nsDB.Eth.SetDepositRoot(depositRoot)
	t.Logf("Performed out-of-band deposit, deposit root is now %s", depositRoot.Hex())

	// Try again, make sure the first key reports as available now
	resp, err = op.SubmitValidatorsRequest()
	require.NoError(t, err)
	require.Len(t, resp.Validators, 1)
	require.Equal(t, resp.Validators[0].PublicKey, pubkeys[0])
	t.Log("First key was submitted for deposits again as expected")

	// Perform the (faked) deposit
	key, err = keygen.GetBlsPrivateKey(0)
	require.NoError(t, err)
	require.Equal(t, key.PublicKey().Marshal(), pubkeys[0][:])
	err = deposit(key, mainNodeOpts)
	require.NoError(t, err)
	err = qMgr.Query(func(mc *batchquery.MultiCaller) error {
		bdc.GetDepositRoot(mc, &depositRoot)
		return nil
	}, nil)
	require.NoError(t, err)
	nsDB.Eth.SetDepositRoot(depositRoot)
	validator0, exists := vault.Validators[mainNodeAddress][pubkeys[0]]
	require.True(t, exists)
	validator0.HasDepositEvent = true
	t.Logf("Fake deposited for validator 0, deposit root is now %s", depositRoot.Hex())

	// Increase the max validators per node to 3 - the next 2 should show up
	vault.MaxValidatorsPerUser = 3
	resp, err = op.SubmitValidatorsRequest()
	require.NoError(t, err)
	require.Len(t, resp.Validators, 2)
	require.Equal(t, resp.Validators[0].PublicKey, pubkeys[1])
	require.Equal(t, resp.Validators[1].PublicKey, pubkeys[2])
	t.Log("Next 2 keys were submitted for deposits as expected")
}

// Perform a deposit to the Beacon deposit contract
func deposit(key *eth2types.BLSPrivateKey, opts *bind.TransactOpts) error {
	sp := mainNode.GetServiceProvider()
	bdc := sp.GetBeaconDepositContract()
	txMgr := sp.GetTransactionManager()
	res := sp.GetResources()

	// Create the deposit data
	depositDatas, err := swcommon.GenerateDepositData(
		logger,
		res,
		[]*eth2types.BLSPrivateKey{key},
	)
	if err != nil {
		return fmt.Errorf("error generating deposit data: %w", err)
	}
	depositData := depositDatas[0]

	// Create the deposit TX
	submissions, err := eth.BatchCreateTransactionSubmissions([]func() (string, *eth.TransactionInfo, error){
		func() (string, *eth.TransactionInfo, error) {
			txInfo, err := bdc.Deposit(
				beacon.ValidatorPubkey(depositData.PublicKey),
				common.Hash(depositData.WithdrawalCredentials),
				beacon.ValidatorSignature(depositData.Signature),
				common.Hash(depositData.DepositDataRoot),
				&bind.TransactOpts{
					From:      opts.From,
					Signer:    opts.Signer,
					Nonce:     nil,
					Context:   opts.Context,
					GasFeeCap: opts.GasFeeCap,
					GasTipCap: opts.GasTipCap,
					Value:     eth.GweiToWei(float64(depositData.Amount)),
				},
			)
			return "creating Beacon deposit", txInfo, err
		},
	}, true)
	if err != nil {
		return fmt.Errorf("error creating transaction submission: %w", err)
	}

	// Execute the transaction
	txs, err := txMgr.BatchExecuteTransactions(submissions, &bind.TransactOpts{
		From:      opts.From,
		Signer:    opts.Signer,
		Nonce:     nil,
		Context:   opts.Context,
		GasFeeCap: opts.GasFeeCap,
		GasTipCap: opts.GasTipCap,
	})
	if err != nil {
		return fmt.Errorf("error submitting mint transactions: %w", err)
	}

	// Mine the block
	err = testMgr.CommitBlock()
	if err != nil {
		return fmt.Errorf("error committing block: %w", err)
	}
	err = txMgr.WaitForTransactions(txs)
	if err != nil {
		return fmt.Errorf("error waiting for deploy transactions: %w", err)
	}
	return nil
}
