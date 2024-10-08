package swwallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type walletGenerateKeysContextFactory struct {
	handler *WalletHandler
}

func (f *walletGenerateKeysContextFactory) Create(args url.Values) (*walletGenerateKeysContext, error) {
	c := &walletGenerateKeysContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("count", args, input.ValidateUint, &c.count),
		server.ValidateArg("restart-vc", args, input.ValidateBool, &c.restartVc),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletGenerateKeysContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*walletGenerateKeysContext, api.WalletGenerateKeysData](
		router, "generate-keys", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletGenerateKeysContext struct {
	handler   *WalletHandler
	count     uint64
	restartVc bool
}

func (c *walletGenerateKeysContext) PrepareData(data *api.WalletGenerateKeysData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	wallet := sp.GetWallet()
	ctx := c.handler.ctx

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Generate and save the keys
	pubkeys := make([]beacon.ValidatorPubkey, c.count)
	for i := 0; i < int(c.count); i++ {
		key, err := wallet.GenerateNewValidatorKey()
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error generating validator key: %w", err)
		}
		pubkeys[i] = beacon.ValidatorPubkey(key.PublicKey().Marshal())
	}
	data.Pubkeys = pubkeys

	// Restart the VC
	if c.restartVc {
		_, err = client.Service.RestartContainer(string(swconfig.ContainerID_StakewiseValidator))
		if err != nil {
			return types.ResponseStatus_Error, err
		}
	}
	return types.ResponseStatus_Success, nil
}
