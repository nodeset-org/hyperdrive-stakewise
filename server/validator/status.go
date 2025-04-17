package swvalidator

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type validatorStatusContextFactory struct {
	handler *ValidatorHandler
}

func (f *validatorStatusContextFactory) Create(args url.Values) (*validatorStatusContext, error) {
	c := &validatorStatusContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateOptionalArg("vault", args, input.ValidateAddress, &c.vault, &c.hasVault),
	}
	return c, errors.Join(inputErrs...)
}

func (f *validatorStatusContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*validatorStatusContext, api.ValidatorStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type validatorStatusContext struct {
	handler  *ValidatorHandler
	vault    common.Address
	hasVault bool
}

func (c *validatorStatusContext) PrepareData(data *api.ValidatorStatusData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	ctx := c.handler.ctx
	res := sp.GetResources()
	bn := sp.GetBeaconClient()

	// Requirements
	err := sp.RequireRegisteredWithNodeSet(ctx)
	if err != nil {
		data.NotRegisteredWithNodeSet = true
		return types.ResponseStatus_Success, err
	}

	// Get the list of vaults for this deployment
	var vaults []v3stakewise.VaultInfo
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
		data.Vaults = []*api.VaultInfo{}
		return types.ResponseStatus_Success, nil
	}

	// Filter on vault address if provided
	if !c.hasVault {
		vaults = response.Data.Vaults
	} else {
		vaults = make([]v3stakewise.VaultInfo, 0)
		for _, vault := range response.Data.Vaults {
			if vault.Address == c.vault {
				vaults = append(vaults, vault)
				break
			}
		}
		if len(vaults) == 0 {
			return types.ResponseStatus_Error, fmt.Errorf("vault [%s] is not a StakeWise vault", c.vault.Hex())
		}
	}

	// For each vault, get the validator keys
	pubkeys := make([]beacon.ValidatorPubkey, 0)
	for _, vault := range vaults {
		vaultInfo := &api.VaultInfo{
			Name:       vault.Name,
			Address:    vault.Address,
			Validators: []*api.ValidatorInfo{},
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
				vaultInfo.Validators = append(vaultInfo.Validators, &api.ValidatorInfo{
					Pubkey: validator.Pubkey,
				})
				pubkeys = append(pubkeys, validator.Pubkey)
			}
		}
		data.Vaults = append(data.Vaults, vaultInfo)
	}

	// Get the status of the validators on Beacon
	statuses, err := bn.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting validator statuses: %w", err)
	}
	for _, vault := range data.Vaults {
		for _, validator := range vault.Validators {
			status, exists := statuses[validator.Pubkey]
			if !exists {
				validator.HasBeaconIndex = false
				continue
			}
			validator.HasBeaconIndex = true
			validator.Index = status.Index
			validator.Balance = status.Balance
			validator.State = status.Status

		}
	}

	return types.ResponseStatus_Success, nil
}
