package swtasks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nodeset-org/hyperdrive-daemon/client"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"

	"github.com/ethereum/go-ethereum/common"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
)

// Update deposit data task
type UpdateDepositDataTask struct {
	logger *log.Logger
	ctx    context.Context
	sp     swcommon.IStakeWiseServiceProvider
	w      *swcommon.Wallet
	hd     *client.ApiClient
	ddMgr  *swcommon.DepositDataManager
	cfg    *swconfig.StakeWiseConfig
	res    *swconfig.MergedResources
}

// Create update deposit data task
func NewUpdateDepositDataTask(ctx context.Context, sp swcommon.IStakeWiseServiceProvider, logger *log.Logger) *UpdateDepositDataTask {
	return &UpdateDepositDataTask{
		logger: logger,
		ctx:    ctx,
		sp:     sp,
		w:      sp.GetWallet(),
		hd:     sp.GetHyperdriveClient(),
		ddMgr:  sp.GetDepositDataManager(),
		cfg:    sp.GetConfig(),
		res:    sp.GetResources(),
	}
}

// Update deposit data
func (t *UpdateDepositDataTask) Run() error {
	t.logger.Info("Checking version of NodeSet data on disk...")

	// Get the version on the server
	versionResponse, err := t.hd.NodeSet_StakeWise.GetDepositDataSetVersion(t.res.DeploymentName, t.res.Vault)
	if err != nil {
		return fmt.Errorf("error getting latest deposit data version: %w", err)
	}
	if versionResponse.Data.NotRegistered {
		t.logger.Warn("Node is not registered with the NodeSet server yet.")
		return nil
	}

	// Compare versions
	localVersion := t.w.GetLatestDepositDataVersion()
	remoteVersion := versionResponse.Data.Version
	if remoteVersion == localVersion {
		t.logger.Info("Local data is up to date", slog.Int("version", localVersion))
		return nil
	}

	// Get the new data
	t.logger.Info("Deposit data is out of date retrieving latest data...", slog.Int("localVersion", localVersion), slog.Int("remoteVersion", remoteVersion))
	setResponse, err := t.hd.NodeSet_StakeWise.GetDepositDataSet(t.res.DeploymentName, t.res.Vault)
	if err != nil {
		return fmt.Errorf("error getting latest deposit data: %w", err)
	}

	// Verify the merkle roots if enabled
	depositData := setResponse.Data.DepositData
	if t.cfg.VerifyDepositsRoot.Value {
		isMatch, err := t.verifyDepositsRoot(depositData)
		if err != nil {
			return err
		}
		if !isMatch {
			return nil
		}
	}

	// Save it
	err = t.ddMgr.UpdateDepositData(depositData)
	if err != nil {
		return fmt.Errorf("error saving deposit data: %w", err)
	}
	err = t.w.SetLatestDepositDataVersion(remoteVersion)
	if err != nil {
		return fmt.Errorf("error updating latest saved version number: %w", err)
	}

	// Restart the Stakewise op container
	t.logger.Info("Restarting Stakewise operator...")
	_, err = t.hd.Service.RestartContainer(string(swconfig.ContainerID_StakewiseOperator))
	if err != nil {
		return fmt.Errorf("error restarting %s container: %w", swconfig.ContainerID_StakewiseOperator, err)
	}

	t.logger.Info("Done! Your deposit data is now up to date.")
	return nil
}

// Verify the Merkle root from the deposits data matches what's on chain before saving
func (t *UpdateDepositDataTask) verifyDepositsRoot(depositData []beacon.ExtendedDepositData) (bool, error) {
	// Get services
	ec := t.sp.GetEthClient()
	res := t.sp.GetResources()
	txMgr := t.sp.GetTransactionManager()
	q := t.sp.GetQueryManager()
	ddMgr := t.sp.GetDepositDataManager()

	// Get the Merkle root from it
	localRoot, err := ddMgr.ComputeMerkleRoot(depositData)
	if err != nil {
		return false, fmt.Errorf("error computing Merkle root from deposit data: %w", err)
	}
	t.logger.Info("Computed Merkle root", slog.String("root", localRoot.Hex()))

	// Get the Merkle root from the vault
	vault, err := swcontracts.NewStakewiseVault(res.Vault, ec, txMgr)
	if err != nil {
		return false, fmt.Errorf("error creating Stakewise Vault binding: %w", err)
	}
	var contractRoot common.Hash
	err = q.Query(func(mc *batch.MultiCaller) error {
		vault.GetValidatorsRoot(mc, &contractRoot)
		return nil
	}, nil)
	if err != nil {
		return false, fmt.Errorf("error getting canonical deposit root from the Stakewise Vault: %w", err)
	}
	t.logger.Info("Retrieved contract's Merkle root", slog.String("root", contractRoot.Hex()))

	// Compare them
	if localRoot != contractRoot {
		t.logger.Warn("Locally computed deposits data root does not match the value stored on chain, refusing to save for safety!")
		return false, nil
	} else {
		t.logger.Info("Locally computed deposits data root matches the root stored on-chain, updating may proceed.")
	}
	return true, nil
}
