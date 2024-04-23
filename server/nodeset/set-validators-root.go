package swnodeset

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type nodesetSetValidatorsRootContextFactory struct {
	handler *NodesetHandler
}

func (f *nodesetSetValidatorsRootContextFactory) Create(args url.Values) (*nodesetSetValidatorsRootContext, error) {
	c := &nodesetSetValidatorsRootContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("root", args, input.ValidateHash, &c.root),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodesetSetValidatorsRootContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetSetValidatorsRootContext, types.TxInfoData](
		router, "set-validators-root", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodesetSetValidatorsRootContext struct {
	handler *NodesetHandler
	root    common.Hash
}

func (c *nodesetSetValidatorsRootContext) PrepareData(data *types.TxInfoData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	res := sp.GetResources()
	txMgr := sp.GetTransactionManager()

	err := sp.RequireStakewiseWalletReady(walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	if res.Vault == nil {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("no Stakewise Vault address has been set yet")
	}

	vault, err := swcontracts.NewStakewiseVault(*res.Vault, ec, txMgr)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Stakewise Vault binding: %w", err)
	}

	data.TxInfo, err = vault.SetDepositDataRoot(c.root, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating SetDepositDataRoot TX: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
