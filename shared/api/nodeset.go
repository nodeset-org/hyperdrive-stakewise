package swapi

import (
	"github.com/nodeset-org/hyperdrive-daemon/shared/types"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type NodesetRegistrationStatus string

const (
	NodesetRegistrationStatus_Registered   NodesetRegistrationStatus = "registered"
	NodesetRegistrationStatus_Unregistered NodesetRegistrationStatus = "unregistered"
	NodesetRegistrationStatus_Unknown      NodesetRegistrationStatus = "unknown"
	NodesetRegistrationStatus_NoWallet     NodesetRegistrationStatus = "no-wallet"
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

type NodeSetRegisterNodeData struct {
	Success           bool `json:"success"`
	AlreadyRegistered bool `json:"alreadyRegistered"`
	NotWhitelisted    bool `json:"notWhitelisted"`
}

type NodeSetRegistrationStatusData struct {
	Status       NodesetRegistrationStatus `json:"status"`
	ErrorMessage string                    `json:"errorMessage"`
}

type NodesetGenerateDepositDataData struct {
	Deposits []*types.ExtendedDepositData `json:"deposits"`
}
