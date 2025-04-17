package swclient

import (
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
)

type NetworkRequester struct {
	context client.IRequesterContext
}

func NewNetworkRequester(context client.IRequesterContext) *NetworkRequester {
	return &NetworkRequester{
		context: context,
	}
}

func (r *NetworkRequester) GetName() string {
	return "Network"
}
func (r *NetworkRequester) GetRoute() string {
	return "network"
}
func (r *NetworkRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Get the NodeSet.io and StakeWise status for all vaults in this deployment
func (r *NetworkRequester) Status() (*types.ApiResponse[swapi.NetworkStatusData], error) {
	args := map[string]string{}
	return client.SendGetRequest[swapi.NetworkStatusData](r, "status", "Status", args)
}
