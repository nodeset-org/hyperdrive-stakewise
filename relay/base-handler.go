package relay

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
)

const (
	ValidatorsPath = "validators"
)

// Base routes for the relay
type baseHandler struct {
	sp     swcommon.IStakeWiseServiceProvider
	logger *slog.Logger
	ctx    context.Context

	validatorsLock sync.Mutex
	validatorsBusy bool
}

// Create a new base handler
func NewBaseHandler(sp swcommon.IStakeWiseServiceProvider, logger *slog.Logger, ctx context.Context) *baseHandler {
	return &baseHandler{
		sp:             sp,
		logger:         logger,
		ctx:            ctx,
		validatorsLock: sync.Mutex{},
		validatorsBusy: false,
	}
}

func (h *baseHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/"+ValidatorsPath, h.getValidators).Methods(http.MethodPost)
}
