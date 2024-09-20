package swapi

import swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"

type ServiceGetResourcesData struct {
	Resources *swconfig.MergedResources `json:"resources"`
}

type ServiceGetNetworkSettingsData struct {
	Settings *swconfig.StakeWiseSettings `json:"settings"`
}

type ServiceGetConfigData struct {
	Config map[string]any `json:"config"`
}

type ServiceVersionData struct {
	Version string `json:"version"`
}
