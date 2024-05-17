package swnodeset

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/goccy/go-json"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/eth"

	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/wallet"
)

const (
	pendingState         string  = "PENDING"
	validatorDepositCost float64 = 0.01
)

// ===============
// === Factory ===
// ===============

type nodesetUploadDepositDataContextFactory struct {
	handler *NodesetHandler
}

func (f *nodesetUploadDepositDataContextFactory) Create(args url.Values) (*nodesetUploadDepositDataContext, error) {
	c := &nodesetUploadDepositDataContext{
		handler: f.handler,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *nodesetUploadDepositDataContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetUploadDepositDataContext, swapi.NodesetUploadDepositDataData](
		router, "upload-deposit-data", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============
type nodesetUploadDepositDataContext struct {
	handler *NodesetHandler
}

func (c *nodesetUploadDepositDataContext) PrepareData(data *swapi.NodesetUploadDepositDataData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ddMgr := sp.GetDepositDataManager()
	nc := sp.GetNodesetClient()
	w := sp.GetWallet()
	ec := sp.GetEthClient()
	ctx := c.handler.ctx
	nodeAddress := walletStatus.Address.NodeAddress

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Fetch status from Nodeset APIs
	nodesetStatusResponse, err := nc.GetRegisteredValidators(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators from Nodeset: %w", err)
	}

	// Fetch private keys and derive public keys
	privateKeys, err := w.GetAllPrivateKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting private keys: %w", err)
	}

	// Create maps of private keys for easy lookup
	privateKeyMap := make(map[beacon.ValidatorPubkey]*eth2types.BLSPrivateKey)
	publicKeyMap := make(map[beacon.ValidatorPubkey]bool)
	publicKeys := []beacon.ValidatorPubkey{}
	for _, privateKey := range privateKeys {
		pubkey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
		publicKeys = append(publicKeys, pubkey)
		publicKeyMap[pubkey] = true
		privateKeyMap[pubkey] = privateKey
	}

	// Sort each key by its upload status
	activePubkeysOnNodeset := []beacon.ValidatorPubkey{}
	pendingPubkeysOnNodeset := []beacon.ValidatorPubkey{}
	for _, validator := range nodesetStatusResponse {
		_, exists := publicKeyMap[validator.Pubkey]
		if exists {
			if validator.Status != pendingState {
				activePubkeysOnNodeset = append(activePubkeysOnNodeset, validator.Pubkey)
			} else {
				pendingPubkeysOnNodeset = append(pendingPubkeysOnNodeset, validator.Pubkey)
			}
		}
	}
	data.TotalCount = uint64(len(publicKeys))

	// Create a list of unregistered keys
	unregisteredKeys := []*eth2types.BLSPrivateKey{}
	unregisteredPubkeys := []beacon.ValidatorPubkey{} // Used for displaying the unregistered keys in the response
	for _, pubkey := range publicKeys {
		if !swcommon.IsUploadedToNodeset(pubkey, activePubkeysOnNodeset) && !swcommon.IsUploadedToNodeset(pubkey, pendingPubkeysOnNodeset) {
			unregisteredKeys = append(unregisteredKeys, privateKeyMap[pubkey])
			unregisteredPubkeys = append(unregisteredPubkeys, pubkey)
		}
	}

	// Short circuit if all keys are already registered
	if len(unregisteredKeys) == 0 {
		return types.ResponseStatus_Success, nil
	}

	// Get the wallet's ETH balance
	err = sp.RequireEthClientSynced(ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}
	balance, err := ec.BalanceAt(ctx, nodeAddress, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = eth.WeiToEth(balance)

	// Subtract the cost of the pending keys
	data.EthPerKey = validatorDepositCost
	costPerKeyBig := eth.EthToWei(validatorDepositCost)
	pendingCountBig := big.NewInt(int64(len(pendingPubkeysOnNodeset)))
	pendingCost := big.NewInt(0).Mul(costPerKeyBig, pendingCountBig)
	remainingBalance := big.NewInt(0).Sub(balance, pendingCost)

	// Register as many keys as possible with the remaining balance
	registeredKeys := []*eth2types.BLSPrivateKey{}
	remainingKeys := []*eth2types.BLSPrivateKey{}
	data.NewPubkeys = []beacon.ValidatorPubkey{}
	data.RemainingPubkeys = []beacon.ValidatorPubkey{}
	for i, unregisteredKey := range unregisteredKeys {
		pubkey := unregisteredPubkeys[i]
		if costPerKeyBig.Cmp(remainingBalance) > 0 {
			// Balance is insufficient, label this key as remaining
			remainingKeys = append(remainingKeys, unregisteredKey)
			data.RemainingPubkeys = append(data.RemainingPubkeys, pubkey)
		} else {
			// Balance is sufficient, register this key and subtract its cost
			registeredKeys = append(registeredKeys, unregisteredKey)
			data.NewPubkeys = append(data.NewPubkeys, pubkey)
			remainingBalance.Sub(remainingBalance, costPerKeyBig)
		}
	}
	data.SufficientBalance = (len(remainingKeys) == 0)

	// Get how much ETH is required to finish registering the remaining keys
	if !data.SufficientBalance {
		remainingKeysBig := big.NewInt(int64(len(remainingKeys)))
		costOfRemainingKeys := big.NewInt(0).Mul(remainingKeysBig, costPerKeyBig)
		remainingEthRequired := big.NewInt(0).Sub(costOfRemainingKeys, remainingBalance)
		data.RemainingEthRequired = eth.WeiToEth(remainingEthRequired)
	}

	// Generate deposit data and submit
	depositData, err := ddMgr.GenerateDepositData(registeredKeys)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error generating deposit data: %w", err)
	}
	serializedData, err := json.Marshal(depositData)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error serializing deposit data: %w", err)
	}
	if response, err := nc.UploadDepositData(ctx, serializedData); err != nil {
		return types.ResponseStatus_Error, err
	} else {
		data.ServerResponse = response
	}

	return types.ResponseStatus_Success, nil
}

// Temp function for unit testing demo
func getRegisterableKeys(pendingKeys []beacon.ValidatorPubkey, unregisteredPubkeys []beacon.ValidatorPubkey, walletBalance *big.Int) ([]beacon.ValidatorPubkey, []beacon.ValidatorPubkey, *big.Int) {
	costPerKeyBig := eth.EthToWei(validatorDepositCost)
	pendingCountBig := big.NewInt(int64(len(pendingKeys)))
	pendingCost := big.NewInt(0).Mul(costPerKeyBig, pendingCountBig)
	remainingBalance := big.NewInt(0).Sub(walletBalance, pendingCost)

	// Register as many keys as possible with the remaining balance
	registeredPubkeys := []beacon.ValidatorPubkey{}
	remainingPubkeys := []beacon.ValidatorPubkey{}
	for i, unregisteredKey := range unregisteredPubkeys {
		pubkey := unregisteredPubkeys[i]
		if costPerKeyBig.Cmp(remainingBalance) > 0 {
			// Balance is insufficient, label this key as remaining
			remainingPubkeys = append(remainingPubkeys, pubkey)
		} else {
			// Balance is sufficient, register this key and subtract its cost
			registeredPubkeys = append(registeredPubkeys, unregisteredKey)
			remainingBalance.Sub(remainingBalance, costPerKeyBig)
		}
	}

	return registeredPubkeys, remainingPubkeys, remainingBalance
}
