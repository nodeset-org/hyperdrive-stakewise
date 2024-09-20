package swclient

import (
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
)

type ServiceRequester struct {
	context client.IRequesterContext
}

func NewServiceRequester(context client.IRequesterContext) *ServiceRequester {
	return &ServiceRequester{
		context: context,
	}
}

func (r *ServiceRequester) GetName() string {
	return "Service"
}
func (r *ServiceRequester) GetRoute() string {
	return "service"
}
func (r *ServiceRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Gets the resources for the daemon's selected network
func (r *ServiceRequester) GetResources() (*types.ApiResponse[swapi.ServiceGetResourcesData], error) {
	return client.SendGetRequest[swapi.ServiceGetResourcesData](r, "get-resources", "GetResources", nil)
}

// Gets the network settings for the daemon's selected network
func (r *ServiceRequester) GetNetworkSettings() (*types.ApiResponse[swapi.ServiceGetNetworkSettingsData], error) {
	return client.SendGetRequest[swapi.ServiceGetNetworkSettingsData](r, "get-network-settings", "GetNetworkSettings", nil)
}

// Gets the version of the daemon
func (r *ServiceRequester) Version() (*types.ApiResponse[swapi.ServiceVersionData], error) {
	return client.SendGetRequest[swapi.ServiceVersionData](r, "version", "Version", nil)
}
