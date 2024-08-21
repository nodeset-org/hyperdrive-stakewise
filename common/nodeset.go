package swcommon

import (
	swtypes "github.com/nodeset-org/hyperdrive-stakewise/shared/types"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
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

func GetNodesetStatus(pubKey beacon.ValidatorPubkey, registeredPubkeysStatusMapping map[beacon.ValidatorPubkey]stakewise.StakeWiseStatus) swtypes.NodesetStatus {
	status, exists := registeredPubkeysStatusMapping[pubKey]
	if !exists {
		return swtypes.NodesetStatus_Generated
	}

	switch status {
	case stakewise.StakeWiseStatus_Registered:
		return swtypes.NodesetStatus_RegisteredToStakewise
	case stakewise.StakeWiseStatus_Uploaded:
		return swtypes.NodesetStatus_UploadedStakewise
	default:
		return swtypes.NodesetStatus_UploadedToNodeset
	}
}
