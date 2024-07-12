package swclient

import (
	"log/slog"
	"net/http/httptrace"
	"net/url"

	"github.com/rocket-pool/node-manager-core/api/client"
)

// Binder for the StakeWise API server
type ApiClient struct {
	context   client.IRequesterContext
	Nodeset   *NodesetRequester
	Validator *ValidatorRequester
	Wallet    *WalletRequester
	Status    *StatusRequester
}

// Creates a new API client instance
func NewApiClient(apiUrl *url.URL, logger *slog.Logger, tracer *httptrace.ClientTrace) *ApiClient {
	context := client.NewNetworkRequesterContext(apiUrl, logger, tracer)

	client := &ApiClient{
		context:   context,
		Nodeset:   NewNodesetRequester(context),
		Validator: NewValidatorRequester(context),
		Wallet:    NewWalletRequester(context),
		Status:    NewStatusRequester(context),
	}
	return client
}
