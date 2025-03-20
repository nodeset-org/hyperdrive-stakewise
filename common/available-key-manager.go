package swcommon

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// The size of the interval to use, in blocks, when scanning blocks for deposit events
	// Set to 1/10th of 1 week, assuming 12 seconds per block
	IntervalSize uint64 = 5040

	// The number of blocks to rewind from the chain head when starting a deposit event scan
	// Set to 1 week, assuming 12 seconds per block
	DepositEventLookbackLimit uint64 = 50400
)

// Info about a key that is available for use in a deposit
type AvailableKey struct {
	// The pubkey
	PublicKey beacon.ValidatorPubkey `json:"pubkey"`

	// If this pubkey was used already in a previous deposit attempt, this is the Beacon deposit contract's deposit root during that attempt.
	// It's used to compare against the current deposit root to determine if the deposit was unsuccessful and the key can be reused.
	LastDepositRoot common.Hash `json:"lastDepositRoot"`
}

// AvailableKeyManager manages the keys that have been generated but not yet used for deposits
type AvailableKeyManager struct {
	dataPath string
	sp       IStakeWiseServiceProvider
	lock     *sync.Mutex

	keys []AvailableKey
}

// Creates a new manager
func NewAvailableKeyManager(sp IStakeWiseServiceProvider) (*AvailableKeyManager, error) {
	dataPath := filepath.Join(sp.GetModuleDir(), swconfig.AvailableKeysFile)

	// Initialize the available key list
	keys := []AvailableKey{}
	_, err := os.Stat(dataPath)
	if err != nil {
		// If the file doesn't exist, that's fine - use the empty list
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("error checking status of available keys file [%s]: %w", dataPath, err)
		}
	} else {
		// Read the file
		bytes, err := os.ReadFile(dataPath)
		if err != nil {
			return nil, fmt.Errorf("error reading available keys file [%s]: %w", dataPath, err)
		}

		// Deserialize it
		var keys []AvailableKey
		err = json.Unmarshal(bytes, &keys)
		if err != nil {
			return nil, fmt.Errorf("error deserializing available keys file [%s]: %w", dataPath, err)
		}
	}

	mgr := &AvailableKeyManager{
		dataPath: dataPath,
		sp:       sp,
		keys:     keys,
		lock:     &sync.Mutex{},
	}
	return mgr, nil
}

// Add a new key to the list of available keys
func (m *AvailableKeyManager) AddNewKey(key beacon.ValidatorPubkey) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Add the new key
	m.keys = append(m.keys, AvailableKey{
		PublicKey: key,
	})

	// Save the new list
	err := m.updateAvailableKeys()
	if err != nil {
		return fmt.Errorf("error updating available keys: %w", err)
	}

	return nil
}

// Get the keys that can be used for new deposits from the list of available keys.
// As a side-effect this refreshes the backing list by filtering out any that have already been used in a deposit and saves it to disk.
func (m *AvailableKeyManager) GetAvailableKeys(ctx context.Context, beaconDepositRoot common.Hash, skipSyncCheck bool) ([]beacon.ValidatorPubkey, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Read the file
	bytes, err := os.ReadFile(m.dataPath)
	if err != nil {
		return nil, fmt.Errorf("error reading available keys file [%s]: %w", m.dataPath, err)
	}

	// Deserialize it
	var keys []AvailableKey
	err = json.Unmarshal(bytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("error deserializing available keys file [%s]: %w", m.dataPath, err)
	}

	// Filter the keys
	keys, err = m.filterKeysOnBeacon(ctx, keys, skipSyncCheck)
	if err != nil {
		return nil, fmt.Errorf("error filtering keys via Beacon indices: %w", err)
	}
	keys, err = m.filterKeysOnDepositContract(ctx, keys, skipSyncCheck)
	if err != nil {
		return nil, fmt.Errorf("error filtering keys via deposit contract events: %w", err)
	}
	m.keys = keys // Save all of the keys before filtering by deposit root
	// because ones with this deposit root are in the mempool and may get reverted.
	// If that happens then they can be reused later.
	eligibleKeys := m.filterKeysOnDepositRoot(keys, beaconDepositRoot)

	// Save the new list
	err = m.updateAvailableKeys()
	if err != nil {
		return nil, fmt.Errorf("error updating available keys: %w", err)
	}

	// Return the eligible ones
	pubkeys := make([]beacon.ValidatorPubkey, len(eligibleKeys))
	for i, key := range eligibleKeys {
		pubkeys[i] = key.PublicKey
	}
	return pubkeys, nil
}

// Set the last deposit root for a list of keys, indicating they will be used in a new deposit
func (m *AvailableKeyManager) SetLastDepositRoot(pubkeys []beacon.ValidatorPubkey, lastDepositRoot common.Hash) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Find the keys
	for _, pubkey := range pubkeys {
		for i := range m.keys {
			if m.keys[i].PublicKey == pubkey {
				m.keys[i].LastDepositRoot = lastDepositRoot
				break
			}
		}
	}

	// Save the new list
	err := m.updateAvailableKeys()
	if err != nil {
		return fmt.Errorf("error updating available keys: %w", err)
	}

	return nil
}

// Filter the list of available keys to remove any that have already been assigned an index on the Beacon chain
func (m *AvailableKeyManager) filterKeysOnBeacon(ctx context.Context, keys []AvailableKey, skipSyncCheck bool) ([]AvailableKey, error) {
	// Get the Beacon client and make sure it's synced
	bn := m.sp.GetBeaconClient()
	if !skipSyncCheck {
		if err := m.sp.RequireBeaconClientSynced(ctx); err != nil {
			return nil, err
		}
	}

	// Remove keys that have already been assigned an index on Beacon
	pubkeys := make([]beacon.ValidatorPubkey, len(keys))
	for i, data := range keys {
		pubkeys[i] = data.PublicKey
	}
	statuses, err := bn.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting validator statuses: %w", err)
	}

	// Ignore keys that are already active on Beacon
	eligibleKeys := []AvailableKey{}
	for _, key := range keys {
		_, exists := statuses[key.PublicKey]
		if exists {
			continue
		}
		eligibleKeys = append(eligibleKeys, key)
	}
	return eligibleKeys, nil
}

// Filter the list of available keys to remove any that have deposit events in the deposit contract logs
func (m *AvailableKeyManager) filterKeysOnDepositContract(ctx context.Context, keys []AvailableKey, skipSyncCheck bool) ([]AvailableKey, error) {
	// Get the Execution client and make sure it's synced
	ec := m.sp.GetEthClient()
	if !skipSyncCheck {
		if err := m.sp.RequireEthClientSynced(ctx); err != nil {
			return nil, err
		}
	}

	// Figure out where to start
	currentBlock, err := ec.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting current block number: %w", err)
	}
	startBlock := currentBlock - DepositEventLookbackLimit
	if DepositEventLookbackLimit > currentBlock {
		startBlock = 0
	}

	// Get the deposit events
	pubkeys := make([]beacon.ValidatorPubkey, len(keys))
	for i, data := range keys {
		pubkeys[i] = data.PublicKey
	}
	depositContract := m.sp.GetBeaconDepositContract()
	depositEvents, err := depositContract.DepositEvents(pubkeys, new(big.Int).SetUint64(startBlock), new(big.Int).SetUint64(IntervalSize))
	if err != nil {
		return nil, fmt.Errorf("error getting deposit events: %w", err)
	}

	// Ignore keys that are already in the deposit logs
	eligibleKeys := []AvailableKey{}
	for _, key := range keys {
		_, exists := depositEvents[key.PublicKey]
		if exists {
			continue
		}
		eligibleKeys = append(eligibleKeys, key)
	}
	return eligibleKeys, nil
}

// Filter the list of available keys to remove any that were already provided for the given deposit root
func (m *AvailableKeyManager) filterKeysOnDepositRoot(keys []AvailableKey, beaconDepositRoot common.Hash) []AvailableKey {
	eligibleKeys := []AvailableKey{}
	for _, key := range keys {
		if key.LastDepositRoot == beaconDepositRoot {
			continue
		}
		eligibleKeys = append(eligibleKeys, key)
	}
	return eligibleKeys
}

// Update the contents of the available keys file
func (m *AvailableKeyManager) updateAvailableKeys() error {
	// Serialize the key list
	bytes, err := json.Marshal(m.keys)
	if err != nil {
		return fmt.Errorf("error serializing available keys: %w", err)
	}

	// Write it
	err = os.WriteFile(m.dataPath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("error saving available keys to disk: %w", err)
	}
	return nil
}
