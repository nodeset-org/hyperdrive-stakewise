package swservice

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type serviceGetNetworkSettingsContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceGetNetworkSettingsContextFactory) Create(args url.Values) (*serviceGetNetworkSettingsContext, error) {
	c := &serviceGetNetworkSettingsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceGetNetworkSettingsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceGetNetworkSettingsContext, swapi.ServiceGetNetworkSettingsData](
		router, "get-network-settings", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceGetNetworkSettingsContext struct {
	handler *ServiceHandler
}

func (c *serviceGetNetworkSettingsContext) PrepareData(data *swapi.ServiceGetNetworkSettingsData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	hdCfg := sp.GetHyperdriveConfig()
	swCfg := sp.GetConfig()
	settingsList := swCfg.GetNetworkSettings()
	network := hdCfg.Network.Value
	for _, settings := range settingsList {
		if settings.Key == network {
			data.Settings = settings
			return types.ResponseStatus_Success, nil
		}
	}
	return types.ResponseStatus_Error, fmt.Errorf("hyperdrive has network [%s] selected but stakewise has no settings for it", network)
}
