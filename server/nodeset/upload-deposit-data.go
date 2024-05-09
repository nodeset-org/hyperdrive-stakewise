package swnodeset

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/wallet"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
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
const stateFilePath = "upload_state.json"

type nodesetUploadDepositDataContext struct {
	handler         *NodesetHandler
	sessionUploaded map[beacon.ValidatorPubkey]bool
	lastBalance     *big.Int
}

type uploadState struct {
	LastBalance     *big.Int        `json:"last_balance"`
	SessionUploaded map[string]bool `json:"session_uploaded"`
}

func (c *nodesetUploadDepositDataContext) PrepareData(data *swapi.NodesetUploadDepositDataData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider

	// Load persisted state
	if err := c.loadState(sp); err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error loading state: %w", err)
	}

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

	// Initialize the session map if not already initialized
	if c.sessionUploaded == nil {
		c.sessionUploaded = make(map[beacon.ValidatorPubkey]bool)
	}

	// Get the list of registered validators
	registeredPubkeyMap := map[beacon.ValidatorPubkey]bool{}
	pubkeyStatusResponse, err := nc.GetRegisteredValidators(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators: %w", err)
	}
	for _, pubkeyStatus := range pubkeyStatusResponse {
		registeredPubkeyMap[pubkeyStatus.Pubkey] = true
	}

	// Get the list of this node's validator keys
	keys, err := w.GetAllPrivateKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting private validator keys: %w", err)
	}
	data.TotalCount = uint64(len(keys))

	// Find the ones that haven't been uploaded yet
	unregisteredKeys := []*eth2types.BLSPrivateKey{}
	newPubkeys := []beacon.ValidatorPubkey{}
	for _, key := range keys {
		pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
		if !registeredPubkeyMap[pubkey] && !c.sessionUploaded[pubkey] {
			unregisteredKeys = append(unregisteredKeys, key)
			newPubkeys = append(newPubkeys, pubkey)
		}
	}
	data.UnregisteredPubkeys = newPubkeys

	if len(unregisteredKeys) == 0 {
		return types.ResponseStatus_Success, nil
	}

	// Make sure validator has enough funds to pay for the deposit
	balance, err := ec.BalanceAt(ctx, opts.From, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = balance

	// Reset sessionUploaded if balance changes
	if c.lastBalance == nil || balance.Cmp(c.lastBalance) != 0 {
		c.sessionUploaded = make(map[beacon.ValidatorPubkey]bool)
		c.lastBalance = balance
	}

	// Calculate the total deposit cost per validator
	validatorDepositCost := eth.EthToWei(0.01)
	totalCost := big.NewInt(int64(len(unregisteredKeys)))
	totalCost.Mul(totalCost, validatorDepositCost)
	data.RequiredBalance = totalCost

	// Calculate sufficient balance based on the initial number of unregistered keys
	data.SufficientBalance = (totalCost.Cmp(balance) <= 0)

	// Remove excess keys if insufficient balance
	if !data.SufficientBalance {
		for len(unregisteredKeys) > 0 && totalCost.Cmp(balance) > 0 {
			unregisteredKeys = unregisteredKeys[:len(unregisteredKeys)-1]
			newPubkeys = newPubkeys[:len(newPubkeys)-1]
			totalCost = totalCost.Sub(totalCost, validatorDepositCost)
		}
		data.UnregisteredPubkeys = newPubkeys
		data.RequiredBalance = totalCost
	}

	// Get the deposit data for the remaining pubkeys
	if len(unregisteredKeys) == 0 {
		return types.ResponseStatus_Success, nil
	}

	depositData, err := ddMgr.GenerateDepositData(unregisteredKeys)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error generating deposit data: %w", err)
	}

	// Serialize it
	bytes, err := json.Marshal(depositData)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error serializing deposit data: %w", err)
	}

	// Submit the upload
	response, err := nc.UploadDepositData(ctx, bytes)
	if err != nil {
		return types.ResponseStatus_Error, err
	}
	data.ServerResponse = response

	// Track newly uploaded keys in this session
	for _, pubkey := range newPubkeys {
		c.sessionUploaded[pubkey] = true
	}

	// Save the updated state
	if err := c.saveState(sp); err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error saving state: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func loadUploadState(sp *swcommon.StakewiseServiceProvider) (*uploadState, error) {
	filePathString := filepath.Join(sp.GetModuleDir(), stateFilePath)
	file, err := os.Open(filePathString)
	if err != nil {
		if os.IsNotExist(err) {
			return &uploadState{
				LastBalance:     big.NewInt(0),
				SessionUploaded: make(map[string]bool),
			}, nil // Return default if the file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	state := &uploadState{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Save the upload state to the file
func saveUploadState(sp *swcommon.StakewiseServiceProvider, state *uploadState) error {
	filePathString := filepath.Join(sp.GetModuleDir(), stateFilePath)
	file, err := os.Create(filePathString)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(state)
}

func (c *nodesetUploadDepositDataContext) loadState(sp *swcommon.StakewiseServiceProvider) error {
	state, err := loadUploadState(sp)
	if err != nil {
		return err
	}

	c.lastBalance = new(big.Int).Set(state.LastBalance)

	c.sessionUploaded = make(map[beacon.ValidatorPubkey]bool)
	for keyStr, uploaded := range state.SessionUploaded {
		pubkey := beacon.ValidatorPubkey([]byte(keyStr))
		if err != nil {
			return err
		}
		c.sessionUploaded[pubkey] = uploaded
	}
	return nil
}

func (c *nodesetUploadDepositDataContext) saveState(sp *swcommon.StakewiseServiceProvider) error {
	state := &uploadState{
		LastBalance:     new(big.Int).Set(c.lastBalance),
		SessionUploaded: make(map[string]bool),
	}
	for pubkey, uploaded := range c.sessionUploaded {
		fmt.Printf("pubkey: %s\n", pubkey.HexWithPrefix())
		state.SessionUploaded[pubkey.HexWithPrefix()] = uploaded
	}
	return saveUploadState(sp, state)
}
