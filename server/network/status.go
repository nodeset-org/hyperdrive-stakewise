package swnetwork

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	api "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type networkStatusContextFactory struct {
	handler *NetworkHandler
}

func (f *networkStatusContextFactory) Create(args url.Values) (*networkStatusContext, error) {
	c := &networkStatusContext{
		handler: f.handler,
	}
	inputErrs := []error{}
	return c, errors.Join(inputErrs...)
}

func (f *networkStatusContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*networkStatusContext, api.NetworkStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkStatusContext struct {
	handler *NetworkHandler
}

func (c *networkStatusContext) PrepareData(data *api.NetworkStatusData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	client := sp.GetHyperdriveClient()
	ctx := c.handler.ctx
	res := sp.GetResources()
	ec := sp.GetEthClient()
	qMgr := sp.GetQueryManager()
	txMgr := sp.GetTransactionManager()

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
		data.Vaults = []*api.NetworkVaultInfo{}
		return types.ResponseStatus_Success, nil
	}

	// For each vault, get the validator keys
	vaultContracts := make(map[common.Address]*swcontracts.IEthVault)
	for _, vault := range response.Data.Vaults {
		vaultInfo := &api.NetworkVaultInfo{
			Name:    vault.Name,
			Address: vault.Address,
		}
		metaResponse, err := client.NodeSet_StakeWise.GetValidatorsInfo(res.DeploymentName, vault.Address)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting vault [%s] info: %w", vault.Address.Hex(), err)
		}
		if metaResponse.Data.NotRegistered {
			data.NotRegisteredWithNodeSet = true
			return types.ResponseStatus_Success, nil
		}
		vaultInfo.MaxValidators = metaResponse.Data.MaxValidators
		vaultInfo.RegisteredValidators = metaResponse.Data.RegisteredValidators
		vaultInfo.AvailableValidators = metaResponse.Data.AvailableValidators

		// Create the vault contract
		vaultContract, err := swcontracts.NewIEthVault(vault.Address, ec, txMgr)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating binding for vault [%s]: %w", vault.Address.Hex(), err)
		}
		vaultContracts[vault.Address] = vaultContract
		data.Vaults = append(data.Vaults, vaultInfo)
	}

	// Get the balances
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		for _, vault := range data.Vaults {
			contract := vaultContracts[vault.Address]
			contract.WithdrawableAssets(mc, &vault.Balance)
		}
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting vault balances: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
