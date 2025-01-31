package config

import (
	"fmt"
	"os"

	"github.com/nodeset-org/hyperdrive-stakewise/adapter/config/ids"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// Configuration manager
type AdapterConfigManager struct {
	// The adapter configuration
	AdapterConfig *StakeWiseConfig

	// The native configuration manager
	// nativeConfigManager *nativecfg.ConfigManager

	// The path to the adapter configuration file
	adapterConfigPath string
}

// Create a new configuration manager for the adapter
func NewAdapterConfigManager(c *cli.Context) (*AdapterConfigManager, error) {
	// configDir := c.String(utils.ConfigDirFlag.Name)
	// if configDir == "" {
	// 	return nil, fmt.Errorf("config directory is required")
	// }
	// return &AdapterConfigManager{
	// 	// nativeConfigManager: nativecfg.NewConfigManager(filepath.Join(configDir, utils.ServiceConfigFile)),
	// 	adapterConfigPath: filepath.Join(configDir, utils.AdapterConfigFile),
	// }, nil
	return nil, nil
}

// Load the configuration from disk
func (m *AdapterConfigManager) LoadConfigFromDisk() (*StakeWiseConfig, error) {
	// Load the native config
	// nativeCfg, err := m.nativeConfigManager.LoadConfigFromFile()
	// if err != nil {
	// 	return nil, fmt.Errorf("error loading service config: %w", err)
	// }

	// // Check if the adapter config exists
	// _, err = os.Stat(m.adapterConfigPath)
	// if errors.Is(err, fs.ErrNotExist) {
	// 	return nil, nil
	// }

	// Load it
	bytes, err := os.ReadFile(m.adapterConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file [%s]: %w", m.adapterConfigPath, err)
	}

	// Deserialize it
	cfgInstance := map[string]any{}
	err = yaml.Unmarshal(bytes, &cfgInstance)
	if err != nil {
		return nil, fmt.Errorf("error deserializing adapter config file [%s]: %w", m.adapterConfigPath, err)
	}
	serverCfg := NewServerConfig()
	portInt := cfgInstance[ids.PortID].(int)
	serverCfg.Port.Value = uint64(portInt)
	serverCfg.PortMode.Value = PortMode(cfgInstance[ids.PortModeID].(string))

	// Merge the configs
	// modCfg := ConvertToMetadata(nativeCfg)
	// modCfg.ServerConfig = serverCfg
	// m.AdapterConfig = modCfg
	// return modCfg, nil
	return nil, nil
}

// Save the configuration to a file. If the config hasn't been loaded yet, this doesn't do anything.
func (m *AdapterConfigManager) SaveConfigToDisk() error {
	if m.AdapterConfig == nil {
		return nil
	}

	// // Save the native config
	// nativeCfg := m.AdapterConfig.ConvertToNative()
	// m.nativeConfigManager.Config = nativeCfg
	// err := m.nativeConfigManager.SaveConfigToFile()
	// if err != nil {
	// 	return fmt.Errorf("error saving service config: %w", err)
	// }

	// Serialize the adapter config
	// modCfg := hdconfig.CreateInstanceFromMetadata(m.AdapterConfig.ServerConfig)
	// bytes, err := yaml.Marshal(modCfg)
	// if err != nil {
	// 	return fmt.Errorf("error serializing adapter config: %w", err)
	// }

	// Write it
	// err = os.WriteFile(m.adapterConfigPath, bytes, nativecfg.ConfigFileMode)
	// if err != nil {
	// 	return fmt.Errorf("error writing adapter config file [%s]: %w", m.adapterConfigPath, err)
	// }
	return nil
}
