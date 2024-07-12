package swstatus

import (
	"context"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
)

type StatusHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider swcommon.IStakeWiseServiceProvider
	factories       []server.IContextFactory
}

func NewStatusHandler(logger *log.Logger, ctx context.Context, serviceProvider swcommon.IStakeWiseServiceProvider) *StatusHandler {
	h := &StatusHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&statusGetValidatorsStatusesContextFactory{h},
	}
	return h
}

func (h *StatusHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/status").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
