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
	validatorDepositCost := eth.EthToWei(0.01)

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
	for _, validator := range nodesetStatusResponse {
		_, exists := publicKeyMap[validator.Pubkey]
		if exists {
			publicKeyMap[validator.Pubkey] = false

			if validator.Status != "PENDING" {
				activePubkeysOnNodeset = append(activePubkeysOnNodeset, validator.Pubkey)
			} else {
				pendingPubkeysOnNodeset = append(pendingPubkeysOnNodeset, validator.Pubkey)
			}
		}
	}

	newPublicKeys := []beacon.ValidatorPubkey{}
	for pubkey, unprocessed := range publicKeyMap {
		if unprocessed {
			newPublicKeys = append(newPublicKeys, pubkey)
		}
	}
	publicKeys = newPublicKeys
	fmt.Printf("!!! len(publicKeys): %v\n", len(publicKeys))

	fmt.Printf("!!! len(activePubkeysOnNodeset): %v\n", len(activePubkeysOnNodeset))
	fmt.Printf("!!! len(pendingPubkeysOnNodeset): %v\n", len(pendingPubkeysOnNodeset))

	// Process public keys based on their status
	unregisteredKeys := []*eth2types.BLSPrivateKey{}
	data.TotalCount = uint64(len(publicKeys))
	// Used for displaying the unregistered keys in the response
	unregisteredPubkeys := []beacon.ValidatorPubkey{}

	for i, pubkey := range publicKeys {
		if !isUploadedToNodeset(pubkey, activePubkeysOnNodeset) {
			unregisteredKeys = append(unregisteredKeys, privateKeys[i])
			unregisteredPubkeys = append(unregisteredPubkeys, pubkey)
		}
	}
	fmt.Printf("!!! len(unregisteredKeys): %v\n", len(unregisteredKeys))

	// Determine if sufficient balance is available for deposits
	balance, err := ec.BalanceAt(ctx, opts.From, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = balance

	totalCost := new(big.Int)

	// Add deposit cost for each unregistered key
	for range unregisteredKeys {
		totalCost.Add(totalCost, validatorDepositCost)
	}
	// Total cost needs to account for pending validators
	for range pendingPubkeysOnNodeset {
		totalCost = totalCost.Add(totalCost, validatorDepositCost)
	}

	fmt.Printf("!!! initial Total cost: %v\n", totalCost)
	data.SufficientBalance = (totalCost.Cmp(balance) <= 0)

	if !data.SufficientBalance {
		fmt.Printf("!!! Insufficient balance for deposits: %v\n", balance)
		fmt.Printf("!!! initial unregistered keys: %v\n", len(unregisteredPubkeys))
		fmt.Printf("!!! initial total cost: %v\n", totalCost)
		for len(unregisteredKeys) > 0 && totalCost.Cmp(balance) > 0 {
			unregisteredKeys = unregisteredKeys[1:]
			unregisteredPubkeys = unregisteredPubkeys[1:]
			totalCost = totalCost.Sub(totalCost, validatorDepositCost)
			fmt.Printf("!!! interm values: %v, %v, %v\n", len(unregisteredKeys), len(unregisteredPubkeys), totalCost)

		}
		data.UnregisteredPubkeys = unregisteredPubkeys
		fmt.Printf("!!! final unregistered keys: %v\n", unregisteredPubkeys)
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
