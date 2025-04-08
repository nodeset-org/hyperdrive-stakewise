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
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/beacon"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	// The size of the interval to use, in blocks, when scanning blocks for deposit events
	// Set to 1/10th of 1 week, assuming 12 seconds per block
	IntervalSize uint64 = 5040

	// The number of blocks to rewind from the chain head when starting a deposit event scan
	// Set to 1 week, assuming 12 seconds per block
	DepositEventLookbackLimit uint64 = 50400
)

var (
	// The size of the interval to use, in blocks, when scanning blocks for deposit events
	intervalSizeBig *big.Int = new(big.Int).SetUint64(IntervalSize)
)

// Info about a key that is available for use in a deposit
type AvailableKey struct {
	// The pubkey
	PublicKey beacon.ValidatorPubkey `json:"pubkey"`

	// The private key - not serialized for obvious reasons
	PrivateKey *eth2types.BLSPrivateKey `json:"-"`

	// Flag indicating whether or not a deposit event scan (starting with the lookback limit) has been performed for this key
	HasLookbackScanned bool `json:"hasLookbackScanned"`

	// If this pubkey was used already in a previous deposit attempt, this is the Beacon deposit contract's deposit root during that attempt.
	// It's used to compare against the current deposit root to determine if the deposit was unsuccessful and the key can be reused.
	LastDepositRoot common.Hash `json:"lastDepositRoot"`
}

// A reason why a key is ineligible for use in a deposit
type IneligibleReason int

const (
	// The key is ineligible because it doesn't have a private key
	IneligibleReason_NoPrivateKey IneligibleReason = iota

	// The key is ineligible because it hasn't been lookback scanned yet
	IneligibleReason_LookbackScanRequired

	// The key is ineligible because it has already been assigned an index on Beacon
	IneligibleReason_OnBeacon

	// The key is ineligible because it has already been used in a deposit contract event
	IneligibleReason_HasDepositEvent

	// The key is ineligible because it has already been used in a deposit contract event with the same deposit root
	IneligibleReason_AlreadyUsedDepositRoot
)

// AvailableKeyManager manages the keys that have been generated but not yet used for deposits
type AvailableKeyManager struct {
	dataPath      string
	sp            IStakeWiseServiceProvider
	lock          *sync.Mutex
	hasLoadedKeys bool

	data *availableKeyManagerData
}

type availableKeyManagerData struct {
	// The next block to scan for deposit events, assuming no new keys have been added
	NextBlockToScan uint64 `json:"nextBlockToScan"`

	// The list of available keys
	Keys []*AvailableKey `json:"keys"`
}

// Creates a new manager
func NewAvailableKeyManager(sp IStakeWiseServiceProvider) (*AvailableKeyManager, error) {
	dataPath := filepath.Join(sp.GetModuleDir(), swconfig.AvailableKeysFile)

	// Initialize the available key list
	data := new(availableKeyManagerData)
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
		err = json.Unmarshal(bytes, data)
		if err != nil {
			return nil, fmt.Errorf("error deserializing available keys file [%s]: %w", dataPath, err)
		}
	}

	mgr := &AvailableKeyManager{
		dataPath: dataPath,
		sp:       sp,
		lock:     &sync.Mutex{},
		data:     data,
	}
	return mgr, nil
}

// Check if the private keys have been loaded yet
func (m *AvailableKeyManager) HasLoadedKeys() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.hasLoadedKeys
}

// Regenerate the private key for all available keys.
func (m *AvailableKeyManager) LoadPrivateKeys(logger *slog.Logger) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.loadPrivateKeysImpl(logger)
}

// Implementation of the LoadPrivateKeys function, used when the lock is already held
func (m *AvailableKeyManager) loadPrivateKeysImpl(logger *slog.Logger) {
	start := time.Now()
	wallet := m.sp.GetWallet()
	count := 0
	for _, key := range m.data.Keys {
		var err error
		key.PrivateKey, err = wallet.GetPrivateKeyForPubkey(key.PublicKey) // Set to nil on error
		if err != nil {
			logger.Warn(
				"Couldn't get private key for pubkey, it will not be eligible for deposits until this is resolved",
				"pubkey", key.PublicKey.Hex(),
				"error", err,
			)
		} else {
			count++
		}
	}
	logger.Debug(
		"Loaded private keys",
		"count", count,
		"total", len(m.data.Keys),
		"elapsed", time.Since(start),
	)
	m.hasLoadedKeys = true
}

// Add a new key to the list of available keys
func (m *AvailableKeyManager) AddNewKey(key *eth2types.BLSPrivateKey) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Add the new key
	pubkey := beacon.ValidatorPubkey(key.PublicKey().Marshal())
	m.data.Keys = append(m.data.Keys, &AvailableKey{
		PublicKey:          pubkey,
		PrivateKey:         key,
		HasLookbackScanned: false,
	})

	// Save the new list
	err := m.updateData()
	if err != nil {
		return fmt.Errorf("error updating available keys: %w", err)
	}

	return nil
}

// Check if there are any candidate keys ready for validation and potential usage
func (m *AvailableKeyManager) HasKeyCandidates() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.data.Keys) > 0
}

// Check if any of the keys in the list require a lookback scan
func (m *AvailableKeyManager) RequiresLookbackScan() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, key := range m.data.Keys {
		if !key.HasLookbackScanned {
			return true
		}
	}
	return false
}

// Options for customizing the behavior of GetAvailableKeys
type GetAvailableKeyOptions struct {
	// Skip the sync check for the Beacon and Execution clients (typically used when you already know they are synced)
	SkipSyncCheck bool

	// If false, this starts the deposit event scan from the last block that was scanned.
	// This is the default behavior, and much faster than scanning the entire history - useful if you've already scanned the logs before.
	// If true, this will start the scan at DepositEventLookbackLimit blocks instead.
	// Set this if you have a new key that hasn't had its history checked yet.
	DoLookbackScan bool
}

// Get the keys that can be used for new deposits from the list of available keys.
// As a side-effect this refreshes the backing list by filtering out any that have already been used in a deposit and saves it to disk.
func (m *AvailableKeyManager) GetAvailableKeys(
	ctx context.Context,
	logger *slog.Logger,
	beaconDepositRoot common.Hash,
	options GetAvailableKeyOptions,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys map[*AvailableKey]IneligibleReason,
	err error,
) {
	start := time.Now()
	m.lock.Lock()
	defer m.lock.Unlock()
	logger.Debug("Got key lock", "elapsed", time.Since(start))

	// Load the keys from disk if they haven't been loaded yet
	if !m.hasLoadedKeys {
		m.loadPrivateKeysImpl(logger)
	}

	// Remove keys that don't have a private key
	ineligibleKeys = make(map[*AvailableKey]IneligibleReason)
	goodKeys, badKeys := m.filterKeysOnPrivateKey(m.data.Keys)
	for _, key := range badKeys {
		ineligibleKeys[key] = IneligibleReason_NoPrivateKey
	}

	// Remove keys that haven't been lookback scanned yet if we aren't doing one
	if !options.DoLookbackScan {
		goodKeys, badKeys = m.filterKeysOnLookbackScanned(goodKeys)
		for _, key := range badKeys {
			ineligibleKeys[key] = IneligibleReason_LookbackScanRequired
		}
	}

	// Remove keys that have already been assigned an index on Beacon
	start = time.Now()
	goodKeys, badKeys, err = m.filterKeysOnBeacon(ctx, goodKeys, options.SkipSyncCheck)
	if err != nil {
		return nil, nil, fmt.Errorf("error filtering keys via Beacon indices: %w", err)
	}
	logger.Debug("Filtered keys on Beacon", "available", len(goodKeys), "elapsed", time.Since(start))
	for _, key := range badKeys {
		ineligibleKeys[key] = IneligibleReason_OnBeacon
	}

	// Remove keys that have already been used in a deposit contract event
	start = time.Now()
	goodKeys, badKeys, err = m.filterKeysOnDepositContract(ctx, logger, goodKeys, options.SkipSyncCheck, options.DoLookbackScan)
	if err != nil {
		return nil, nil, fmt.Errorf("error filtering keys via deposit contract events: %w", err)
	}
	logger.Debug("Filtered keys on deposit contract", "available", len(goodKeys), "elapsed", time.Since(start))
	for _, key := range badKeys {
		ineligibleKeys[key] = IneligibleReason_HasDepositEvent
	}

	m.data.Keys = goodKeys // Save all of the keys before filtering by deposit root
	// because ones with this deposit root are in the mempool and may get reverted.
	// If that happens then they can be reused later.
	start = time.Now()
	goodKeys, badKeys = m.filterKeysOnDepositRoot(goodKeys, beaconDepositRoot)
	logger.Debug("Filtered keys on deposit root", "available", len(goodKeys), "elapsed", time.Since(start))
	for _, key := range badKeys {
		ineligibleKeys[key] = IneligibleReason_AlreadyUsedDepositRoot
	}

	// Save the new list
	err = m.updateData()
	if err != nil {
		return nil, nil, fmt.Errorf("error updating available keys: %w", err)
	}
	logger.Debug(
		"Updated available keys",
		"available", len(goodKeys),
		"elapsed", time.Since(start),
	)

	// Return the eligible ones
	return goodKeys, ineligibleKeys, nil
}

// Set the last deposit root for a list of keys, indicating they will be used in a new deposit
func (m *AvailableKeyManager) SetLastDepositRoot(keys []*AvailableKey, lastDepositRoot common.Hash) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Update and save
	for _, key := range keys {
		key.LastDepositRoot = lastDepositRoot
	}
	err := m.updateData()
	if err != nil {
		return fmt.Errorf("error updating available keys: %w", err)
	}

	return nil
}

// Filter the list of available keys to remove any that don't have a private key
func (m *AvailableKeyManager) filterKeysOnPrivateKey(
	keys []*AvailableKey,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys []*AvailableKey,
) {
	eligibleKeys = []*AvailableKey{}
	ineligibleKeys = []*AvailableKey{}
	for _, key := range keys {
		if key.PrivateKey == nil {
			ineligibleKeys = append(ineligibleKeys, key)
		} else {
			eligibleKeys = append(eligibleKeys, key)
		}
	}
	return eligibleKeys, ineligibleKeys
}

// Filter the list of available keys that haven't lookback scanned yet (only used if getting available keys without lookback enabled)
func (m *AvailableKeyManager) filterKeysOnLookbackScanned(
	keys []*AvailableKey,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys []*AvailableKey,
) {
	eligibleKeys = []*AvailableKey{}
	ineligibleKeys = []*AvailableKey{}
	for _, key := range keys {
		if key.HasLookbackScanned {
			eligibleKeys = append(eligibleKeys, key)
		} else {
			ineligibleKeys = append(ineligibleKeys, key)
		}
	}
	return eligibleKeys, ineligibleKeys
}

// Filter the list of available keys to remove any that have already been assigned an index on the Beacon chain
func (m *AvailableKeyManager) filterKeysOnBeacon(
	ctx context.Context,
	keys []*AvailableKey,
	skipSyncCheck bool,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys []*AvailableKey,
	err error,
) {
	// Get the Beacon client and make sure it's synced
	bn := m.sp.GetBeaconClient()
	if !skipSyncCheck {
		if err := m.sp.RequireBeaconClientSynced(ctx); err != nil {
			return nil, nil, err
		}
	}

	// Remove keys that have already been assigned an index on Beacon
	pubkeys := make([]beacon.ValidatorPubkey, len(keys))
	for i, data := range keys {
		pubkeys[i] = data.PublicKey
	}
	statuses, err := bn.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting validator statuses: %w", err)
	}

	// Ignore keys that are already active on Beacon
	eligibleKeys = []*AvailableKey{}
	ineligibleKeys = []*AvailableKey{}
	for _, key := range keys {
		_, exists := statuses[key.PublicKey]
		if exists {
			key.PrivateKey = nil // Empty out the private key so it's not resident in memory
			ineligibleKeys = append(ineligibleKeys, key)
		} else {
			eligibleKeys = append(eligibleKeys, key)
		}
	}
	return eligibleKeys, ineligibleKeys, nil
}

// Filter the list of available keys to remove any that have deposit events in the deposit contract logs
func (m *AvailableKeyManager) filterKeysOnDepositContract(
	ctx context.Context,
	logger *slog.Logger,
	keys []*AvailableKey,
	skipSyncCheck bool,
	doLookbackScan bool,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys []*AvailableKey,
	err error,
) {
	// Get the Execution client and make sure it's synced
	ec := m.sp.GetEthClient()
	if !skipSyncCheck {
		if err := m.sp.RequireEthClientSynced(ctx); err != nil {
			return nil, nil, err
		}
	}

	// Figure out where to start
	currentBlock, err := ec.BlockNumber(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting current block number: %w", err)
	}

	var startBlock uint64
	if doLookbackScan {
		if DepositEventLookbackLimit > currentBlock {
			// If the limit is greater than the current block, like for new chains, start from 0
			startBlock = 0
		} else {
			startBlock = currentBlock - DepositEventLookbackLimit
		}
	} else {
		startBlock = m.data.NextBlockToScan
		if currentBlock-startBlock > DepositEventLookbackLimit {
			// If the next block to scan is too far in the past, start from the limit
			startBlock = currentBlock - DepositEventLookbackLimit
		}
	}

	// Get the deposit events
	logger.Debug(
		"Getting deposit events",
		"start", startBlock,
		"end", currentBlock,
	)
	pubkeys := make([]beacon.ValidatorPubkey, len(keys))
	for i, data := range keys {
		pubkeys[i] = data.PublicKey
	}
	depositContract := m.sp.GetBeaconDepositContract()
	depositEvents, err := depositContract.DepositEventsForPubkeys(
		pubkeys,
		new(big.Int).SetUint64(startBlock),
		new(big.Int).SetUint64(currentBlock),
		intervalSizeBig,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting deposit events: %w", err)
	}

	// Set the lookback flag for all keys
	if doLookbackScan {
		for _, key := range keys {
			key.HasLookbackScanned = true
		}
	}

	// Ignore keys that are already in the deposit logs
	eligibleKeys = []*AvailableKey{}
	ineligibleKeys = []*AvailableKey{}
	for _, key := range keys {
		_, exists := depositEvents[key.PublicKey]
		if exists {
			key.PrivateKey = nil // Empty out the private key so it's not resident in memory
			ineligibleKeys = append(ineligibleKeys, key)
		} else {
			eligibleKeys = append(eligibleKeys, key)
		}
	}

	// Update the next block to scan if there aren't any errors
	m.data.NextBlockToScan = currentBlock + 1
	return eligibleKeys, ineligibleKeys, nil
}

// Filter the list of available keys to remove any that were already provided for the given deposit root
func (m *AvailableKeyManager) filterKeysOnDepositRoot(
	keys []*AvailableKey,
	beaconDepositRoot common.Hash,
) (
	eligibleKeys []*AvailableKey,
	ineligibleKeys []*AvailableKey,
) {
	eligibleKeys = []*AvailableKey{}
	ineligibleKeys = []*AvailableKey{}
	for _, key := range keys {
		if key.LastDepositRoot == beaconDepositRoot {
			ineligibleKeys = append(ineligibleKeys, key)
		} else {
			eligibleKeys = append(eligibleKeys, key)
		}
	}
	return eligibleKeys, ineligibleKeys
}

// Update the contents of the available keys file
func (m *AvailableKeyManager) updateData() error {
	// Serialize the key list
	bytes, err := json.Marshal(m.data)
	if err != nil {
		return fmt.Errorf("error serializing available keys data: %w", err)
	}

	// Write it
	err = os.WriteFile(m.dataPath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("error saving available keys data to disk: %w", err)
	}
	return nil
}
