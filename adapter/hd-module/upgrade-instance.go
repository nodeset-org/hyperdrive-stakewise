package hdmodule

import (
	"github.com/nodeset-org/hyperdrive-stakewise/adapter/utils"
	modconfig "github.com/nodeset-org/hyperdrive/modules/config"
	"github.com/urfave/cli/v2"
)

// Request format for `upgrade-instance`
type upgradeInstanceRequest struct {
	utils.KeyedRequest

	// The current config instance
	Instance *modconfig.ModuleInstance `json:"instance"`
}

// Handle the `upgrade-instance` command
func upgradeInstance(c *cli.Context) error {
	return nil
	// // Get the request
	// request, err := utils.HandleKeyedRequest[*upgradeInstanceRequest](c)
	// if err != nil {
	// 	return err
	// }
	// modInstance := request.Instance

	// // Switch on the instance version
	// var settings config.HyperdriveEthereumConfigSettings
	// v0_2_0 := semver.MustParse("0.2.0")
	// version := semver.MustParse(modInstance.Version)
	// if version.LT(v0_2_0) {
	// 	// Upgrade from 0.1.0 to the latest
	// 	oldSettings, err := deserializeSettings_v0_1_0(modInstance.Settings)
	// 	if err != nil {
	// 		return fmt.Errorf("error deserializing settings: %w", err)
	// 	}
	// 	settings = config.UpgradeSettings(oldSettings)
	// } else {
	// 	// Deserialize the settings
	// 	settings, err = deserializeSettings_Latest(modInstance.Settings)
	// 	if err != nil {
	// 		return fmt.Errorf("error deserializing settings: %w", err)
	// 	}
	// }

	// // Create the response
	// response := modconfig.ModuleInstance{
	// 	Enabled: modInstance.Enabled,
	// 	Version: shared.HyperdriveEthereumVersion,
	// }
	// response.SetSettingsFromKnownType(settings)

	// // Marshal it
	// bytes, err := json.Marshal(response)
	// if err != nil {
	// 	return fmt.Errorf("error marshalling upgrade-instance response: %w", err)
	// }

	// // Print it
	// fmt.Println(string(bytes))
	// return nil
}

// Deserialize the settings for a v0.1.0 configuration
// func deserializeSettings_v0_1_0(settings map[string]any) (*v0_1_0.ExampleConfigSettings, error) {
// 	bytes, err := json.Marshal(settings)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshalling settings: %w", err)
// 	}

// 	var cfg v0_1_0.ExampleConfigSettings
// 	err = json.Unmarshal(bytes, &cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("error unmarshalling settings: %w", err)
// 	}
// 	return &cfg, nil
// }

// Deserialize the settings for the latest configuration
// func deserializeSettings_Latest(settings map[string]any) (*config.ExampleConfigSettings, error) {
// 	bytes, err := json.Marshal(settings)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshalling settings: %w", err)
// 	}

// 	var cfg config.ExampleConfigSettings
// 	err = json.Unmarshal(bytes, &cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("error unmarshalling settings: %w", err)
// 	}
// 	return &cfg, nil
// }
