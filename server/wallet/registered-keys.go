package swwallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type walletRegisteredKeysContextFactory struct {
	handler *WalletHandler
}

func (f *walletRegisteredKeysContextFactory) Create(args url.Values) (*walletRegisteredKeysContext, error) {
	c := &walletRegisteredKeysContext{
		handler: f.handler,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *walletRegisteredKeysContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*walletRegisteredKeysContext, api.WalletRegisteredKeysData](
		router, "registered-keys", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRegisteredKeysContext struct {
	handler   *WalletHandler
	count     uint64
	restartVc bool
}

func (c *walletRegisteredKeysContext) PrepareData(data *api.WalletRegisteredKeysData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	ctx := c.handler.ctx
	res := sp.GetResources()

	// Requirements
	err := sp.RequireRegisteredWithNodeSet(ctx)
	if err != nil {
		data.NotRegisteredWithNodeSet = true
		return types.ResponseStatus_Success, err
	}

	// Get the list of vaults for this deployment
	response, err := client.NodeSet_StakeWise.GetVaults(res.DeploymentName)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("failed to get vaults: %w", err)
	}
	if response.Data.NotRegistered {
		data.NotRegisteredWithNodeSet = true
		return types.ResponseStatus_Success, nil
	}
	if response.Data.InvalidPermissions {
		data.InvalidPermissions = true
		return types.ResponseStatus_Success, nil
	}
	if response.Data.Vaults == nil {
		data.Vaults = []api.VaultInfo{}
		return types.ResponseStatus_Success, nil
	}

	// For each vault, get the validator keys
	for _, vault := range response.Data.Vaults {
		vaultInfo := api.VaultInfo{
			Name:       vault.Name,
			Address:    vault.Address,
			Validators: []beacon.ValidatorPubkey{},
		}
		vaultResponse, err := client.NodeSet_StakeWise.GetRegisteredValidators(res.DeploymentName, vault.Address)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting registered validators for vault [%s]: %w", vault.Address.Hex(), err)
		}
		if vaultResponse.Data.NotRegistered {
			data.NotRegisteredWithNodeSet = true
			return types.ResponseStatus_Success, nil
		}
		if vaultResponse.Data.InvalidPermissions {
			vaultInfo.HasPermission = false
		} else {
			vaultInfo.HasPermission = true
			for _, validator := range vaultResponse.Data.Validators {
				vaultInfo.Validators = append(vaultInfo.Validators, validator.Pubkey)
			}
		}
		data.Vaults = append(data.Vaults, vaultInfo)
	}

	return types.ResponseStatus_Success, nil
}
