package swclient

import (
	"strconv"

	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
)

type WalletRequester struct {
	context client.IRequesterContext
}

func NewWalletRequester(context client.IRequesterContext) *WalletRequester {
	return &WalletRequester{
		context: context,
	}
}

func (r *WalletRequester) GetName() string {
	return "Wallet"
}
func (r *WalletRequester) GetRoute() string {
	return "wallet"
}
func (r *WalletRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Generate and save new validator keys
func (r *WalletRequester) GenerateKeys(count uint64, restartVc bool) (*types.ApiResponse[swapi.WalletGenerateKeysData], error) {
	args := map[string]string{
		"count":      strconv.FormatUint(count, 10),
		"restart-vc": strconv.FormatBool(restartVc),
	}
	return client.SendGetRequest[swapi.WalletGenerateKeysData](r, "generate-keys", "GenerateKeys", args)
}

// Export the wallet in encrypted ETH key format
func (r *WalletRequester) Initialize() (*types.ApiResponse[swapi.WalletInitializeData], error) {
	return client.SendGetRequest[swapi.WalletInitializeData](r, "initialize", "Initialize", nil)
}

// Get the keys that are available for new deposits
func (r *WalletRequester) GetAvailableKeys() (*types.ApiResponse[swapi.WalletGetAvailableKeysData], error) {
	return client.SendGetRequest[swapi.WalletGetAvailableKeysData](r, "get-available-keys", "GetAvailableKeys", nil)
}

// Get the keys that have been registered with NodeSet for each vault in the currently active deployment
func (r *WalletRequester) GetRegisteredKeys() (*types.ApiResponse[swapi.WalletRegisteredKeysData], error) {
	return client.SendGetRequest[swapi.WalletRegisteredKeysData](r, "registered-keys", "GetRegisteredKeys", nil)
}
