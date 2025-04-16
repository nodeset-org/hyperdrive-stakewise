package swclient

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type ValidatorRequester struct {
	context client.IRequesterContext
}

func NewValidatorRequester(context client.IRequesterContext) *ValidatorRequester {
	return &ValidatorRequester{
		context: context,
	}
}

func (r *ValidatorRequester) GetName() string {
	return "Validator"
}
func (r *ValidatorRequester) GetRoute() string {
	return "validator"
}
func (r *ValidatorRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Exit the provided validators from the Beacon Chain (or simply return their signed exit messages for later use without broadcasting),
// with an optional epoch parameter. If not specified, the epoch from the current chain head will be used.
func (r *ValidatorRequester) Exit(pubkeys []beacon.ValidatorPubkey, epoch *uint64, noBroadcastBool bool) (*types.ApiResponse[swapi.ValidatorExitData], error) {
	args := map[string]string{
		"pubkeys":      client.MakeBatchArg(pubkeys),
		"no-broadcast": strconv.FormatBool(noBroadcastBool),
	}
	if epoch != nil {
		args["epoch"] = strconv.FormatUint(*epoch, 10)
	}
	return client.SendGetRequest[swapi.ValidatorExitData](r, "exit", "Exit", args)
}

// Get the status on Beacon for all of the validator keys that have been registered with StakeWise.
// If vault is provided, only the keys in that vault will be returned.
// Otherwise the keys for all vaults will be returned.
func (r *ValidatorRequester) Status(vault *common.Address) (*types.ApiResponse[swapi.ValidatorStatusData], error) {
	args := map[string]string{}
	if vault != nil {
		args["vault"] = vault.Hex()
	}
	return client.SendGetRequest[swapi.ValidatorStatusData](r, "status", "Status", args)
}
