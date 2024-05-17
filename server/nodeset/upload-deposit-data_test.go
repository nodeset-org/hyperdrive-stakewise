package swnodeset

import (
	"errors"
	"testing"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/stretchr/testify/assert"
)

const (
	walletAddressHex string = "0x7e5700eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	pubkey0Hex       string = "beac0900bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	pubkey1Hex       string = "beac0901bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	pubkey2Hex       string = "beac0902bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestUploadDepositData_3Keys(t *testing.T) {
	// Set the initial wallet balance
	balance := eth.EthToWei(validatorDepositCost * 1.5)

	pubkey0, err0 := beacon.HexToValidatorPubkey(pubkey0Hex)
	pubkey1, err1 := beacon.HexToValidatorPubkey(pubkey1Hex)
	pubkey2, err2 := beacon.HexToValidatorPubkey(pubkey2Hex)
	if err := errors.Join(err0, err1, err2); err != nil {
		t.Fatalf("error parsing pubkeys: %s", err.Error())
	}

	pendingKeys := []beacon.ValidatorPubkey{}
	unregisterKeys := []beacon.ValidatorPubkey{
		pubkey0,
		pubkey1,
		pubkey2,
	}

	registeredKeys, remainingKeys, remainingBalance := getRegisterableKeys(pendingKeys, unregisterKeys, balance)
	assert.Equal(t, 1, len(registeredKeys))
	assert.Equal(t, pubkey0, registeredKeys[0])
	assert.Equal(t, 2, len(remainingKeys))
	assert.Equal(t, pubkey1, remainingKeys[0])
	assert.Equal(t, pubkey2, remainingKeys[1])
	expectedRemainingBalance := eth.EthToWei(validatorDepositCost * 0.5)
	assert.Equal(t, remainingBalance.Cmp(expectedRemainingBalance), 0)
}

func TestUploadDepositData_1Key(t *testing.T) {
	// Set the initial wallet balance
	balance := eth.EthToWei(validatorDepositCost * 1.5)

	pubkey0, err := beacon.HexToValidatorPubkey(pubkey0Hex)
	if err != nil {
		t.Fatalf("error parsing pubkeys: %s", err.Error())
	}

	pendingKeys := []beacon.ValidatorPubkey{}
	unregisterKeys := []beacon.ValidatorPubkey{
		pubkey0,
	}

	registeredKeys, remainingKeys, remainingBalance := getRegisterableKeys(pendingKeys, unregisterKeys, balance)
	assert.Equal(t, 1, len(registeredKeys))
	assert.Equal(t, pubkey0, registeredKeys[0])
	assert.Equal(t, 0, len(remainingKeys))
	expectedRemainingBalance := eth.EthToWei(validatorDepositCost * 0.5)
	assert.Equal(t, remainingBalance.Cmp(expectedRemainingBalance), 0)
}

func TestUploadDepositData_PendingKeys(t *testing.T) {
	// Set the initial wallet balance
	balance := eth.EthToWei(validatorDepositCost * 1.5)

	pubkey0, err0 := beacon.HexToValidatorPubkey(pubkey0Hex)
	pubkey1, err1 := beacon.HexToValidatorPubkey(pubkey1Hex)
	pubkey2, err2 := beacon.HexToValidatorPubkey(pubkey2Hex)
	if err := errors.Join(err0, err1, err2); err != nil {
		t.Fatalf("error parsing pubkeys: %s", err.Error())
	}

	pendingKeys := []beacon.ValidatorPubkey{
		pubkey0,
	}
	unregisterKeys := []beacon.ValidatorPubkey{
		pubkey1,
		pubkey2,
	}

	registeredKeys, remainingKeys, remainingBalance := getRegisterableKeys(pendingKeys, unregisterKeys, balance)
	assert.Equal(t, 0, len(registeredKeys))
	assert.Equal(t, 2, len(remainingKeys))
	assert.Equal(t, pubkey1, remainingKeys[0])
	assert.Equal(t, pubkey2, remainingKeys[1])
	expectedRemainingBalance := eth.EthToWei(validatorDepositCost * 0.5)
	assert.Equal(t, remainingBalance.Cmp(expectedRemainingBalance), 0)
}
