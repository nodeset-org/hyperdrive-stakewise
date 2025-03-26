package swapi

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

type NodeSetValidatorStatus struct {
	Registered          bool `json:"registered"`
	ExitMessageUploaded bool `json:"exitMessageUploaded"`
}

type ValidatorStateInfo struct {
	Pubkey       beacon.ValidatorPubkey `json:"pubkey"`
	Index        string                 `json:"index"`
	BeaconStatus beacon.ValidatorState  `json:"beaconStatus"`
	NodeSet      NodeSetValidatorStatus `json:"nodeSet"`
}

type ValidatorStatusData struct {
	States []ValidatorStateInfo `json:"states"`
}
