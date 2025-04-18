package swapi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
)

type WalletInitializeData struct {
	AccountAddress common.Address `json:"accountAddress"`
}

type WalletGenerateKeysData struct {
	Pubkeys []beacon.ValidatorPubkey `json:"pubkeys"`
}

type WalletClaimRewardsData struct {
	NativeToken             common.Address       `json:"nativeToken"`
	TokenName               string               `json:"tokenName"`
	TokenSymbol             string               `json:"tokenSymbol"`
	DistributableToken      *big.Int             `json:"distributableToken"`
	DistributableEth        *big.Int             `json:"distributableEth"`
	WithdrawableToken       *big.Int             `json:"withdrawableToken"`
	WithdrawableNativeToken *big.Int             `json:"withdrawableNativeToken"`
	TxInfo                  *eth.TransactionInfo `json:"txInfo"`
}

type WalletGetAvailableKeysData struct {
	SufficientBalance         bool                     `json:"sufficientBalance"`
	Balance                   float64                  `json:"balance"`
	AvailablePubkeys          []beacon.ValidatorPubkey `json:"availablePubkeys"`
	EthPerKey                 float64                  `json:"ethPerKey"`
	RemainingEthRequired      float64                  `json:"remainingEthRequired"`
	KeysMissingPrivateKey     []beacon.ValidatorPubkey `json:"keysMissingPrivateKey"`
	KeysRequiringLookbackScan []beacon.ValidatorPubkey `json:"keysRequiringLookbackScan"`
	KeysAlreadyOnBeacon       []beacon.ValidatorPubkey `json:"keysAlreadyOnBeacon"`
	KeysWithDepositEvents     []beacon.ValidatorPubkey `json:"keysWithDepositEvents"`
	KeysUsedWithDepositRoot   []beacon.ValidatorPubkey `json:"keysUsedWithDepositRoot"`
}

type WalletRecoverKeysBody struct {
	Pubkeys     []beacon.ValidatorPubkey `json:"pubkeys"`
	StartIndex  uint64                   `json:"startIndex"`
	Count       uint64                   `json:"count"`
	SearchLimit uint64                   `json:"searchLimit"`
	RestartVc   bool                     `json:"restartVc"`
}

type RecoveredKey struct {
	Pubkey beacon.ValidatorPubkey `json:"pubkey"`
	Index  uint64                 `json:"index"`
}

type WalletRecoverKeysData struct {
	NotRegisteredWithNodeSet bool           `json:"notRegisteredWithNodeSet"`
	Keys                     []RecoveredKey `json:"keys"`
	SearchEnd                uint64         `json:"searchEnd"`
}
