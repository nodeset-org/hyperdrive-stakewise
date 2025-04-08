package swcommon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// The key for the last deposit event block number checked
	LatestDepositEventBlockKey = "lastDepositEventBlock"

	// Prefix for validator pubkeys
	ValidatorKeyPrefix = "vk:"
)

// PebbleManager manages a Pebble database for storing validator information and block indexing
type PebbleManager struct {
	// The database connection
	db *pebble.DB
}

// Creates a new PebbleManager
func NewPebbleManager(dbPath string) (*PebbleManager, error) {
	db, err := pebble.Open(dbPath, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("error opening pebble database: %w", err)
	}

	return &PebbleManager{
		db: db,
	}, nil
}

// Gets the pubkey of a Beacon validator with the given index. If the validator does not exist, it returns nil.
func (m *PebbleManager) DoesValidatorExist(pubkey beacon.ValidatorPubkey) (bool, error) {
	key := []byte(ValidatorKeyPrefix + pubkey.Hex())
	start := time.Now()
	_, closer, err := m.db.Get(key)
	fmt.Printf("Get took %s\n", time.Since(start))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("error getting validator pubkey: %w", err)
	}
	defer closer.Close()
	return true, nil
}

// Get the number of validators in the database - useful to get the next index
func (m *PebbleManager) GetValidatorCount() (uint64, error) {
	iter, err := m.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(ValidatorKeyPrefix),
	})
	if err != nil {
		return 0, fmt.Errorf("error creating iterator: %w", err)
	}
	defer iter.Close()

	count := uint64(0)
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if strings.HasPrefix(string(key), ValidatorKeyPrefix) {
			count++
		}
	}

	return count, nil
}

// Get the latest block number checked for deposits
func (m *PebbleManager) GetLatestBlockChecked() (int64, error) {
	key := []byte(LatestDepositEventBlockKey)
	value, closer, err := m.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return -1, nil
		}
		return 0, fmt.Errorf("error getting latest deposit event block: %w", err)
	}
	defer closer.Close()

	// Convert the value to an int64
	blockNumber := int64(binary.LittleEndian.Uint64(value))
	return blockNumber, nil
}

// Adds new validator pubkeys to the database and updates the latest block number checked
func (m *PebbleManager) AddValidatorPubkeys(latestBlockChecked uint64, pubkeys []beacon.ValidatorPubkey) error {
	batch := m.db.NewBatch()
	defer batch.Close()

	// Insert the pubkeys
	for _, pubkey := range pubkeys {
		key := []byte(ValidatorKeyPrefix + pubkey.Hex())
		value := []byte{}
		err := batch.Set(key, value, pebble.NoSync)
		if err != nil {
			return fmt.Errorf("error inserting validator pubkey %s: %w", pubkey.HexWithPrefix(), err)
		}
	}

	// Update the latest block number
	latestBlockBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(latestBlockBytes, latestBlockChecked)
	err := batch.Set([]byte(LatestDepositEventBlockKey), latestBlockBytes, pebble.NoSync)
	if err != nil {
		return fmt.Errorf("error setting latest deposit event block: %w", err)
	}

	// Commit the batch
	err = batch.Commit(pebble.Sync)
	if err != nil {
		return fmt.Errorf("error committing batch: %w", err)
	}
	stats := batch.CommitStats()
	fmt.Printf("Commit took %s\n", stats.TotalDuration)
	return nil
}

// Close the database
func (m *PebbleManager) Close() error {
	if err := m.db.Close(); err != nil {
		return fmt.Errorf("error closing pebble database: %w", err)
	}
	return nil
}
