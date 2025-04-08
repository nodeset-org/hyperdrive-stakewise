package swcommon

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Block index table
	BlockIndexTableName      string = "blockIndex"
	BlockIndexDepositsColumn string = "deposits"

	// Table for validator info
	ValidatorsTableName    string = "validators"
	ValidatorsPubkeyColumn string = "pubkey"
	ValidatorsIndexColumn  string = "beaconIndex"
)

// Manages a SQLite database for storing validator information and block indexing
type DatabaseManager struct {
	// The database connection
	db *sql.DB
}

// Creates a new database manager, initializing the database file and tables if it hasn't been created yet
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	// Connect to the database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Start a table creation transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the block index table if it doesn't exist
	createMetadataTableCommand := "CREATE TABLE IF NOT EXISTS " +
		BlockIndexTableName + " (" +
		BlockIndexDepositsColumn + " INTEGER PRIMARY KEY" +
		");"
	_, err = tx.Exec(createMetadataTableCommand)
	if err != nil {
		return nil, fmt.Errorf("error creating %s table: %w", BlockIndexTableName, err)
	}

	// Initialize the metadata table with the latest block numbers
	_, err = tx.Exec(fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (?)",
		BlockIndexTableName,
		BlockIndexDepositsColumn,
	), -1) // -1 means we haven't started indexing yet
	if err != nil {
		return nil, fmt.Errorf("error initializing %s table: %w", BlockIndexTableName, err)
	}

	// Create the validator info table if it doesn't exist
	createValidatorsTableCommand := "CREATE TABLE IF NOT EXISTS " +
		ValidatorsTableName + " (" +
		ValidatorsPubkeyColumn + " BLOB PRIMARY KEY NOT NULL" +
		");"
	fmt.Println(createValidatorsTableCommand)
	_, err = tx.Exec(createValidatorsTableCommand)
	if err != nil {
		return nil, fmt.Errorf("error creating %s table: %w", ValidatorsTableName, err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error creating database tables: %w", err)
	}

	return &DatabaseManager{
		db: db,
	}, nil
}

// Gets the index of a Beacon validator. If the validator does not exist, it returns -1.
func (m *DatabaseManager) DoesValidatorExist(ctx context.Context, pubkey []byte) (bool, error) {
	var index uint64
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s = ?", ValidatorsTableName, ValidatorsPubkeyColumn)
	err := m.db.QueryRowContext(ctx, query, pubkey).Scan(&index)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, fmt.Errorf("error getting validator index: %w", err)
		}
	}
	return true, nil
}

// Get the number of validators in the database - useful to get the next index
func (m *DatabaseManager) GetValidatorCount(ctx context.Context) (uint64, error) {
	var count uint64
	query := fmt.Sprintf("SELECT COUNT(%s) FROM %s", ValidatorsPubkeyColumn, ValidatorsTableName)
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting next validator index: %w", err)
	}
	return count, nil
}

// Get the latest block number checked for deposits
func (m *DatabaseManager) GetLatestBlockChecked(ctx context.Context) (int64, error) {
	var latestBlockChecked int64
	query := fmt.Sprintf("SELECT %s FROM %s", BlockIndexDepositsColumn, BlockIndexTableName)
	err := m.db.QueryRowContext(ctx, query).Scan(&latestBlockChecked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil // No rows means we haven't started indexing yet
		} else {
			return 0, fmt.Errorf("error getting latest block number checked for deposit indexing: %w", err)
		}
	}
	return latestBlockChecked, nil
}

// Adds new validator pubkeys to the database and updates the latest block number checked
func (m *DatabaseManager) AddValidatorPubkeys(ctx context.Context, latestBlockChecked uint64, pubkeys []beacon.ValidatorPubkey) error {
	// Start a batch transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the validators
	for _, pubkey := range pubkeys {
		insertValidatorCommand := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (?)",
			ValidatorsTableName,
			ValidatorsPubkeyColumn,
		)
		_, err := tx.ExecContext(ctx, insertValidatorCommand, pubkey[:])
		if err != nil {
			return fmt.Errorf("error inserting validator %s: %w", pubkey.HexWithPrefix(), err)
		}
	}

	// Update the latest block number
	updateBlockCommand := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (?)",
		BlockIndexTableName,
		BlockIndexDepositsColumn,
	)
	_, err = tx.ExecContext(ctx, updateBlockCommand, latestBlockChecked)
	if err != nil {
		return fmt.Errorf("error updating latest deposits block number: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	return nil
}

// Closes the database connection and saves any changes
func (m *DatabaseManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
