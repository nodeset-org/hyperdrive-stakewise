package swapi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type NetworkVaultInfo struct {
	Name                 string         `json:"name"`
	Address              common.Address `json:"address"`
	MaxValidators        int            `json:"maxValidators"`
	RegisteredValidators int            `json:"registeredValidators"`
	AvailableValidators  int            `json:"availableValidators"`
	Balance              *big.Int       `json:"balance"`
}

type NetworkStatusData struct {
	NotRegisteredWithNodeSet bool                `json:"notRegisteredWithNodeSet"`
	InvalidPermissions       bool                `json:"invalidPermissions"`
	Vaults                   []*NetworkVaultInfo `json:"vaults"`
}
