package swapi

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type ValidatorExitInfo struct {
	Pubkey    beacon.ValidatorPubkey    `json:"pubkey"`
	Index     uint64                    `json:"index"`
	Signature beacon.ValidatorSignature `json:"signature"`
}

type ValidatorExitData struct {
	Epoch     uint64              `json:"epoch"`
	ExitInfos []ValidatorExitInfo `json:"exitInfos"`
}

type ValidatorInfo struct {
	Pubkey         beacon.ValidatorPubkey `json:"pubkey"`
	HasBeaconIndex bool                   `json:"hasBeaconIndex"`
	Index          string                 `json:"index"`
	Balance        uint64                 `json:"balance"`
	State          beacon.ValidatorState  `json:"state"`
}

type VaultInfo struct {
	Name          string           `json:"name"`
	Address       common.Address   `json:"address"`
	HasPermission bool             `json:"hasPermission"`
	Validators    []*ValidatorInfo `json:"validators"`
}

type ValidatorStatusData struct {
	NotRegisteredWithNodeSet bool         `json:"notRegisteredWithNodeSet"`
	InvalidPermissions       bool         `json:"invalidPermissions"`
	Vaults                   []*VaultInfo `json:"vaults"`
}
