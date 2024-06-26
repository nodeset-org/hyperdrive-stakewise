package swstatus

import (
	"errors"
	"fmt"
	"net/url"

	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"

	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/wallet"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
)

// ===============
// === Factory ===
// ===============

type statusGetValidatorsStatusesContextFactory struct {
	handler *StatusHandler
}

func (f *statusGetValidatorsStatusesContextFactory) Create(args url.Values) (*statusGetValidatorsStatusesContext, error) {
	c := &statusGetValidatorsStatusesContext{
		handler: f.handler,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *statusGetValidatorsStatusesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*statusGetValidatorsStatusesContext, swapi.ValidatorStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type statusGetValidatorsStatusesContext struct {
	handler *StatusHandler
}

func (c *statusGetValidatorsStatusesContext) PrepareData(data *swapi.ValidatorStatusData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	bc := sp.GetBeaconClient()
	w := sp.GetWallet()
	nc := sp.GetNodesetClient()
	ctx := c.handler.ctx

	logger := c.handler.logger

	// Requirements
	err := sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}
	err = sp.RequireBeaconClientSynced(ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	logger.Debug("Getting validator statuses from NodeSet")
	nodesetStatusResponse, err := nc.GetRegisteredValidators(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting nodeset statuses: %w", err)
	}

	logger.Debug("Getting private keys")
	privateKeys, err := w.GetAllPrivateKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting private keys: %w", err)
	}

	logger.Debug("Deriving public keys")
	publicKeys, err := w.DerivePubKeys(privateKeys)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting public keys: %w", err)
	}

	logger.Debug("Getting validator statuses from Beacon client")
	beaconStatusResponse, err := bc.GetValidatorStatuses(ctx, publicKeys, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting validator statuses: %w", err)
	}

	registeredPubkeysStatusMapping := make(map[beacon.ValidatorPubkey]string)
	for _, pubkeyStatus := range nodesetStatusResponse {
		registeredPubkeysStatusMapping[pubkeyStatus.Pubkey] = pubkeyStatus.Status
	}

	// Get status info for each pubkey
	data.States = make([]swapi.ValidatorStateInfo, len(publicKeys))
	for i, pubkey := range publicKeys {
		state := &data.States[i]
		state.Pubkey = pubkey

		// Beacon status
		status, exists := beaconStatusResponse[pubkey]
		if exists {
			state.BeaconStatus = status.Status
			state.Index = status.Index
		} else {
			state.BeaconStatus = ""
		}

		// NodeSet status
		state.NodesetStatus = swcommon.GetNodesetStatus(pubkey, registeredPubkeysStatusMapping)
	}

	return types.ResponseStatus_Success, nil
}
