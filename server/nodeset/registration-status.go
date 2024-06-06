package swnodeset

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type nodesetRegistrationStatusContextFactory struct {
	handler *NodesetHandler
}

func (f *nodesetRegistrationStatusContextFactory) Create(args url.Values) (*nodesetRegistrationStatusContext, error) {
	c := &nodesetRegistrationStatusContext{
		handler: f.handler,
	}

	return c, nil
}

func (f *nodesetRegistrationStatusContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetRegistrationStatusContext, swapi.NodeSetRegistrationStatusData](
		router, "registration-status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodesetRegistrationStatusContext struct {
	handler *NodesetHandler
}

func (c *nodesetRegistrationStatusContext) PrepareData(data *swapi.NodeSetRegistrationStatusData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ctx := c.handler.ctx

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Register the node
	ns := sp.GetNodesetClient()
	data.Status, err = ns.GetNodeRegistrationStatus(ctx)
	if err != nil {
		data.ErrorMessage = err.Error()
	}

	return types.ResponseStatus_Success, nil
}
