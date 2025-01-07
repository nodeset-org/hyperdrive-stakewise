package config

import "github.com/rocket-pool/node-manager-core/config"

// Network settings with a field for Hyperdrive-specific settings
type HyperdriveSettings struct {
	*config.NetworkSettings `yaml:",inline"`

	// Hyperdrive resources for the network
	HyperdriveResources *HyperdriveResources `yaml:"hyperdriveResources" json:"hyperdriveResources"`
}

// A collection of network-specific resources and getters for them
type HyperdriveResources struct {
	// The URL for the NodeSet API server
	NodeSetApiUrl string `yaml:"nodeSetApiUrl" json:"nodeSetApiUrl"`

	// The pubkey used to encrypt messages to nodeset.io
	EncryptionPubkey string `yaml:"encryptionPubkey" json:"encryptionPubkey"`
}

// An aggregated collection of resources for the selected network, including Hyperdrive resources
type MergedResources struct {
	// Base network resources
	*config.NetworkResources

	// Hyperdrive resources
	*HyperdriveResources
}
