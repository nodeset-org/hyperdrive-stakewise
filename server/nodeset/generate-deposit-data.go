package swnodeset

import (
	"errors"
	"fmt"
	"net/url"

	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/utils/input"

	duserver "github.com/nodeset-org/hyperdrive-daemon/module-utils/server"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// ===============
// === Factory ===
// ===============

type nodesetGenerateDepositDataContextFactory struct {
	handler *NodesetHandler
}

func (f *nodesetGenerateDepositDataContextFactory) Create(args url.Values) (*nodesetGenerateDepositDataContext, error) {
	c := &nodesetGenerateDepositDataContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateOptionalArgBatch("pubkeys", args, 0, input.ValidatePubkey, &c.pubkeys, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodesetGenerateDepositDataContextFactory) RegisterRoute(router *mux.Router) {
	duserver.RegisterQuerylessGet[*nodesetGenerateDepositDataContext, swapi.NodesetGenerateDepositDataData](
		router, "generate-deposit-data", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============
type nodesetGenerateDepositDataContext struct {
	handler *NodesetHandler
	pubkeys []beacon.ValidatorPubkey
}

func (c *nodesetGenerateDepositDataContext) PrepareData(data *swapi.NodesetGenerateDepositDataData, walletStatus wallet.WalletStatus, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ddMgr := sp.GetDepositDataManager()
	w := sp.GetWallet()

	// Fetch private keys and derive public keys
	privateKeys, err := w.GetAllPrivateKeys()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting private keys: %w", err)
	}

	// Get the keys to generate deposit data for
	keysToGenerate := []*eth2types.BLSPrivateKey{}
	pubkeyMap := map[beacon.ValidatorPubkey]bool{}
	if len(c.pubkeys) == 0 {
		keysToGenerate = privateKeys
	} else {
		for _, pubkey := range c.pubkeys {
			pubkeyMap[pubkey] = true
		}

		// Make a list of the desired private keys
		for _, privateKey := range privateKeys {
			pubkey := beacon.ValidatorPubkey(privateKey.PublicKey().Marshal())
			if _, exists := pubkeyMap[pubkey]; !exists {
				continue
			}
			keysToGenerate = append(keysToGenerate, privateKey)
		}
	}

	// Generate deposit data
	depositData, err := ddMgr.GenerateDepositData(c.handler.logger.Logger, keysToGenerate)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error generating deposit data: %w", err)
	}
	data.Deposits = depositData
	return types.ResponseStatus_Success, nil
}
