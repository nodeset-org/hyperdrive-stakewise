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
	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Fetch private keys and derive public keys
	privateKeys, err := w.GetAllPrivateKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting private keys: %w", err)
	}
	publicKeys, err := w.DerivePubKeys(privateKeys)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error deriving public keys: %w", err)
	}
	publicKeyMap := make(map[beacon.ValidatorPubkey]bool)
	for _, pubkey := range publicKeys {
		publicKeyMap[pubkey] = true
	}

	// Fetch status from Nodeset APIs
	nodesetStatusResponse, err := nc.GetRegisteredValidators(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators from Nodeset: %w", err)
	}

	activePubkeysOnNodeset := []beacon.ValidatorPubkey{}
	pendingPubkeysOnNodeset := []beacon.ValidatorPubkey{}
	newPublicKeys := []beacon.ValidatorPubkey{}

	for _, validator := range nodesetStatusResponse {
		_, exists := publicKeyMap[validator.Pubkey]
		if exists {
			publicKeyMap[validator.Pubkey] = false

			if validator.Status != pendingState {
				activePubkeysOnNodeset = append(activePubkeysOnNodeset, validator.Pubkey)
			} else {
				pendingPubkeysOnNodeset = append(pendingPubkeysOnNodeset, validator.Pubkey)
			}
		} else {
			newPublicKeys = append(newPublicKeys, validator.Pubkey)
		}
	}
	publicKeys = newPublicKeys

	// Process public keys based on their status
	unregisteredKeys := []*eth2types.BLSPrivateKey{}
	data.TotalCount = uint64(len(publicKeys))
	// Used for displaying the unregistered keys in the response
	unregisteredPubkeys := []beacon.ValidatorPubkey{}

	for i, pubkey := range publicKeys {
		if !swcommon.IsUploadedToNodeset(pubkey, activePubkeysOnNodeset) {
			unregisteredKeys = append(unregisteredKeys, privateKeys[i])
			unregisteredPubkeys = append(unregisteredPubkeys, pubkey)
		}
	}

	// Determine if sufficient balance is available for deposits
	balance, err := ec.BalanceAt(ctx, opts.From, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = balance

	totalCost := big.NewInt(0)
	costPerKey := eth.EthToWei(validatorDepositCost)

	unregisteredKeysCount := len(unregisteredKeys)
	pendingPubkeysOnNodesetCount := len(pendingPubkeysOnNodeset)

	totalCost.Add(totalCost, new(big.Int).Mul(costPerKey, big.NewInt(int64(unregisteredKeysCount+pendingPubkeysOnNodesetCount))))

	data.SufficientBalance = (totalCost.Cmp(balance) <= 0)

	if !data.SufficientBalance {
		for totalCost.Cmp(balance) > 0 {
			unregisteredKeys = unregisteredKeys[1:]
			unregisteredPubkeys = unregisteredPubkeys[1:]
			totalCost = totalCost.Sub(totalCost, costPerKey)
		}
		data.UnregisteredPubkeys = unregisteredPubkeys
	}

	if len(unregisteredKeys) == 0 {
		return types.ResponseStatus_Success, nil
	}
	// Generate deposit data and submit
	depositData, err := ddMgr.GenerateDepositData(unregisteredKeys)
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

	remainingCost := new(big.Int).Sub(totalCost, balance)
	if remainingCost.Sign() < 0 {
		remainingCost.SetInt64(0)
	}
	data.RemainingEthRequired = eth.WeiToEth(remainingCost)

	data.EthPerKey = validatorDepositCost
	return types.ResponseStatus_Success, nil
}
