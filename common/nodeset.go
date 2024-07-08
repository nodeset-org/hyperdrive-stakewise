package swcommon

import (
	swtypes "github.com/nodeset-org/hyperdrive-stakewise/shared/types"
	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
	"github.com/rocket-pool/node-manager-core/beacon"
)

func IsUploadedToNodeset(pubKey beacon.ValidatorPubkey, registeredPubkeys []beacon.ValidatorPubkey) bool {
	for _, registeredPubKey := range registeredPubkeys {
		if registeredPubKey == pubKey {
			return true
		}
	}
	return false
}

func GetNodesetStatus(pubKey beacon.ValidatorPubkey, registeredPubkeysStatusMapping map[beacon.ValidatorPubkey]apiv1.StakeWiseStatus) swtypes.NodesetStatus {
	status, exists := registeredPubkeysStatusMapping[pubKey]
	if !exists {
		return swtypes.NodesetStatus_Generated
	}

	switch status {
	case apiv1.StakeWiseStatus_Registered:
		return swtypes.NodesetStatus_RegisteredToStakewise
	case apiv1.StakeWiseStatus_Uploaded:
		return swtypes.NodesetStatus_UploadedStakewise
	default:
		return swtypes.NodesetStatus_UploadedToNodeset
	}
}
