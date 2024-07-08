package swapi

import (
	"github.com/rocket-pool/node-manager-core/beacon"
)

type NodesetUploadDepositDataData struct {
	UnregisteredNode             bool                     `json:"unregisteredNode"`
	InvalidWithdrawalCredentials bool                     `json:"invalidWithdrawalCredentials"`
	NotAuthorizedForMainnet      bool                     `json:"notAuthorizedForMainnet"`
	SufficientBalance            bool                     `json:"sufficientBalance"`
	Balance                      float64                  `json:"balance"`
	NewPubkeys                   []beacon.ValidatorPubkey `json:"newPubkeys"`
	RemainingPubkeys             []beacon.ValidatorPubkey `json:"remainingPubkeys"`
	TotalCount                   uint64                   `json:"totalCount"`
	ActiveCount                  uint64                   `json:"activeCount"`
	PendingCount                 uint64                   `json:"pendingCount"`
	EthPerKey                    float64                  `json:"ethPerKey"`
	RemainingEthRequired         float64                  `json:"remainingEthRequired"`
}

type NodesetGenerateDepositDataData struct {
	Deposits []beacon.ExtendedDepositData `json:"deposits"`
}
