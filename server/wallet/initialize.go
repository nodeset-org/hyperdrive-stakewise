package swwallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type walletInitializeContextFactory struct {
	handler *WalletHandler
}

func (f *walletInitializeContextFactory) Create(args url.Values) (*walletInitializeContext, error) {
	c := &walletInitializeContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletInitializeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletInitializeContext, api.WalletInitializeData](
		router, "initialize", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletInitializeContext struct {
	handler *WalletHandler
}

func (c *walletInitializeContext) PrepareData(data *api.WalletInitializeData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady(walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Get the Geth keystore in JSON format
	ethkeyResponse, err := client.Wallet.ExportEthKey()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting geth-style keystore: %w", err)
	}
	ethKey := ethkeyResponse.Data.EthKeyJson
	password := ethkeyResponse.Data.Password

	// Write the Stakewise wallet files to disk
	err = w.SaveStakewiseWallet(ethKey, password)
	if err != nil {
		return types.ResponseStatus_Error, err
	}

	data.AccountAddress = walletStatus.Wallet.WalletAddress
	return types.ResponseStatus_Success, nil
}
