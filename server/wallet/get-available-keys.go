package swwallet

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/wallet"
)

const (
	validatorDepositCost float64 = 0.01
)

// ===============
// === Factory ===
// ===============

type walletGetAvailableKeysContextFactory struct {
	handler *WalletHandler
}

func (f *walletGetAvailableKeysContextFactory) Create(args url.Values) (*walletGetAvailableKeysContext, error) {
	c := &walletGetAvailableKeysContext{
		handler: f.handler,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *walletGetAvailableKeysContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*walletGetAvailableKeysContext, api.WalletGetAvailableKeysData](
		router, "get-available-keys", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletGetAvailableKeysContext struct {
	handler *WalletHandler
}

func (c *walletGetAvailableKeysContext) PrepareData(data *api.WalletGetAvailableKeysData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	hd := sp.GetHyperdriveClient()
	keyMgr := sp.GetAvailableKeyManager()
	qMgr := sp.GetQueryManager()
	ec := sp.GetEthClient()
	ctx := c.handler.ctx
	nodeAddress := walletStatus.Address.NodeAddress
	res := sp.GetResources()

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}
	err = sp.RequireEthClientSynced(ctx)
	if err != nil {
		if errors.Is(err, services.ErrExecutionClientNotSynced) {
			return types.ResponseStatus_ClientsNotSynced, err
		}
		return types.ResponseStatus_Error, err
	}
	err = sp.RequireBeaconClientSynced(ctx)
	if err != nil {
		if errors.Is(err, services.ErrBeaconNodeNotSynced) {
			return types.ResponseStatus_ClientsNotSynced, err
		}
		return types.ResponseStatus_Error, err
	}

	// Fetch status from NodeSet
	response, err := hd.NodeSet_StakeWise.GetRegisteredValidators(res.DeploymentName, res.Vault)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators from Nodeset: %w", err)
	}
	if response.Data.NotRegistered {
		data.UnregisteredNode = true
		return types.ResponseStatus_Success, nil
	}

	// Get the current Beacon deposit root
	var depositRoot common.Hash
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		sp.GetBeaconDepositContract().GetDepositRoot(mc, &depositRoot)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest Beacon deposit root: %w", err)
	}

	// Get the list of keys ready for depositing
	keys, err := keyMgr.GetAvailableKeys(ctx, depositRoot, true)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting available keys: %w", err)
	}
	data.AvailablePubkeys = keys

	// Get the wallet's ETH balance
	balance, err := ec.BalanceAt(ctx, nodeAddress, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = eth.WeiToEth(balance)

	// Subtract the cost of the pending keys
	data.EthPerKey = validatorDepositCost
	costPerKeyBig := eth.EthToWei(validatorDepositCost)
	pendingCountBig := big.NewInt(int64(len(keys)))
	pendingCost := new(big.Int).Mul(costPerKeyBig, pendingCountBig)
	remainingBalance := new(big.Int).Sub(balance, pendingCost)
	requiredBalance := new(big.Int).Abs(remainingBalance)

	data.SufficientBalance = (remainingBalance.Cmp(common.Big0) >= 0)
	if !data.SufficientBalance {
		data.RemainingEthRequired = eth.WeiToEth(requiredBalance)
	}

	return types.ResponseStatus_Success, nil
}
