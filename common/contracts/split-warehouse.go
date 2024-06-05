package swcontracts

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	splitWarehouseAbiString string = `[{"type":"constructor","inputs":[{"name":"_native_token_name","type":"string","internalType":"string"},{"name":"_native_token_symbol","type":"string","internalType":"string"}],"stateMutability":"nonpayable"},{"type":"function","name":"APPROVE_AND_CALL_TYPE_HASH","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"DOMAIN_SEPARATOR","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"NATIVE_TOKEN","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"NATIVE_TOKEN_ID","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"PERCENTAGE_SCALE","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"allowance","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"spender","type":"address","internalType":"address"},{"name":"tokenId","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"amount","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"approve","inputs":[{"name":"_spender","type":"address","internalType":"address"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"approveBySig","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_spender","type":"address","internalType":"address"},{"name":"_operator","type":"bool","internalType":"bool"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_nonce","type":"uint256","internalType":"uint256"},{"name":"_deadline","type":"uint48","internalType":"uint48"},{"name":"_signature","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"balanceOf","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"id","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"amount","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"batchDeposit","inputs":[{"name":"_receivers","type":"address[]","internalType":"address[]"},{"name":"_token","type":"address","internalType":"address"},{"name":"_amounts","type":"uint256[]","internalType":"uint256[]"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"batchTransfer","inputs":[{"name":"_receivers","type":"address[]","internalType":"address[]"},{"name":"_token","type":"address","internalType":"address"},{"name":"_amounts","type":"uint256[]","internalType":"uint256[]"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"decimals","inputs":[{"name":"id","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"uint8","internalType":"uint8"}],"stateMutability":"view"},{"type":"function","name":"deposit","inputs":[{"name":"_receiver","type":"address","internalType":"address"},{"name":"_token","type":"address","internalType":"address"},{"name":"_amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"eip712Domain","inputs":[],"outputs":[{"name":"fields","type":"bytes1","internalType":"bytes1"},{"name":"name","type":"string","internalType":"string"},{"name":"version","type":"string","internalType":"string"},{"name":"chainId","type":"uint256","internalType":"uint256"},{"name":"verifyingContract","type":"address","internalType":"address"},{"name":"salt","type":"bytes32","internalType":"bytes32"},{"name":"extensions","type":"uint256[]","internalType":"uint256[]"}],"stateMutability":"view"},{"type":"function","name":"invalidateNonce","inputs":[{"name":"_nonce","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"isOperator","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"operator","type":"address","internalType":"address"}],"outputs":[{"name":"approved","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"isValidNonce","inputs":[{"name":"_from","type":"address","internalType":"address"},{"name":"_nonce","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"name","inputs":[{"name":"id","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"nonceBitMap","inputs":[{"name":"account","type":"address","internalType":"address"},{"name":"word","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"bitMap","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"setOperator","inputs":[{"name":"_operator","type":"address","internalType":"address"},{"name":"_approved","type":"bool","internalType":"bool"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"setWithdrawConfig","inputs":[{"name":"_config","type":"tuple","internalType":"struct SplitsWarehouse.WithdrawConfig","components":[{"name":"incentive","type":"uint16","internalType":"uint16"},{"name":"paused","type":"bool","internalType":"bool"}]}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"supportsInterface","inputs":[{"name":"_interfaceId","type":"bytes4","internalType":"bytes4"}],"outputs":[{"name":"supported","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"symbol","inputs":[{"name":"id","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"temporaryApproveAndCall","inputs":[{"name":"_spender","type":"address","internalType":"address"},{"name":"_operator","type":"bool","internalType":"bool"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_target","type":"address","internalType":"address"},{"name":"_data","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"temporaryApproveAndCallBySig","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_spender","type":"address","internalType":"address"},{"name":"_operator","type":"bool","internalType":"bool"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_target","type":"address","internalType":"address"},{"name":"_data","type":"bytes","internalType":"bytes"},{"name":"_nonce","type":"uint256","internalType":"uint256"},{"name":"_deadline","type":"uint48","internalType":"uint48"},{"name":"_signature","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transfer","inputs":[{"name":"_receiver","type":"address","internalType":"address"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transferFrom","inputs":[{"name":"_sender","type":"address","internalType":"address"},{"name":"_receiver","type":"address","internalType":"address"},{"name":"_id","type":"uint256","internalType":"uint256"},{"name":"_amount","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"withdraw","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_tokens","type":"address[]","internalType":"address[]"},{"name":"_amounts","type":"uint256[]","internalType":"uint256[]"},{"name":"_withdrawer","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdraw","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_token","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawConfig","inputs":[{"name":"owner","type":"address","internalType":"address"}],"outputs":[{"name":"incentive","type":"uint16","internalType":"uint16"},{"name":"paused","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"event","name":"Approval","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"spender","type":"address","indexed":true,"internalType":"address"},{"name":"id","type":"uint256","indexed":true,"internalType":"uint256"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"EIP712DomainChanged","inputs":[],"anonymous":false},{"type":"event","name":"NonceInvalidation","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"nonce","type":"uint256","indexed":true,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"OperatorSet","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"spender","type":"address","indexed":true,"internalType":"address"},{"name":"approved","type":"bool","indexed":false,"internalType":"bool"}],"anonymous":false},{"type":"event","name":"Transfer","inputs":[{"name":"caller","type":"address","indexed":false,"internalType":"address"},{"name":"sender","type":"address","indexed":true,"internalType":"address"},{"name":"receiver","type":"address","indexed":true,"internalType":"address"},{"name":"id","type":"uint256","indexed":true,"internalType":"uint256"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"Withdraw","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"token","type":"address","indexed":true,"internalType":"address"},{"name":"withdrawer","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"reward","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"WithdrawConfigUpdated","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"config","type":"tuple","indexed":false,"internalType":"struct SplitsWarehouse.WithdrawConfig","components":[{"name":"incentive","type":"uint16","internalType":"uint16"},{"name":"paused","type":"bool","internalType":"bool"}]}],"anonymous":false},{"type":"error","name":"ExpiredSignature","inputs":[{"name":"deadline","type":"uint48","internalType":"uint48"}]},{"type":"error","name":"InvalidAck","inputs":[]},{"type":"error","name":"InvalidAmount","inputs":[]},{"type":"error","name":"InvalidNonce","inputs":[]},{"type":"error","name":"InvalidPermitParams","inputs":[]},{"type":"error","name":"InvalidShortString","inputs":[]},{"type":"error","name":"InvalidSigner","inputs":[]},{"type":"error","name":"LengthMismatch","inputs":[]},{"type":"error","name":"Overflow","inputs":[]},{"type":"error","name":"StringTooLong","inputs":[{"name":"str","type":"string","internalType":"string"}]},{"type":"error","name":"WithdrawalPaused","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"ZeroOwner","inputs":[]}]`
)

// ABI cache
var splitWarehouseAbi abi.ABI
var splitWarehouseOnce sync.Once

type SplitWarehouse struct {
	contract *eth.Contract
	txMgr    *eth.TransactionManager
}

// Create a new SplitWarehouse instance
func NewSplitWarehouse(address common.Address, ec eth.IExecutionClient, txMgr *eth.TransactionManager) (*SplitWarehouse, error) {
	// Parse the ABI
	var err error
	splitWarehouseOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(splitWarehouseAbiString))
		if err == nil {
			splitWarehouseAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing SplitWarehouse ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, splitWarehouseAbi, ec, ec, ec),
		Address:      address,
		ABI:          &splitWarehouseAbi,
	}

	return &SplitWarehouse{
		contract: contract,
		txMgr:    txMgr,
	}, nil
}

// =============
// === Calls ===
// =============

func (c *SplitWarehouse) BalanceOf(mc *batch.MultiCaller, out **big.Int, account common.Address, token common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "balanceOf", account, ToUint256(token))
}

func (c *SplitWarehouse) NativeToken(mc *batch.MultiCaller, out *common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "NATIVE_TOKEN")
}

// ====================
// === Transactions ===
// ====================

func (c *SplitWarehouse) Withdraw(address common.Address, claimTokenList []common.Address, amountWithdraw []*big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "withdraw", opts, address, claimTokenList, amountWithdraw, address)
}

func ToUint256(address common.Address) *big.Int {
	return new(big.Int).SetBytes(address.Bytes())
}
