package swclient

// TODO: Talk to Joe about removing this entirely
import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alessio/shellescape"
	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
	"gopkg.in/yaml.v2"
)

// Wrapper for global configuration
type GlobalConfig struct {
	ExternalIP string

	// Hyperdrive
	Hyperdrive          *hdconfig.HyperdriveConfig
	HyperdriveResources *hdconfig.MergedResources
}

// Serialize the config and all modules
func (c *GlobalConfig) Serialize() map[string]any {
	return c.Hyperdrive.Serialize(c.GetAllModuleConfigs(), false)
}

// Get the configs for all of the modules in the system
func (c *GlobalConfig) GetAllModuleConfigs() []hdconfig.IModuleConfig {
	return []hdconfig.IModuleConfig{}
}

// Saves a config
func SaveConfig(cfg *GlobalConfig, directory string, filename string) error {
	path := filepath.Join(directory, filename)

	settings := cfg.Serialize()
	configBytes, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("could not serialize settings file: %w", err)
	}

	// Make a tmp file
	// The empty string directs CreateTemp to use the OS's $TMPDIR (or GetTempPath) on windows
	// The * in the second string is replaced with random characters by CreateTemp
	f, err := os.CreateTemp(directory, ".tmp-"+filename+"-*")
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("could not create file to save config to disk... do you need to clean your tmpdir (%s)?: %w", os.TempDir(), err)
		}

		return fmt.Errorf("could not create file to save config to disk: %w", err)
	}
	// Clean up the temporary files
	// This prevents us from filling up `directory` with partially written files on failure
	// If the file is successfully written, it fails with an error since it will be renamed
	// before it is deleted, which we explicitly ignore / don't care about.
	defer func() {
		// Clean up tmp files, if any found
		oldFiles, err := filepath.Glob(filepath.Join(directory, ".tmp-"+filename+"-*"))
		if err != nil {
			// Only possible error is ErrBadPattern, which we should catch
			// during development, since the pattern is a comptime constant.
			panic(err.Error())
		}

		for _, match := range oldFiles {
			os.RemoveAll(match)
		}
	}()

	// Save the serialized settings to the temporary file
	if _, err := f.Write(configBytes); err != nil {
		return fmt.Errorf("could not write Hyperdrive config to %s: %w", shellescape.Quote(path), err)
	}

	// Close the file for writing
	if err := f.Close(); err != nil {
		return fmt.Errorf("error saving Hyperdrive config to %s: %w", shellescape.Quote(path), err)
	}

	// Rename the temp file to overwrite the actual file.
	// On Unix systems this operation is atomic and won't fail if the disk is now full
	if err := os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("error replacing old Hyperdrive config with %s: %w", f.Name(), err)
	}

	// Just in case the rename didn't overwrite (and preserve the perms of) the original file, set them now.
	if err := os.Chmod(path, 0664); err != nil {
		return fmt.Errorf("error updating permissions of %s: %w", path, err)
	}

	return nil
}

// Make a new global config
func NewGlobalConfig(hdCfg *hdconfig.HyperdriveConfig, hdSettings []*hdconfig.HyperdriveSettings) (*GlobalConfig, error) {
	// Make the config
	cfg := &GlobalConfig{
		Hyperdrive: hdCfg,
	}

	// Get the HD resources
	network := hdCfg.Network.Value
	for _, setting := range hdSettings {
		if setting.Key == network {
			cfg.HyperdriveResources = &hdconfig.MergedResources{
				NetworkResources:    setting.NetworkResources,
				HyperdriveResources: setting.HyperdriveResources,
			}
			break
		}
	}
	if cfg.HyperdriveResources == nil {
		return nil, fmt.Errorf("could not find hyperdrive resources for network [%s]", network)
	}
	return cfg, nil
}

// Deserialize the config's modules (assumes the Hyperdrive config itself has already been deserialized)
func (c *GlobalConfig) DeserializeModules() error {

	return nil
}
