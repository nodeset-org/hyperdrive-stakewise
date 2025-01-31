package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	sharedconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Configuration manager
type AdapterConfigManager struct {
	// The adapter configuration
	AdapterConfig *StakeWiseConfigSettings

	// The native configuration manager
	nativeConfigManager *sharedconfig.ConfigManager

	// The path to the adapter configuration file
	adapterConfigPath string
}

// Create a new configuration manager for the adapter
func NewAdapterConfigManager(c *cli.Context) (*AdapterConfigManager, error) {
	configDir := c.String(utils.ConfigDirFlag.Name)
	if configDir == "" {
		return nil, fmt.Errorf("config directory is required")
	}
	return &AdapterConfigManager{
		nativeConfigManager: sharedconfig.NewConfigManager(filepath.Join(configDir, utils.ServiceConfigFile)),
		adapterConfigPath:   filepath.Join(configDir, utils.AdapterConfigFile),
	}, nil
}

// Load the configuration from disk
func (m *AdapterConfigManager) LoadConfigFromDisk() (*StakeWiseConfigSettings, error) {
	// Load the native config
	nativeCfg, err := m.nativeConfigManager.LoadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("error loading service config: %w", err)
	}

	// Check if the adapter config exists
	_, err = os.Stat(m.adapterConfigPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	// Load it
	// bytes, err := os.ReadFile(m.adapterConfigPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("error reading config file [%s]: %w", m.adapterConfigPath, err)
	// }

	// Deserialize it
	modCfg := CreateInstanceFromNativeConfig(nativeCfg)
	// err = yaml.Unmarshal(bytes, &modCfg.ServerConfig)
	// if err != nil {
	// 	return nil, fmt.Errorf("error deserializing adapter config file [%s]: %w", m.adapterConfigPath, err)
	// }
	m.AdapterConfig = modCfg
	return modCfg, nil
}

// Save the configuration to a file. If the config hasn't been loaded yet, this doesn't do anything.
func (m *AdapterConfigManager) SaveConfigToDisk() error {
	if m.AdapterConfig == nil {
		return nil
	}

	// Save the native config
	nativeCfg := ConvertInstanceToNativeConfig(m.AdapterConfig)
	m.nativeConfigManager.Config = nativeCfg
	err := m.nativeConfigManager.SaveConfigToFile()
	if err != nil {
		return fmt.Errorf("error saving service config: %w", err)
	}

	// Serialize the adapter config
	bytes, err := yaml.Marshal(m.AdapterConfig)
	if err != nil {
		return fmt.Errorf("error serializing adapter config: %w", err)
	}

	// Write it
	err = os.WriteFile(m.adapterConfigPath, bytes, sharedconfig.ConfigFileMode)
	if err != nil {
		return fmt.Errorf("error writing adapter config file [%s]: %w", m.adapterConfigPath, err)
	}
	return nil
}
