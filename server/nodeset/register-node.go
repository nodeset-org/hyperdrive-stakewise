package swnodeset

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type nodesetRegisterNodeContexttFactory struct {
	handler *NodesetHandler
}

func (f *nodesetRegisterNodeContexttFactory) Create(args url.Values) (*nodesetRegisterNodeContext, error) {
	c := &nodesetRegisterNodeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("email", args, &c.email),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodesetRegisterNodeContexttFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetRegisterNodeContext, types.SuccessData](
		router, "register-node", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodesetRegisterNodeContext struct {
	handler *NodesetHandler
	email   string
}

func (c *nodesetRegisterNodeContext) PrepareData(data *types.SuccessData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ctx := c.handler.ctx

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Register the node
	ns := sp.GetNodesetClient()
	err = ns.RegisterNode(ctx, c.email, walletStatus.Wallet.WalletAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error registering node: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
