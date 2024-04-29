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
	splitMainAbiString_1_1 string = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"Create2Error","type":"error"},{"inputs":[],"name":"CreateError","type":"error"},{"inputs":[{"internalType":"address","name":"newController","type":"address"}],"name":"InvalidNewController","type":"error"},{"inputs":[{"internalType":"uint256","name":"accountsLength","type":"uint256"},{"internalType":"uint256","name":"allocationsLength","type":"uint256"}],"name":"InvalidSplit__AccountsAndAllocationsMismatch","type":"error"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"InvalidSplit__AccountsOutOfOrder","type":"error"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"InvalidSplit__AllocationMustBePositive","type":"error"},{"inputs":[{"internalType":"uint32","name":"allocationsSum","type":"uint32"}],"name":"InvalidSplit__InvalidAllocationsSum","type":"error"},{"inputs":[{"internalType":"uint32","name":"distributorFee","type":"uint32"}],"name":"InvalidSplit__InvalidDistributorFee","type":"error"},{"inputs":[{"internalType":"bytes32","name":"hash","type":"bytes32"}],"name":"InvalidSplit__InvalidHash","type":"error"},{"inputs":[{"internalType":"uint256","name":"accountsLength","type":"uint256"}],"name":"InvalidSplit__TooFewAccounts","type":"error"},{"inputs":[{"internalType":"address","name":"sender","type":"address"}],"name":"Unauthorized","type":"error"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"}],"name":"CancelControlTransfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":true,"internalType":"address","name":"previousController","type":"address"},{"indexed":true,"internalType":"address","name":"newController","type":"address"}],"name":"ControlTransfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":false,"internalType":"address[]","name":"accounts","type":"address[]"},{"indexed":false,"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"indexed":false,"internalType":"uint32","name":"distributorFee","type":"uint32"},{"indexed":false,"internalType":"address","name":"controller","type":"address"}],"name":"CreateSplit","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":true,"internalType":"contract ERC20","name":"token","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"address","name":"distributorAddress","type":"address"}],"name":"DistributeERC20","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"address","name":"distributorAddress","type":"address"}],"name":"DistributeETH","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":true,"internalType":"address","name":"newPotentialController","type":"address"}],"name":"InitiateControlTransfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"split","type":"address"},{"indexed":false,"internalType":"address[]","name":"accounts","type":"address[]"},{"indexed":false,"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"indexed":false,"internalType":"uint32","name":"distributorFee","type":"uint32"}],"name":"UpdateSplit","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"uint256","name":"ethAmount","type":"uint256"},{"indexed":false,"internalType":"contract ERC20[]","name":"tokens","type":"address[]"},{"indexed":false,"internalType":"uint256[]","name":"tokenAmounts","type":"uint256[]"}],"name":"Withdrawal","type":"event"},{"inputs":[],"name":"PERCENTAGE_SCALE","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"acceptControl","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"cancelControlTransfer","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"},{"internalType":"address","name":"controller","type":"address"}],"name":"createSplit","outputs":[{"internalType":"address","name":"split","type":"address"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"contract ERC20","name":"token","type":"address"},{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"},{"internalType":"address","name":"distributorAddress","type":"address"}],"name":"distributeERC20","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"},{"internalType":"address","name":"distributorAddress","type":"address"}],"name":"distributeETH","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"getController","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"contract ERC20","name":"token","type":"address"}],"name":"getERC20Balance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"getETHBalance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"getHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"getNewPotentialController","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"}],"name":"makeSplitImmutable","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"}],"name":"predictImmutableSplitAddress","outputs":[{"internalType":"address","name":"split","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"address","name":"newController","type":"address"}],"name":"transferControl","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"contract ERC20","name":"token","type":"address"},{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"},{"internalType":"address","name":"distributorAddress","type":"address"}],"name":"updateAndDistributeERC20","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"},{"internalType":"address","name":"distributorAddress","type":"address"}],"name":"updateAndDistributeETH","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"split","type":"address"},{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"uint32[]","name":"percentAllocations","type":"uint32[]"},{"internalType":"uint32","name":"distributorFee","type":"uint32"}],"name":"updateSplit","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"walletImplementation","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"withdrawETH","type":"uint256"},{"internalType":"contract ERC20[]","name":"tokens","type":"address[]"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
)

// ABI cache
var splitMainAbi_1_1 abi.ABI
var splitMainOnce_1_1 sync.Once

// Binding for the SplitMain "v1.1" contract
// See https://github.com/0xSplits/splits-contracts/blob/holesky-2/contracts/interfaces/ISplitMain.sol
type SplitMain1_1 struct {
	contract *eth.Contract
	txMgr    *eth.TransactionManager
}

// Create a new SplitMain instance
func NewSplitMain1_1(address common.Address, ec eth.IExecutionClient, txMgr *eth.TransactionManager) (*SplitMain1_1, error) {
	// Parse the ABI
	var err error
	splitMainOnce_1_1.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(splitMainAbiString_1_1))
		if err == nil {
			splitMainAbi_1_1 = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing Stakewise Vault ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, splitMainAbi_1_1, ec, ec, ec),
		Address:      address,
		ABI:          &splitMainAbi_1_1,
	}

	return &SplitMain1_1{
		contract: contract,
		txMgr:    txMgr,
	}, nil
}

// =============
// === Calls ===
// =============

func (c *SplitMain1_1) GetEthBalance(mc *batch.MultiCaller, out **big.Int, account common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "getETHBalance", account)
}

func (c *SplitMain1_1) GetErc20Balance(mc *batch.MultiCaller, out **big.Int, account common.Address, token common.Address) {
	eth.AddCallToMulticaller(mc, c.contract, out, "getERC20Balance", account, token)
}

// ====================
// === Transactions ===
// ====================

func (c *SplitMain1_1) Withdraw(address common.Address, ethAmountWithdraw *big.Int, claimTokenList []common.Address, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "withdraw", opts, address, ethAmountWithdraw, claimTokenList)
}