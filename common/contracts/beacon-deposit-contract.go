package swcontracts

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	beaconDepositAbiString string = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes","name":"pubkey","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"withdrawal_credentials","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"amount","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"signature","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"index","type":"bytes"}],"name":"DepositEvent","type":"event"},{"inputs":[{"internalType":"bytes","name":"pubkey","type":"bytes"},{"internalType":"bytes","name":"withdrawal_credentials","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"},{"internalType":"bytes32","name":"deposit_data_root","type":"bytes32"}],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"get_deposit_count","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"get_deposit_root","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes4","name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"pure","type":"function"}]`

	depositEventName string = "DepositEvent"
)

// Cache
var depositContractAbi abi.ABI
var depositContractOnce sync.Once

type BeaconDepositContract struct {
	contract *eth.Contract
	ec       eth.IExecutionClient
	txMgr    *eth.TransactionManager
}

// Create a new contract instance
func NewBeaconDepositContract(address common.Address, ec eth.IExecutionClient, txMgr *eth.TransactionManager) (*BeaconDepositContract, error) {
	// Parse the ABI
	var err error
	depositContractOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(beaconDepositAbiString))
		if err == nil {
			depositContractAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing deposit contract ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, depositContractAbi, ec, ec, ec),
		Address:      address,
		ABI:          &depositContractAbi,
	}

	return &BeaconDepositContract{
		contract: contract,
		ec:       ec,
		txMgr:    txMgr,
	}, nil
}

// =============
// === Calls ===
// =============

func (c *BeaconDepositContract) GetDepositCount(mc *batch.MultiCaller) func() uint64 {
	var buffer []byte
	eth.AddCallToMulticaller(mc, c.contract, &buffer, "get_deposit_count")
	return func() uint64 {
		return binary.LittleEndian.Uint64(buffer)
	}
}

func (c *BeaconDepositContract) GetDepositRoot(mc *batch.MultiCaller, out *common.Hash) {
	eth.AddCallToMulticaller(mc, c.contract, out, "get_deposit_root")
}

// ====================
// === Transactions ===
// ====================

// Creates a deposit to the Beacon deposit contract.
// Note that the amount deposited must be set in opts.Value, and must match the deposit amount provided in the deposit data.
func (c *BeaconDepositContract) Deposit(pubkey beacon.ValidatorPubkey, withdrawalCredentials common.Hash, signature beacon.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "deposit", opts, pubkey[:], withdrawalCredentials[:], signature[:], depositDataRoot)
}

// ==============
// === Events ===
// ==============

// depositEventImpl represents a DepositEvent event raised by the BeaconDeposit contract.
type depositEventImpl struct {
	Pubkey                []byte `abi:"pubkey"`
	WithdrawalCredentials []byte `abi:"withdrawal_credentials"`
	Amount                []byte `abi:"amount"`
	Signature             []byte `abi:"signature"`
	Index                 []byte `abi:"index"`
}

// Deposit event
type DepositEvent struct {
	Pubkey                beacon.ValidatorPubkey    `json:"pubkey"`
	WithdrawalCredentials common.Hash               `json:"withdrawalCredentials"`
	Amount                uint64                    `json:"amount"`
	Signature             beacon.ValidatorSignature `json:"signature"`
	TxHash                common.Hash               `json:"txHash"`
	BlockNumber           uint64                    `json:"blockNumber"`
	TxIndex               uint                      `json:"txIndex"`
}

// Get the deposit events for the provided block range. If pubkeys is provided, only deposits for those pubkeys will be returned.
func (c *BeaconDepositContract) DepositEventsForPubkeys(pubkeys []beacon.ValidatorPubkey, startBlock *big.Int, endBlock *big.Int, intervalSize *big.Int) (map[beacon.ValidatorPubkey][]DepositEvent, error) {
	// Create the initial map and pubkey lookup
	requestedPubkeys := make(map[beacon.ValidatorPubkey]struct{}, len(pubkeys))
	for _, pubkey := range pubkeys {
		requestedPubkeys[pubkey] = struct{}{}
	}
	depositMap := make(map[beacon.ValidatorPubkey][]DepositEvent, len(pubkeys))

	// Get the deposit events
	addressFilter := []common.Address{c.contract.Address}
	topicFilter := [][]common.Hash{{c.contract.ABI.Events[depositEventName].ID}}
	logs, err := eth.GetLogs(c.ec, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return nil, err
	}

	// Process each event
	for _, log := range logs {
		depositEvent := new(depositEventImpl)
		err = c.contract.ContractImpl.UnpackLog(depositEvent, depositEventName, log)
		if err != nil {
			return nil, err
		}

		// Check if this is a deposit for one of the pubkeys we're looking for
		pubkey := beacon.ValidatorPubkey(depositEvent.Pubkey)
		if len(requestedPubkeys) > 0 {
			_, exists := requestedPubkeys[pubkey]
			if !exists {
				continue
			}
		}

		// Convert the deposit amount from little-endian binary to a uint64
		amount := binary.LittleEndian.Uint64(depositEvent.Amount)

		// Create the deposit data wrapper and add it to this pubkey's collection
		depositData := DepositEvent{
			Pubkey:                pubkey,
			WithdrawalCredentials: common.BytesToHash(depositEvent.WithdrawalCredentials),
			Amount:                amount,
			Signature:             beacon.ValidatorSignature(depositEvent.Signature),
			TxHash:                log.TxHash,
			BlockNumber:           log.BlockNumber,
			TxIndex:               log.TxIndex,
		}
		depositMap[pubkey] = append(depositMap[pubkey], depositData)
	}

	// Sort deposits by time
	for _, deposits := range depositMap {
		if len(deposits) > 1 {
			sortDepositData(deposits)
		}
	}

	return depositMap, nil
}

// Get the deposit events for the provided block range.
func (c *BeaconDepositContract) DepositEvent(startBlock *big.Int, endBlock *big.Int, intervalSize *big.Int) ([]DepositEvent, error) {
	// Get the deposit events
	addressFilter := []common.Address{c.contract.Address}
	topicFilter := [][]common.Hash{{c.contract.ABI.Events[depositEventName].ID}}
	logs, err := eth.GetLogs(c.ec, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return nil, err
	}

	// Process each event
	events := make([]DepositEvent, 0, len(logs))
	for _, log := range logs {
		depositEvent := new(depositEventImpl)
		err = c.contract.ContractImpl.UnpackLog(depositEvent, depositEventName, log)
		if err != nil {
			return nil, err
		}

		// Convert the deposit amount from little-endian binary to a uint64
		amount := binary.LittleEndian.Uint64(depositEvent.Amount)

		// Create the deposit data wrapper and add it to this pubkey's collection
		event := DepositEvent{
			Pubkey:                beacon.ValidatorPubkey(depositEvent.Pubkey),
			WithdrawalCredentials: common.BytesToHash(depositEvent.WithdrawalCredentials),
			Amount:                amount,
			Signature:             beacon.ValidatorSignature(depositEvent.Signature),
			TxHash:                log.TxHash,
			BlockNumber:           log.BlockNumber,
			TxIndex:               log.TxIndex,
		}
		events = append(events, event)
	}

	// Sort deposits by time
	sortDepositData(events)
	return events, nil
}

// Sorts a slice of deposit data entries - lower blocks come first, and if multiple transactions occur
// in the same block, lower transaction indices come first
func sortDepositData(data []DepositEvent) {
	sort.Slice(data, func(i int, j int) bool {
		first := data[i]
		second := data[j]
		if first.BlockNumber == second.BlockNumber {
			return first.TxIndex < second.TxIndex
		}
		return first.BlockNumber < second.BlockNumber
	})
}
