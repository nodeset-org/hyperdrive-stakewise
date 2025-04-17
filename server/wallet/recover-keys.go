package swwallet

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type walletRecoverKeysContextFactory struct {
	handler *WalletHandler
}

func (f *walletRecoverKeysContextFactory) Create(body api.WalletRecoverKeysBody) (*walletRecoverKeysContext, error) {
	c := &walletRecoverKeysContext{
		handler: f.handler,
		body:    body,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *walletRecoverKeysContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessPost[*walletRecoverKeysContext, api.WalletRecoverKeysBody, api.WalletRecoverKeysData](
		router, "recover-keys", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRecoverKeysContext struct {
	handler *WalletHandler
	body    api.WalletRecoverKeysBody
}

func (c *walletRecoverKeysContext) PrepareData(data *api.WalletRecoverKeysData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	wallet := sp.GetWallet()
	ctx := c.handler.ctx

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}
	err = sp.RequireRegisteredWithNodeSet(ctx)
	if err != nil {
		data.NotRegisteredWithNodeSet = true
		return types.ResponseStatus_Success, err
	}

	// Recover the keys
	keys, lastIndexSearched, err := wallet.RecoverValidatorKeys(c.body.Pubkeys, c.body.StartIndex, c.body.Count, c.body.SearchLimit)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error recovering validator keys: %w", err)
	}
	data.Keys = keys
	data.SearchEnd = lastIndexSearched

	// Restart the VC
	if c.body.RestartVc {
		_, err = client.Service.RestartContainer(string(swconfig.ContainerID_StakewiseValidator))
		if err != nil {
			return types.ResponseStatus_Error, err
		}
	}
	return types.ResponseStatus_Success, nil
}
