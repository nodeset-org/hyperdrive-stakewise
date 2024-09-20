package swservice

import (
	"context"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
)

type ServiceHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider swcommon.IStakeWiseServiceProvider
	factories       []server.IContextFactory
}

func NewServiceHandler(logger *log.Logger, ctx context.Context, serviceProvider swcommon.IStakeWiseServiceProvider) *ServiceHandler {
	h := &ServiceHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&serviceGetNetworkSettingsContextFactory{h},
		&serviceGetResourcesContextFactory{h},
		&serviceVersionContextFactory{h},
	}
	return h
}

func (h *ServiceHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/service").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
