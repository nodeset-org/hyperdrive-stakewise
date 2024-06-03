package swwallet

import (
	"fmt"
	"math/big"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth/contracts"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type walletClaimRewardsContextFactory struct {
	handler *WalletHandler
}

func (f *walletClaimRewardsContextFactory) Create(args url.Values) (*walletClaimRewardsContext, error) {
	c := &walletClaimRewardsContext{
		handler: f.handler,
	}

	return c, nil
}

func (f *walletClaimRewardsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletClaimRewardsContext, swapi.WalletClaimRewardsData](
		router, "claim-rewards", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletClaimRewardsContext struct {
	handler *WalletHandler
	// address common.Address
}

// Return the transaction data
func (c *walletClaimRewardsContext) PrepareData(data *swapi.WalletClaimRewardsData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	logger := c.handler.logger
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	res := sp.GetResources()
	qMgr := sp.GetQueryManager()
	txMgr := sp.GetTransactionManager()
	nodeAddress := walletStatus.Address.NodeAddress

	// Requirements
	err := sp.RequireNodeAddress(walletStatus)
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}

	if res.SplitWarehouse == nil {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("no SplitWarehouse contract has been set yet")
	}
	if res.Vault == nil {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("no Stakewise Vault address has been set yet")
	}

	// Create bindings
	logger.Debug("Preparing data for claim reward")
	splitWarehouseContract, err := swcontracts.NewSplitWarehouse(*res.SplitWarehouse, ec, txMgr) // NOTE: need to parse the actual contract version once event support is added
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating SplitWarehouse binding: %w", err)
	}
	token, err := contracts.NewErc20Contract(*res.Vault, ec, qMgr, txMgr, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Stakewise token binding: %w", err)
	}
	data.TokenName = token.Name()
	data.TokenSymbol = token.Symbol()

	// Get the native token address
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		splitWarehouseContract.NativeToken(mc, &data.NativeToken)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error querying claimable rewards: %w", err)
	}

	// Get the claimable rewards
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		splitWarehouseContract.BalanceOf(mc, &data.WithdrawableToken, nodeAddress, *res.Vault)
		splitWarehouseContract.BalanceOf(mc, &data.WithdrawableEth, nodeAddress, data.NativeToken)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error querying claimable rewards: %w", err)
	}

	tokensToWithdraw := []common.Address{}
	amountsToWithdraw := []*big.Int{}

	if data.WithdrawableEth.Cmp(common.Big0) > 0 {
		// Only withdraw ETH if there is a balance
		tokensToWithdraw = append(tokensToWithdraw, data.NativeToken)
		amountsToWithdraw = append(amountsToWithdraw, data.WithdrawableEth)
	}

	if data.WithdrawableToken.Cmp(common.Big0) > 0 {
		// Only withdraw tokens if there is a balance
		tokensToWithdraw = append(tokensToWithdraw, *res.Vault)
		amountsToWithdraw = append(amountsToWithdraw, data.WithdrawableToken)
	}

	data.TxInfo, err = splitWarehouseContract.Withdraw(nodeAddress, amountsToWithdraw, tokensToWithdraw, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Withdraw TX: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
