package swnodeset

import (
	"context"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
)

type NodesetHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *swcommon.StakeWiseServiceProvider
	factories       []server.IContextFactory
}

func NewNodesetHandler(logger *log.Logger, ctx context.Context, serviceProvider *swcommon.StakeWiseServiceProvider) *NodesetHandler {
	h := &NodesetHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&nodesetRegistrationStatusContextFactory{h},
		&nodesetRegisterNodeContextFactory{h},
		&nodesetSetValidatorsRootContextFactory{h},
		&nodesetUploadDepositDataContextFactory{h},
		&nodesetGenerateDepositDataContextFactory{h},
	}
	return h
}

func (h *NodesetHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/nodeset").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
