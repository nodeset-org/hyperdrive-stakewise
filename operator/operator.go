// This file is a port of some of the functionality from the StakeWise v3 Operator application
// https://github.com/stakewise/v3-operator

package operator

import (
	"fmt"
	"math/big"

	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	batchquery "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	// The hardcoded deposit amount for StakeWise validators, in ETH
	DepositAmountEth float64 = 32

	// The minimum number of validators to create within a single loop
	MinValidatorsToCreate uint64 = 1
)

var (
	// The hardcoded deposit amount for StakeWise validators, in wei
	DepositAmount *big.Int = eth.EthToWei(DepositAmountEth)
)

type Operator struct {
	sp           swcommon.IStakeWiseServiceProvider
	vault        *swcontracts.IEthVault
	oraclesCache *oraclesCache
}

func NewOperator(sp swcommon.IStakeWiseServiceProvider) (*Operator, error) {
	vaultAddress := sp.GetResources().Vault
	ec := sp.GetEthClient()
	txMgr := sp.GetTransactionManager()
	vault, err := swcontracts.NewIEthVault(vaultAddress, ec, txMgr)
	if err != nil {
		return nil, fmt.Errorf("error creating Stakewise vault contract instance: %w", err)
	}

	return &Operator{
		sp:    sp,
		vault: vault,
	}, nil
}

// https://github.com/stakewise/v3-operator/blob/eb02dd26d2337d39072f924386a3e97a16852a16/src/validators/tasks.py#L87
func (o *Operator) ProcessValidators() error {
	logger := o.sp.GetApiLogger()
	validatorsCount, err := o.getValidatorsCountFromVaultAssets()
	if err != nil {
		return fmt.Errorf("error getting validators count from vault assets: %w", err)
	}

	// Not enough ETH in the vault to create a validator
	if validatorsCount == 0 {
		return nil
	}

	// Make sure there's enough ETH in the vault to cover the minimum validator count
	if validatorsCount < MinValidatorsToCreate {
		logger.Debug(
			"Not enough ETH in the vault to create the minimum of validators",
			"min", MinValidatorsToCreate,
			"available", validatorsCount,
		)
		return nil
	}

	// Get the latest config
	//protocolConfig := o.getProtocolConfig()

	return nil
}

// https://github.com/stakewise/v3-operator/blob/eb02dd26d2337d39072f924386a3e97a16852a16/src/validators/tasks.py#L256
func (o *Operator) getValidatorsCountFromVaultAssets() (uint64, error) {
	// Get the vault balance
	var vaultBalance *big.Int
	qMgr := o.sp.GetQueryManager()
	err := qMgr.Query(func(mc *batchquery.MultiCaller) error {
		o.vault.WithdrawableAssets(mc, &vaultBalance)
		return nil
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("error querying vault balance: %w", err)
	}

	// Calculate the number of validators
	count := new(big.Int).Div(vaultBalance, DepositAmount)
	return count.Uint64(), nil
}

// https://github.com/stakewise/v3-operator/blob/eb02dd26d2337d39072f924386a3e97a16852a16/src/common/execution.py#L41
/*
func (o *Operator) getProtocolConfig() (*swcommon.ProtocolConfig, error) {
	cache, err := o.updateOraclesCache()
	if err != nil {
		return nil, fmt.Errorf("error updating stakewise oracles cache: %w", err)
	}

	pc, err := o.buildProtocolConfig(cache.config, cache.rewardsThreshold, cache.validatorsThreshold)
	if err != nil {
		return nil, fmt.Errorf("error building protocol config: %w", err)
	}
	return pc, nil
}
*/

/*
func (o *Operator) updateOraclesCache() error {
	logger := o.sp.GetApiLogger()

	// Get the start block
	var fromBlock *big.Int
	if o.oraclesCache == nil {
		fromBlock = o.sp.GetResources().KeeperGenesisBlock
	} else {
		fromBlock = new(big.Int).Add(o.oraclesCache.checkpointBlock, common.Big1)
	}

	// Get the end block
	ec := o.sp.GetEthClient()
	toBlockUint, err := ec.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current block number: %w", err)
	}
	toBlock := new(big.Int).SetUint64(toBlockUint)
	if fromBlock.Cmp(toBlock) > 0 {
		return nil
	}

	logger.Debug("Updating oracles cache", "from", fromBlock, "to", toBlock)

}
*/
