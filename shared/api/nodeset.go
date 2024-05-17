package swapi

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

type NodesetUploadDepositDataData struct {
	SufficientBalance    bool                     `json:"sufficientBalance"`
	Balance              float64                  `json:"balance"`
	ServerResponse       []byte                   `json:"serverResponse"`
	UnregisteredPubkeys  []beacon.ValidatorPubkey `json:"newPubkeys"`
	TotalCount           uint64                   `json:"totalCount"`
	EthPerKey            float64                  `json:"ethPerKey"`
	RemainingEthRequired float64                  `json:"remainingEthRequired"`
}
