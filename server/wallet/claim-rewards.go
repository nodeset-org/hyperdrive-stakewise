package swwallet

import (
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
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

	if res.SplitMain == nil {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("no SplitMain contract has been set yet")
	}
	if res.Vault == nil {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("no Stakewise Vault address has been set yet")
	}

	logger.Debug("Preparing data for claim reward")
	splitMainContract, err := swcontracts.NewSplitMain(*res.SplitMain, ec, txMgr)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Stakewise Vault binding: %w", err)
	}

	// Get the claimable rewards
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		splitMainContract.GetErc20Balance(mc, &data.WithdrawableToken, nodeAddress, *res.Vault)
		splitMainContract.GetEthBalance(mc, &data.WithdrawableEth, nodeAddress)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error querying claimable rewards: %w", err)
	}

	data.TxInfo, err = splitMainContract.Withdraw(*res.Vault, *res.ClaimEthAmount, res.ClaimTokenList, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Withdraw TX: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
