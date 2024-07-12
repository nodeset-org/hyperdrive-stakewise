package server

import (
	"fmt"
	"sync"

	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swnodeset "github.com/nodeset-org/hyperdrive-stakewise/server/nodeset"
	swstatus "github.com/nodeset-org/hyperdrive-stakewise/server/status"
	swvalidator "github.com/nodeset-org/hyperdrive-stakewise/server/validator"
	swwallet "github.com/nodeset-org/hyperdrive-stakewise/server/wallet"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/api/server"
)

// ServerManager manages the API server run by the daemon
type ServerManager struct {
	// The server for clients to interact with
	apiServer *server.NetworkSocketApiServer
}

// Creates a new server manager
func NewServerManager(sp swcommon.IStakeWiseServiceProvider, ip string, port uint16, stopWg *sync.WaitGroup) (*ServerManager, error) {
	// Start the API server
	apiServer, err := createServer(sp, ip, port)
	if err != nil {
		return nil, fmt.Errorf("error creating API server: %w", err)
	}
	err = apiServer.Start(stopWg)
	if err != nil {
		return nil, fmt.Errorf("error starting API server: %w", err)
	}
	port = apiServer.GetPort()
	fmt.Printf("API server started on %s:%d\n", ip, port)

	// Create the manager
	mgr := &ServerManager{
		apiServer: apiServer,
	}
	return mgr, nil
}

// Returns the port the server is running on
func (m *ServerManager) GetPort() uint16 {
	return m.apiServer.GetPort()
}

// Stops and shuts down the servers
func (m *ServerManager) Stop() {
	err := m.apiServer.Stop()
	if err != nil {
		fmt.Printf("WARNING: API server didn't shutdown cleanly: %s\n", err.Error())
	}
}

// Creates a new Hyperdrive API server
func createServer(sp swcommon.IStakeWiseServiceProvider, ip string, port uint16) (*server.NetworkSocketApiServer, error) {
	apiLogger := sp.GetApiLogger()
	ctx := apiLogger.CreateContextWithLogger(sp.GetBaseContext())

	handlers := []server.IHandler{
		swnodeset.NewNodesetHandler(apiLogger, ctx, sp),
		swvalidator.NewValidatorHandler(apiLogger, ctx, sp),
		swwallet.NewWalletHandler(apiLogger, ctx, sp),
		swstatus.NewStatusHandler(apiLogger, ctx, sp),
	}
	server, err := server.NewNetworkSocketApiServer(apiLogger.Logger, ip, port, handlers, swconfig.DaemonBaseRoute, swconfig.ApiVersion)
	if err != nil {
		return nil, err
	}
	return server, nil
}
