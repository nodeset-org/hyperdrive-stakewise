package config

import (
	"github.com/rocket-pool/node-manager-core/config"
)

// Constants
const (
// mevBoostTag string = "flashbots/mev-boost:1.8.1"
)

type MevSelectionMode string

type MevRelayID string

// A MEV relay
type MevRelay struct {
	ID          MevRelayID
	Name        string
	Description string
	Urls        map[string]string
}

// Configuration for MEV-Boost
type MevBoostConfig struct {
	// Toggle to enable / disable
	Enable config.Parameter[bool]

	// Ownership mode
	Mode config.Parameter[config.ClientMode]

	// The mode for relay selection
	SelectionMode config.Parameter[MevSelectionMode]

	// Flashbots relay
	FlashbotsRelay config.Parameter[bool]

	// bloXroute max profit relay
	BloxRouteMaxProfitRelay config.Parameter[bool]

	// bloXroute regulated relay
	BloxRouteRegulatedRelay config.Parameter[bool]

	// Titan regional relay
	TitanRegionalRelay config.Parameter[bool]

	// Custom relays provided by the user
	CustomRelays config.Parameter[string]

	// The RPC port
	Port config.Parameter[uint16]

	// Toggle for forwarding the HTTP port outside of Docker
	OpenRpcPort config.Parameter[config.RpcPortMode]

	// The Docker Hub tag for MEV-Boost
	ContainerTag config.Parameter[string]

	// Custom command line flags
	AdditionalFlags config.Parameter[string]

	// The URL of an external MEV-Boost client
	ExternalUrl config.Parameter[string]

	///////////////////////////
	// Non-editable settings //
	///////////////////////////

	// parent   *HyperdriveConfig
	// relays   []MevRelay
	// relayMap map[MevRelayID]MevRelay
}
