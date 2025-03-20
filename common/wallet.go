package swcommon

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	"github.com/nodeset-org/hyperdrive-daemon/shared"
	"github.com/nodeset-org/hyperdrive-daemon/shared/config"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	walletDataFilename string = "wallet_data"
)

// Data relating to Stakewise's wallet
type stakewiseWalletData struct {
	// The next account to generate the key for
	NextAccount uint64 `json:"nextAccount"`
}

// Wallet manager for the Stakewise daemon
type Wallet struct {
	validatorManager          *validator.ValidatorManager
	stakewiseWalletFilePath   string
	stakewisePasswordFilePath string
	stakewiseKeystoreManager  *stakewiseKeystoreManager
	data                      stakewiseWalletData
	sp                        services.IModuleServiceProvider
}

// Create a new wallet
func NewWallet(sp services.IModuleServiceProvider) (*Wallet, error) {
	moduleDir := sp.GetModuleDir()
	validatorPath := filepath.Join(moduleDir, config.ValidatorsDirectory)
	wallet := &Wallet{
		sp:                        sp,
		validatorManager:          validator.NewValidatorManager(validatorPath),
		stakewiseWalletFilePath:   filepath.Join(moduleDir, swconfig.WalletFilename),
		stakewisePasswordFilePath: filepath.Join(moduleDir, swconfig.PasswordFilename),
	}

	err := wallet.Reload()
	if err != nil {
		return nil, fmt.Errorf("error loading wallet: %w", err)
	}
	return wallet, nil
}

// Reload the wallet data from disk
func (w *Wallet) Reload() error {
	// Check if the wallet data exists
	moduleDir := w.sp.GetModuleDir()
	dataPath := filepath.Join(moduleDir, walletDataFilename)
	_, err := os.Stat(dataPath)
	if errors.Is(err, fs.ErrNotExist) {
		// No data yet, so make some
		w.data = stakewiseWalletData{
			NextAccount: 0,
		}

		// Save it
		err = w.saveData()
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("error checking status of wallet file [%s]: %w", dataPath, err)
	} else {
		// Read it
		bytes, err := os.ReadFile(dataPath)
		if err != nil {
			return fmt.Errorf("error loading wallet data: %w", err)
		}
		var data stakewiseWalletData
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			return fmt.Errorf("error deserializing wallet data: %w", err)
		}
		w.data = data
	}

	// Make the Stakewise keystore manager
	stakewiseKeystoreMgr, err := newStakewiseKeystoreManager(moduleDir)
	if err != nil {
		return fmt.Errorf("error creating Stakewise keystore manager: %w", err)
	}
	w.stakewiseKeystoreManager = stakewiseKeystoreMgr
	return nil
}

// Generate a new validator key and save it
func (w *Wallet) GenerateNewValidatorKey() (*eth2types.BLSPrivateKey, error) {
	// Get the path for the next validator key
	path := fmt.Sprintf(shared.StakeWiseValidatorPath, w.data.NextAccount)

	// Ask the HD daemon to generate the key
	client := w.sp.GetHyperdriveClient()
	response, err := client.Wallet.GenerateValidatorKey(path)
	if err != nil {
		return nil, fmt.Errorf("error generating validator key for path [%s]: %w", path, err)
	}

	// Increment the next account index first for safety
	w.data.NextAccount++
	err = w.saveData()
	if err != nil {
		return nil, err
	}

	// Save the key to the VC stores
	key, err := eth2types.BLSPrivateKeyFromBytes(response.Data.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error converting BLS private key for path %s: %w", path, err)
	}
	err = w.validatorManager.StoreKey(key, path)
	if err != nil {
		return nil, fmt.Errorf("error saving validator key: %w", err)
	}

	// Save the key to the Stakewise folder
	err = w.stakewiseKeystoreManager.StoreValidatorKey(key, path)
	if err != nil {
		return nil, fmt.Errorf("error saving validator key to the Stakewise store: %w", err)
	}
	return key, nil
}

// Get the private validator key with the corresponding pubkey
func (w *Wallet) GetPrivateKeyForPubkey(pubkey beacon.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {
	return w.stakewiseKeystoreManager.LoadValidatorKey(pubkey)
}

// Get the private validator key with the corresponding pubkey
func (w *Wallet) DerivePubKeys(privateKeys []*eth2types.BLSPrivateKey) ([]beacon.ValidatorPubkey, error) {
	publicKeys := make([]beacon.ValidatorPubkey, 0, len(privateKeys))

	for i, privateKey := range privateKeys {
		if privateKey == nil {
			return nil, fmt.Errorf("nil private key encountered at index %d", i)
		}

		validatorPubkey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
		publicKeys = append(publicKeys, validatorPubkey)
	}

	return publicKeys, nil
}

// Gets all of the validator private keys that are stored in the Stakewise keystore folder
func (w *Wallet) GetAllPrivateKeys() ([]*eth2types.BLSPrivateKey, error) {
	dir := w.stakewiseKeystoreManager.GetKeystoreDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating Stakewise keystore folder [%s]: %w", dir, err)
	}

	// Go through each file
	keys := []*eth2types.BLSPrivateKey{}
	for _, file := range files {
		filename := file.Name()
		if !strings.HasPrefix(filename, keystorePrefix) || !strings.HasSuffix(filename, keystoreSuffix) {
			continue
		}

		// Get the pubkey from the filename
		trimmed := strings.TrimPrefix(filename, keystorePrefix)
		trimmed = strings.TrimSuffix(trimmed, keystoreSuffix)
		pubkey, err := beacon.HexToValidatorPubkey(trimmed)
		if err != nil {
			return nil, fmt.Errorf("error getting pubkey for keystore file [%s]: %w", filename, err)
		}

		// Load it
		key, err := w.stakewiseKeystoreManager.LoadValidatorKey(pubkey)
		if err != nil {
			return nil, fmt.Errorf("error loading validator keystore file [%s]: %w", filename, err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Saves the Stakewise wallet and password files
func (w *Wallet) SaveStakewiseWallet(ethKey []byte, password string) error {
	// Write the wallet to disk
	err := os.WriteFile(w.stakewiseWalletFilePath, ethKey, fileMode)
	if err != nil {
		return fmt.Errorf("error saving wallet keystore to disk: %w", err)
	}

	// Write the password to disk
	err = os.WriteFile(w.stakewisePasswordFilePath, []byte(password), fileMode)
	if err != nil {
		return fmt.Errorf("error saving wallet password to disk: %w", err)
	}
	return nil
}

// Check if the Stakewise wallet and password files exist
func (w *Wallet) CheckIfStakewiseWalletExists() (bool, error) {
	// Check the wallet file
	_, err := os.Stat(w.stakewiseWalletFilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking status of Stakewise wallet file [%s]: %w", w.stakewiseWalletFilePath, err)
	}

	// Check the password file
	_, err = os.Stat(w.stakewisePasswordFilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking status of Stakewise password file [%s]: %w", w.stakewisePasswordFilePath, err)
	}

	return true, nil
}

// Write the wallet data to disk
func (w *Wallet) saveData() error {
	// Serialize it
	dataPath := filepath.Join(w.sp.GetModuleDir(), walletDataFilename)
	bytes, err := json.Marshal(w.data)
	if err != nil {
		return fmt.Errorf("error serializing wallet data: %w", err)
	}

	// Save it
	err = os.WriteFile(dataPath, bytes, fileMode)
	if err != nil {
		return fmt.Errorf("error saving wallet data: %w", err)
	}
	return nil
}
