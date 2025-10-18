package relay

import "net/http"

// Response to the StakeWise Operator for an info reqest
// See https://docs.stakewise.io/operator/alternative-key-management/api-mode
type InfoResponse struct {
	Network string `json:"network"`
}

// Handle a request to get info from the StakeWise Operator
func (h *baseHandler) getInfo(w http.ResponseWriter, r *http.Request) {
	// Get the services
	logger := h.logger
	sp := h.sp
	cfg := sp.GetHyperdriveConfig()
	response := &InfoResponse{
		Network: cfg.GetEthNetworkName(),
	}
	HandleSuccess(w, logger, response)
}
