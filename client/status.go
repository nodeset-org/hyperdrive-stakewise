package swclient

import (
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
)

type StatusRequester struct {
	context client.IRequesterContext
}

func NewStatusRequester(context client.IRequesterContext) *StatusRequester {
	return &StatusRequester{
		context: context,
	}
}

func (r *StatusRequester) GetName() string {
	return "Status"
}

func (r *StatusRequester) GetRoute() string {
	return "status"
}

func (r *StatusRequester) GetContext() client.IRequesterContext {
	return r.context
}

func (r *StatusRequester) GetValidatorStatuses() (*types.ApiResponse[swapi.ValidatorStatusData], error) {
	return client.SendGetRequest[swapi.ValidatorStatusData](r, "status", "Status", nil)
}
