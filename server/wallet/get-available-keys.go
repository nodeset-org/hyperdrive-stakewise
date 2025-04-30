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
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
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
	inputErrs := []error{
		server.ValidateOptionalArg("lookback", args, input.ValidateBool, &c.doLookback, nil),
	}
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
	handler    *WalletHandler
	doLookback bool
}

func (c *walletGetAvailableKeysContext) PrepareData(data *api.WalletGetAvailableKeysData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	keyMgr := sp.GetAvailableKeyManager()
	qMgr := sp.GetQueryManager()
	ec := sp.GetEthClient()
	ctx := c.handler.ctx
	logger := c.handler.logger.Logger
	nodeAddress := walletStatus.Address.NodeAddress

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

	// Get the current Beacon deposit root
	var depositRoot common.Hash
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		sp.GetBeaconDepositContract().GetDepositRoot(mc, &depositRoot)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest Beacon deposit root: %w", err)
	}

	// Get the current block number
	currentBlock, err := sp.GetEthClient().BlockNumber(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting current block number: %w", err)
	}

	// Get the list of keys ready for depositing
	scanOpts := swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: false,
	}
	if keyMgr.RequiresLookbackScan(currentBlock) || c.doLookback {
		scanOpts.DoLookbackScan = true
	}
	goodKeys, badKeys, err := keyMgr.GetAvailableKeys(ctx, logger, depositRoot, currentBlock, scanOpts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting available keys: %w", err)
	}
	for _, key := range goodKeys {
		data.AvailablePubkeys = append(data.AvailablePubkeys, key.PublicKey)
	}
	for key, reason := range badKeys {
		switch reason {
		case swcommon.IneligibleReason_NoPrivateKey:
			data.KeysMissingPrivateKey = append(data.KeysMissingPrivateKey, key.PublicKey)
		case swcommon.IneligibleReason_LookbackScanRequired:
			data.KeysRequiringLookbackScan = append(data.KeysRequiringLookbackScan, key.PublicKey)
		case swcommon.IneligibleReason_OnBeacon:
			data.KeysAlreadyOnBeacon = append(data.KeysAlreadyOnBeacon, key.PublicKey)
		case swcommon.IneligibleReason_HasDepositEvent:
			data.KeysWithDepositEvents = append(data.KeysWithDepositEvents, key.PublicKey)
		case swcommon.IneligibleReason_AlreadyUsedDepositRoot:
			data.KeysUsedWithDepositRoot = append(data.KeysUsedWithDepositRoot, key.PublicKey)
		default:
			logger.Warn(
				"Key is not available for an unknown reason",
				"key", key.PublicKey.HexWithPrefix(),
				"reason", reason,
			)
		}
	}

	// Get the wallet's ETH balance
	balance, err := ec.BalanceAt(ctx, nodeAddress, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting balance: %w", err)
	}
	data.Balance = eth.WeiToEth(balance)

	// Subtract the cost of the pending keys
	data.EthPerKey = validatorDepositCost
	costPerKeyBig := eth.EthToWei(validatorDepositCost)
	pendingCountBig := big.NewInt(int64(len(goodKeys)))
	pendingCost := new(big.Int).Mul(costPerKeyBig, pendingCountBig)
	remainingBalance := new(big.Int).Sub(balance, pendingCost)
	requiredBalance := new(big.Int).Abs(remainingBalance)

	data.SufficientBalance = (remainingBalance.Cmp(common.Big0) >= 0)
	if !data.SufficientBalance {
		data.RemainingEthRequired = eth.WeiToEth(requiredBalance)
	}

	return types.ResponseStatus_Success, nil
}
