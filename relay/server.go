package relay

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/log"
)

type RelayServer struct {
	logPath string
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	sp      swcommon.IStakeWiseServiceProvider
	ctx     context.Context

	// Route handlers
	baseHandler *baseHandler
}

func NewRelayServer(sp swcommon.IStakeWiseServiceProvider, ip string, port uint16) (*RelayServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create a context with the logger attached
	hdCfg := sp.GetHyperdriveConfig()
	relayLogPath := hdCfg.GetModuleLogFilePath(swconfig.ModuleName, swconfig.RelayLogName)
	relayLogger, err := log.NewLogger(relayLogPath, hdCfg.GetLoggerOptions())
	if err != nil {
		return nil, fmt.Errorf("error creating relay logger: %w", err)
	}
	ctx := relayLogger.CreateContextWithLogger(sp.GetBaseContext())

	// Create the manager
	server := &RelayServer{
		logPath: relayLogPath,
		logger:  relayLogger.Logger,
		ip:      ip,
		port:    port,
		router:  router,
		server: http.Server{
			Handler: router,
		},
		sp:  sp,
		ctx: ctx,
	}
	server.baseHandler = NewBaseHandler(sp, relayLogger.Logger, ctx)

	// Register base routes
	server.baseHandler.RegisterRoutes(router)

	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *RelayServer) Start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Get the port if random
	if s.port == 0 {
		s.port = uint16(socket.Addr().(*net.TCPAddr).Port)
	}

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error while listening for HTTP requests", log.Err(err))
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *RelayServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Get the port the server is listening on
func (s *RelayServer) GetPort() uint16 {
	return s.port
}

// Get the path to the log file
func (s *RelayServer) GetLogPath() string {
	return s.logPath
}
