package swcommon

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-json"
	swcontracts "github.com/nodeset-org/hyperdrive-stakewise/common/contracts"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

const (
	// The size of the interval to use, in blocks, when scanning blocks for Oracle config events
	// Set to 1/10th of 1 week, assuming 12 seconds per block
	OracleConfigEventIntervalSize uint64 = 5040
)

type oracleManagerData struct {
	LatestEventBlock *big.Int `json:"latestEventBlock"`
	LatestConfigHash string   `json:"latestConfigHash"`
}

type OracleManager struct {
	sp       IStakeWiseServiceProvider
	dataPath string
	lock     *sync.Mutex

	data   oracleManagerData
	keeper *swcontracts.IKeeper
}

func NewOracleManager(sp IStakeWiseServiceProvider) (*OracleManager, error) {
	dataPath := filepath.Join(sp.GetModuleDir(), swconfig.OracleManagerFile)

	// Load the Oracle data
	var data oracleManagerData
	_, err := os.Stat(dataPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("error checking status of oracle data file [%s]: %w", dataPath, err)
		}
	} else {
		// Read the file
		bytes, err := os.ReadFile(dataPath)
		if err != nil {
			return nil, fmt.Errorf("error reading oracle data file [%s]: %w", dataPath, err)
		}

		// Deserialize it
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			return nil, fmt.Errorf("error deserializing oracle data file [%s]: %w", dataPath, err)
		}
	}

	// Default to the genesis block
	res := sp.GetResources()
	genesisBlock := res.KeeperGenesisBlock
	if data.LatestEventBlock == nil || data.LatestEventBlock.Cmp(genesisBlock) < 0 {
		data.LatestEventBlock = new(big.Int).Set(genesisBlock)
		data.LatestConfigHash = ""
	}

	// Create the contracts
	keeperAddress := res.Keeper
	ec := sp.GetEthClient()
	txMgr := sp.GetTransactionManager()
	keeper, err := swcontracts.NewIKeeper(keeperAddress, ec, txMgr)
	if err != nil {
		return nil, fmt.Errorf("error creating Keeper contract instance: %w", err)
	}

	mgr := &OracleManager{
		sp:       sp,
		dataPath: dataPath,
		lock:     &sync.Mutex{},
		data:     data,
		keeper:   keeper,
	}
	return mgr, nil
}

func (m *OracleManager) GetLatestConfigHash() (string, error) {
	return "", nil
}

func (m *OracleManager) updateEvent(ctx context.Context, logger *slog.Logger) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	ec := m.sp.GetEthClient()

	// Get the scan range
	startBlock := m.data.LatestEventBlock
	latestBlockUint, err := ec.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("error getting latest block number: %w", err)
	}
	latestBlock := new(big.Int).SetUint64(latestBlockUint)

	logger.Debug("Scanning for Oracle config events", "startBlock", startBlock, "latestBlock", latestBlock)
	events, err := m.keeper.ConfigUpdated(startBlock, latestBlock, new(big.Int).SetUint64(OracleConfigEventIntervalSize))
	if err != nil {
		return fmt.Errorf("error scanning for Oracle config events: %w", err)
	}

	// TODO
	logger.Debug(events[0].ConfigIPFSHash)
	return nil
}
