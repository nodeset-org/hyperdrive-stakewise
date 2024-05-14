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
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/wallet"
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
	fmt.Printf("!!! publicKeys: %v\n", publicKeys)
	// Fetch status from Nodeset APIs
	nodesetStatusResponse, err := nc.GetRegisteredValidators(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators from Nodeset: %w", err)
	}

	activePubkeysOnNodeset := []beacon.ValidatorPubkey{}
	for _, validator := range nodesetStatusResponse {
		if validator.Status != "PENDING" {
			activePubkeysOnNodeset = append(activePubkeysOnNodeset, validator.Pubkey)
		}
	}
	fmt.Printf("!!! activePubkeysOnNodeset: %v\n", activePubkeysOnNodeset)

	// Process public keys based on their status
	unregisteredKeys := []*eth2types.BLSPrivateKey{}
	data.TotalCount = uint64(len(publicKeys))
	// Used for displaying the unregistered keys in the response
	unregisteredPubkeys := []beacon.ValidatorPubkey{}

	fmt.Printf("!!! Looping over pubkeys...\n")
	for i, pubkey := range publicKeys {
		if !isUploadedToNodeset(pubkey, activePubkeysOnNodeset) {
			unregisteredKeys = append(unregisteredKeys, privateKeys[i])
			unregisteredPubkeys = append(unregisteredPubkeys, pubkey)
		}
	}
	fmt.Printf("!!! unregisteredKeys: %v\n", unregisteredKeys)
	fmt.Printf("!!! unregisteredPubkeys: %v\n", unregisteredPubkeys)

	// Determine if sufficient balance is available for deposits
	balance, err := ec.BalanceAt(ctx, opts.From, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = balance

	validatorDepositCost := eth.EthToWei(0.01)
	totalCost := big.NewInt(int64(len(unregisteredKeys))).Mul(big.NewInt(int64(len(unregisteredKeys))), validatorDepositCost)
	data.RequiredBalance = totalCost
	data.SufficientBalance = (totalCost.Cmp(balance) <= 0)

	if !data.SufficientBalance {
		fmt.Printf("!!! Insufficient balance for deposits: %v\n", balance)
		for len(unregisteredKeys) > 0 && totalCost.Cmp(balance) > 0 {
			unregisteredKeys = unregisteredKeys[:len(unregisteredKeys)-1]
			unregisteredPubkeys = unregisteredPubkeys[:len(unregisteredPubkeys)-1]
			totalCost = totalCost.Sub(totalCost, validatorDepositCost)
		}
		data.UnregisteredPubkeys = unregisteredPubkeys
		data.RequiredBalance = totalCost
		fmt.Printf("!!! New unregistered keys: %v\n", unregisteredPubkeys)
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

	return types.ResponseStatus_Success, nil
}

// TODO: refactor into reusable functions
func isUploadedToNodeset(pubKey beacon.ValidatorPubkey, registeredPubkeys []beacon.ValidatorPubkey) bool {
	for _, registeredPubKey := range registeredPubkeys {
		if registeredPubKey == pubKey {
			return true
		}
	}
	return false
}

// func isRegisteredToStakewise(pubKey beacon.ValidatorPubkey, statuses map[beacon.ValidatorPubkey]beacon.ValidatorStatus) bool {
// 	// TODO: Implement
// 	return false
// }

// func isUploadedStakewise(pubKey beacon.ValidatorPubkey, statuses map[beacon.ValidatorPubkey]beacon.ValidatorStatus) bool {
// 	// TODO: Implement
// 	return false
// }
