package swclient

import (
	"log/slog"
	"net/http/httptrace"
	"net/url"

	"github.com/nodeset-org/hyperdrive-daemon/shared/auth"
	"github.com/rocket-pool/node-manager-core/api/client"
)

// Binder for the StakeWise API server
type ApiClient struct {
	context   client.IRequesterContext
	Network   *NetworkRequester
	Validator *ValidatorRequester
	Wallet    *WalletRequester
	Service   *ServiceRequester
}

// Creates a new API client instance
func NewApiClient(apiUrl *url.URL, logger *slog.Logger, tracer *httptrace.ClientTrace, authMgr *auth.AuthorizationManager) *ApiClient {
	context := client.NewNetworkRequesterContext(apiUrl, logger, tracer, authMgr.AddAuthHeader)

	client := &ApiClient{
		context:   context,
		Network:   NewNetworkRequester(context),
		Validator: NewValidatorRequester(context),
		Wallet:    NewWalletRequester(context),
		Service:   NewServiceRequester(context),
	}
	return client
}
