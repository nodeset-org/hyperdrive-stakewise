package swcommon

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	// Stakewise validators deposit a full 32 ETH
	StakewiseDepositAmount uint64 = 32e9
)

// DepositDataManager manages the aggregated deposit data file that Stakewise uses
type DepositDataManager struct {
	sp IStakeWiseServiceProvider
}

// Creates a new manager
func NewDepositDataManager(sp IStakeWiseServiceProvider) (*DepositDataManager, error) {
	ddMgr := &DepositDataManager{
		sp: sp,
	}

	// Remove the legacy deposit data file if it exists
	legacyDepositDataPath := filepath.Join(sp.GetModuleDir(), swconfig.DepositDataFile)
	_ = os.Remove(legacyDepositDataPath)
	return ddMgr, nil
}

// Generates deposit data for the provided keys
func (m *DepositDataManager) GenerateDepositData(logger *slog.Logger, keys []*eth2types.BLSPrivateKey) ([]beacon.ExtendedDepositData, error) {
	resources := m.sp.GetResources()

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
