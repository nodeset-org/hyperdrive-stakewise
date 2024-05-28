package swnodeset

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type nodesetBlahContextFactory struct {
	handler *NodesetHandler
}

func (f *nodesetBlahContextFactory) Create(args url.Values) (*nodesetBlahContext, error) {
	c := &nodesetBlahContext{
		handler: f.handler,
	}
	inputErrs := []error{}

	// TODO: Validation
	c.nodeAddress = args.Get("nodeAddress")
	c.email = args.Get("email")
	c.signature = args.Get("signature")

	return c, errors.Join(inputErrs...)
}

func (f *nodesetBlahContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetBlahContext, types.TxInfoData](
		router, "blah", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodesetBlahContext struct {
	handler     *NodesetHandler
	email       string
	nodeAddress string
	signature   string
}

func (c *nodesetBlahContext) PrepareData(data *types.TxInfoData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	nc := sp.GetNodesetClient()

	ctx := c.handler.ctx

	resp, err := nc.SubmitAuthorizedAddress(ctx, c.email, c.nodeAddress, c.signature)
	if err != nil {
		fmt.Printf("Error submitting authorized address: %v\n", err)
		return types.ResponseStatus_Error, err
	}
	fmt.Printf("Response: %v\n", resp)

	return types.ResponseStatus_Success, nil
}
