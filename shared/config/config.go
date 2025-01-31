package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileMode os.FileMode = 0644
)

type NativeStakeWiseConfigSettings struct {
	Enabled              bool   `json:"enabled"`
	ApiPort              uint16 `json:"apiPort"`
	VerifyDepositsRoot   bool   `json:"verifyDepositsRoot"`
	DaemonContainerTag   string `json:"daemonContainerTag"`
	OperatorContainerTag string `json:"operatorContainerTag"`
	AdditionalOpFlags    string `json:"additionalOpFlags"`

	// VcCommon   *config.ValidatorClientCommonConfig `json:"vcCommon"`
	// Lighthouse *config.LighthouseVcConfig          `json:"lighthouse"`
	// Lodestar   *config.LodestarVcConfig            `json:"lodestar"`
	// Nimbus     *config.NimbusVcConfig              `json:"nimbus"`
	// Prysm      *config.PrysmVcConfig               `json:"prysm"`
	// Teku       *config.TekuVcConfig                `json:"teku"`

	Version string `json:"Version"`
}

// Configuration manager
type ConfigManager struct {
	// The configuration
	Config *NativeStakeWiseConfigSettings

	// The path to the configuration file
	ConfigPath string
}

// Create a new configuration manager
func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{
		ConfigPath: path,
	}
}

// Load the configuration from a file
func (m *ConfigManager) LoadConfigFromFile() (*NativeStakeWiseConfigSettings, error) {
	// Check if the file exists
	_, err := os.Stat(m.ConfigPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	// Load it
	bytes, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file [%s]: %w", m.ConfigPath, err)
	}

	// Deserialize it
	cfg := NativeStakeWiseConfigSettings{}
	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error deserializing config file [%s]: %w", m.ConfigPath, err)
	}
	return &cfg, nil
}

// Save the configuration to a file. If the config hasn't been loaded yet, this doesn't do anything.
func (m *ConfigManager) SaveConfigToFile() error {
	if m.Config == nil {
		return nil
	}

	// Serialize it
	bytes, err := yaml.Marshal(m.Config)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	// Write it
	err = os.WriteFile(m.ConfigPath, bytes, ConfigFileMode)
	if err != nil {
		return fmt.Errorf("error writing config file [%s]: %w", m.ConfigPath, err)
	}
	return nil
}
