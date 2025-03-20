package relay

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
)

// Base routes for the relay
type baseHandler struct {
	sp     swcommon.IStakeWiseServiceProvider
	logger *slog.Logger
	ctx    context.Context
}

// Create a new base handler
func NewBaseHandler(sp swcommon.IStakeWiseServiceProvider, logger *slog.Logger, ctx context.Context) *baseHandler {
	return &baseHandler{
		sp:     sp,
		logger: logger,
		ctx:    ctx,
	}
}

func (h *baseHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/validators", h.getValidators).Methods(http.MethodPost)
}
