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
	TokenName          string               `json:"tokenName"`
	TokenSymbol        string               `json:"tokenSymbol"`
	DistributableToken *big.Int             `json:"distributableToken"`
	DistributableEth   *big.Int             `json:"distributableEth"`
	WithdrawableToken  *big.Int             `json:"withdrawableToken"`
	WithdrawableEth    *big.Int             `json:"withdrawableEth"`
	TxInfo             *eth.TransactionInfo `json:"txInfo"`
}
