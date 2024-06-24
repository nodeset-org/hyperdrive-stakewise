package api_test

import (
	"testing"

	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/nodeset-org/osha"
	"github.com/stretchr/testify/require"
)

func TestRegistrationStatus_Registered(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(testMgr, snapshotName)

	client := testMgr.GetApiClient()
	response, err := client.Nodeset.RegistrationStatus()
	require.NoError(t, err)

	require.Equal(t, "", response.Data.ErrorMessage)
	require.Equal(t, swapi.NodesetRegistrationStatus_Registered, response.Data.Status)
}
func TestUploadDepositData(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(testMgr, snapshotName)
}

func TestRegisterNode_AlreadyRegisteredError(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := testMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(testMgr, snapshotName)

	client := testMgr.GetApiClient()
	_, err = client.Nodeset.RegisterNode("test@nodeset.io")
	require.Error(t, err)
}

func TestRegisterNode_Success(t *testing.T) {
	// Take a snapshot, revert at the end
	snapshotName, err := unregisteredTestMgr.CreateCustomSnapshot(osha.Service_EthClients | osha.Service_Filesystem)
	if err != nil {
		fail("Error creating custom snapshot: %v", err)
	}
	defer status_cleanup(unregisteredTestMgr, snapshotName)
	client := unregisteredTestMgr.GetApiClient()
	_, err = client.Nodeset.RegisterNode("test@nodeset.io")
	require.NoError(t, err)
}
