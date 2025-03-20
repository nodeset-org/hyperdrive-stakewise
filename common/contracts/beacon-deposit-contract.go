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
	"github.com/ethereum/go-ethereum/core/types"
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
}

// Create a new contract instance
func NewBeaconDepositContract(address common.Address, ec eth.IExecutionClient) (*BeaconDepositContract, error) {
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

// ==============
// === Events ===
// ==============

// BeaconDepositEvent represents a DepositEvent event raised by the BeaconDeposit contract.
type BeaconDepositEvent struct {
	Pubkey                []byte    `abi:"pubkey"`
	WithdrawalCredentials []byte    `abi:"withdrawal_credentials"`
	Amount                []byte    `abi:"amount"`
	Signature             []byte    `abi:"signature"`
	Index                 []byte    `abi:"index"`
	Raw                   types.Log // Raw event data
}

// Deposit Data event
type DepositData struct {
	Pubkey                beacon.ValidatorPubkey    `json:"pubkey"`
	WithdrawalCredentials common.Hash               `json:"withdrawalCredentials"`
	Amount                uint64                    `json:"amount"`
	Signature             beacon.ValidatorSignature `json:"signature"`
	TxHash                common.Hash               `json:"txHash"`
	BlockNumber           uint64                    `json:"blockNumber"`
	TxIndex               uint                      `json:"txIndex"`
}

// Get the deposit events for the provided block range. If pubkeys is provided, only deposits for those pubkeys will be returned.
func (c *BeaconDepositContract) DepositEvents(pubkeys []beacon.ValidatorPubkey, startBlock *big.Int, intervalSize *big.Int) (map[beacon.ValidatorPubkey][]DepositData, error) {
	// Create the initial map and pubkey lookup
	requestedPubkeys := make(map[beacon.ValidatorPubkey]struct{}, len(pubkeys))
	for _, pubkey := range pubkeys {
		requestedPubkeys[pubkey] = struct{}{}
	}
	depositMap := make(map[beacon.ValidatorPubkey][]DepositData, len(pubkeys))

	// Get the deposit events
	addressFilter := []common.Address{c.contract.Address}
	topicFilter := [][]common.Hash{{c.contract.ABI.Events[depositEventName].ID}}
	logs, err := eth.GetLogs(c.ec, addressFilter, topicFilter, intervalSize, startBlock, nil, nil)
	if err != nil {
		return nil, err
	}

	// Process each event
	for _, log := range logs {
		depositEvent := new(BeaconDepositEvent)
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
		depositData := DepositData{
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

// Sorts a slice of deposit data entries - lower blocks come first, and if multiple transactions occur
// in the same block, lower transaction indices come first
func sortDepositData(data []DepositData) {
	sort.Slice(data, func(i int, j int) bool {
		first := data[i]
		second := data[j]
		if first.BlockNumber == second.BlockNumber {
			return first.TxIndex < second.TxIndex
		}
		return first.BlockNumber < second.BlockNumber
	})
}
