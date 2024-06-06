package swclient

import (
	"github.com/ethereum/go-ethereum/common"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
)

type NodesetRequester struct {
	context client.IRequesterContext
}

func NewNodesetRequester(context client.IRequesterContext) *NodesetRequester {
	return &NodesetRequester{
		context: context,
	}
}

func (r *NodesetRequester) GetName() string {
	return "Nodeset"
}
func (r *NodesetRequester) GetRoute() string {
	return "nodeset"
}
func (r *NodesetRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Set the validators root for the NodeSet vault
func (r *NodesetRequester) SetValidatorsRoot(root common.Hash) (*types.ApiResponse[types.TxInfoData], error) {
	args := map[string]string{
		"root": root.Hex(),
	}
	return client.SendGetRequest[types.TxInfoData](r, "set-validators-root", "SetValidatorsRoot", args)
}

// Upload the aggregated deposit data file to NodeSet's servers
func (r *NodesetRequester) UploadDepositData() (*types.ApiResponse[swapi.NodesetUploadDepositDataData], error) {
	args := map[string]string{}
	return client.SendGetRequest[swapi.NodesetUploadDepositDataData](r, "upload-deposit-data", "UploadDepositData", args)
}

// Register node with NodeSet
func (r *NodesetRequester) RegisterNode(email string) (*types.ApiResponse[swapi.NodeSetRegisterNodeData], error) {
	args := map[string]string{
		"email": email,
	}
	return client.SendGetRequest[swapi.NodeSetRegisterNodeData](r, "register-node", "RegisterNode", args)
}

// Get the node's NodeSet registration status
func (r *NodesetRequester) RegistrationStatus() (*types.ApiResponse[swapi.NodeSetRegistrationStatusData], error) {
	args := map[string]string{}
	return client.SendGetRequest[swapi.NodeSetRegistrationStatusData](r, "registration-status", "RegistrationStatus", args)
}
