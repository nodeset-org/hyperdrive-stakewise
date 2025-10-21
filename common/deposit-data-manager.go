package swcommon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	// Stakewise validators deposit a full 32 ETH
	StakewiseDepositAmount uint64 = 32e9
)

// DEPRECATED: This was only necessary for StakeWise v1 support and now just creates a blank file.
// Once StakeWise no longer needs the file at all, this can be removed.
type DepositDataManager struct {
	sp IStakeWiseServiceProvider
}

// Creates a new manager
func NewDepositDataManager(sp IStakeWiseServiceProvider) (*DepositDataManager, error) {
	ddMgr := &DepositDataManager{
		sp: sp,
	}

	// Ensure the deposit data file is empty
	// Replace with deleteDepositDataFile(sp) when v4 is introduced
	err := emptyDepositDataFile(sp)
	if err != nil {
		return nil, fmt.Errorf("error initializing deposit data manager: %w", err)
	}
	return ddMgr, nil
}

// Generates deposit data for the provided keys
func GenerateDepositData(logger *slog.Logger, resources *swconfig.MergedResources, keys []*eth2types.BLSPrivateKey) ([]beacon.ExtendedDepositData, error) {
	// Stakewise uses the same withdrawal creds for each validator
	withdrawalCreds := validator.GetWithdrawalCredsFromAddress(resources.Vault)

	// Create the new aggregated deposit data for all generated keys
	dataList := make([]beacon.ExtendedDepositData, len(keys))
	for i, key := range keys {
		depositData, err := validator.GetDepositData(logger, key, withdrawalCreds, resources.GenesisForkVersion, StakewiseDepositAmount, resources.EthNetworkName)
		if err != nil {
			pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
			return nil, fmt.Errorf("error getting deposit data for key %s: %w", pubkey.HexWithPrefix(), err)
		}
		dataList[i] = depositData
	}
	return dataList, nil
}

// Calculates the deposit domain for Beacon deposits
func GetGenesisDepositDomain(genesisForkVersion []byte) ([]byte, error) {
	return signing.ComputeDomain(eth2types.DomainDeposit, genesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
}

// Validates the signature for deposit data
func ValidateDepositInfo(logger *slog.Logger, depositDomain []byte, depositAmount uint64, pubkey []byte, withdrawalCredentials []byte, signature []byte) error {
	if logger != nil {
		logger.Debug("Validating deposit data",
			slog.String("domain", hex.EncodeToString(depositDomain)),
		)
	}
	depositData := &ethpb.Deposit_Data{
		Amount:                depositAmount,
		PublicKey:             pubkey,
		WithdrawalCredentials: withdrawalCredentials,
		Signature:             signature,
	}
	return prdeposit.VerifyDepositSignature(depositData, depositDomain)
}

// Empties out the v3 deposit data file
func emptyDepositDataFile(sp IStakeWiseServiceProvider) error {
	depositDataPath := filepath.Join(sp.GetModuleDir(), swconfig.DepositDataFile)
	bytes := []byte("{}")
	err := os.WriteFile(depositDataPath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("error emptying deposit data file: %w", err)
	}
	return nil
}

// Deletes the old deposit data file used in v3
//
//nolint:unused
func deleteDepositDataFile(sp IStakeWiseServiceProvider) error {
	depositDataPath := filepath.Join(sp.GetModuleDir(), swconfig.DepositDataFile)
	err := os.Remove(depositDataPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error deleting deposit data file: %w", err)
	}
	return nil
}
