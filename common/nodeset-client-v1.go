package swcommon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/nodeset-org/hyperdrive-daemon/shared/types"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/nodeset-org/hyperdrive-stakewise/shared/keys"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Format for signing node registration messages
	nodeRegistrationMessageFormat string = `{"email":"%s","node_address":"%s"}`

	// Format for signing login messages
	loginMessageFormat string = `{"nonce":"%s","address":"%s"}`

	// Header to include when sending messages that have been logged in
	authHeader string = "Authorization"

	// Format for the authorization header
	authHeaderFormat string = "Bearer %s"

	// Value of the auth response header if the node hasn't registered yet
	unregisteredAddressKey string = "unregistered_address"

	// Value of the auth response header if the login token has expired
	invalidSessionKey string = "invalid_session"

	// The node address has already been confirmed on a NodeSet account
	addressAlreadyAuthorizedKey string = "address_already_authorized"

	// The node address hasn't been whitelisted on the provided NodeSet account
	addressMissingWhitelistKey string = "address_missing_whitelist"

	// API paths
	devPath         string = "dev"
	registerPath    string = "node-address"
	noncePath       string = "nonce"
	loginPath       string = "login"
	depositDataPath string = "deposit-data"
	metaPath        string = "meta"
	validatorsPath  string = "validators"
)

var (
	ErrUnregisteredNode  error = errors.New("node hasn't been registered with the NodeSet server yet")
	ErrAlreadyRegistered error = errors.New("node has already been registered with the NodeSet server")
	ErrNotWhitelisted    error = errors.New("node address hasn't been whitelisted on the provided NodeSet account")
)

// =================
// === Requests  ===
// =================

// Request to register a node with the NodeSet server
type RegisterNodeRequest struct {
	Email       string `json:"email"`
	NodeAddress string `json:"node_address"`
	Signature   string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Request to log into the NodeSet server
type LoginRequest struct {
	Nonce     string `json:"nonce"`
	Address   string `json:"address"`
	Signature string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Details of an exit message
type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

// Voluntary exit message
type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

// Data for a pubkey's voluntary exit message
type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exit_message"`
}

// =================
// === Responses ===
// =================

// All responses from the NodeSet API will have this format
// `message` may or may not be populated (but should always be populated if `ok` is false)
// `data` should be populated if `ok` is true, and will be omitted if `ok` is false
type NodeSetResponse[DataType any] struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Data    DataType `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Response to a login request
type LoginData struct {
	Token string `json:"token"`
}

// Data used returned from nonce requests
type NonceData struct {
	Nonce string `json:"nonce"`
	Token string `json:"token"`
}

// Response to a deposit data meta request
type DepositDataMetaData struct {
	Version int `json:"version"`
}

// Response to a deposit data request
type DepositDataData struct {
	Version     int                         `json:"version"`
	DepositData []types.ExtendedDepositData `json:"depositData"`
}

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	Status              string                 `json:"status"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// ==============
// === Client ===
// ==============

// Client for interacting with the Nodeset server
type NodeSetClient_v1 struct {
	sp    *StakeWiseServiceProvider
	res   *swconfig.StakewiseResources
	token string

	// Stores the node's registration status
	nodeRegistrationStatus swapi.NodesetRegistrationStatus
	lock                   *sync.Mutex
}

// Creates a new Nodeset client
func NewNodeSetClient_v1(sp *StakeWiseServiceProvider) *NodeSetClient_v1 {
	return &NodeSetClient_v1{
		sp:                     sp,
		res:                    sp.GetResources(),
		nodeRegistrationStatus: swapi.NodesetRegistrationStatus_Unknown,
		lock:                   &sync.Mutex{},
	}
}

// Logs out of the NodeSet service, resetting the registration status in the process.
func (c *NodeSetClient_v1) Logout() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.token = ""
	c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unknown
}

// Returns whether the node is registered with the NodeSet server
func (c *NodeSetClient_v1) GetNodeRegistrationStatus(ctx context.Context) (swapi.NodesetRegistrationStatus, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Force refresh the registration status if it hasn't been determined yet
	if c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_Unknown ||
		c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_NoWallet {
		err := c.loginImpl(ctx)
		return c.nodeRegistrationStatus, err
	}
	return c.nodeRegistrationStatus, nil
}

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func (c *NodeSetClient_v1) RegisterNode(ctx context.Context, email string, nodeWallet common.Address) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Sign the message
	hd := c.sp.GetHyperdriveClient()
	message := fmt.Sprintf(nodeRegistrationMessageFormat, email, nodeWallet.Hex())
	signResponse, err := hd.Wallet.SignMessage([]byte(message))
	if err != nil {
		return fmt.Errorf("error signing registration message: %w", err)
	}

	// Create the request
	signature := utils.EncodeHexWithPrefix(signResponse.Data.SignedMessage)
	request := RegisterNodeRequest{
		Email:       email,
		NodeAddress: nodeWallet.Hex(),
		Signature:   signature,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling registration request: %w", err)
	}

	logger.Debug("Sending Nodeset register node request",
		slog.String(log.BodyKey, string(jsonData)),
	)

	// Submit the request
	code, response, err := submitRequest_v1[string](c, ctx, false, http.MethodPost, bytes.NewBuffer(jsonData), nil, devPath, registerPath)
	if err != nil {
		return fmt.Errorf("error registering node: %w", err)
	}

	// Check for special errors
	if code == http.StatusBadRequest {
		switch response.Error {
		case addressAlreadyAuthorizedKey:
			return ErrAlreadyRegistered
		case addressMissingWhitelistKey:
			return ErrNotWhitelisted
		}
	}

	// Handle general errors
	if code != http.StatusOK {
		return fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}

	// Node successfully registered
	c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Registered
	return nil
}

// Uploads deposit data to Nodeset
func (c *NodeSetClient_v1) UploadDepositData(ctx context.Context, depositData []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	code, response, err := submitRequest_v1[string](c, ctx, true, http.MethodPost, bytes.NewBuffer(depositData), nil, devPath, depositDataPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}
	if err != nil {
		return fmt.Errorf("error uploading deposit data: %w", err)
	}
	return nil
}

// Submit signed exit data to Nodeset
func (c *NodeSetClient_v1) UploadSignedExitData(ctx context.Context, exitData []ExitData) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Serialize the exit data into JSON
	jsonData, err := json.Marshal(exitData)
	if err != nil {
		return fmt.Errorf("error marshalling exit data to JSON: %w", err)
	}
	params := map[string]string{
		"network": c.res.EthNetworkName,
	}

	// Submit the PATCH request with the serialized JSON
	code, response, err := submitRequest_v1[string](c, ctx, true, http.MethodPatch, bytes.NewBuffer(jsonData), params, devPath, validatorsPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}
	if err != nil {
		return fmt.Errorf("error submitting exit data: %w", err)
	}
	return nil
}

// Get the current version of the aggregated deposit data on the server
func (c *NodeSetClient_v1) GetServerDepositDataVersion(ctx context.Context) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	vault := utils.RemovePrefix(strings.ToLower(c.res.Vault.Hex()))
	params := map[string]string{
		"vault":   vault,
		"network": c.res.EthNetworkName,
	}
	code, response, err := submitRequest_v1[DepositDataMetaData](c, ctx, true, http.MethodGet, nil, params, devPath, depositDataPath, metaPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}
	if err != nil {
		return 0, fmt.Errorf("error getting deposit data version: %w", err)
	}
	return response.Data.Version, nil
}

// Get the aggregated deposit data from the server
func (c *NodeSetClient_v1) GetServerDepositData(ctx context.Context) (int, []types.ExtendedDepositData, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	vault := utils.RemovePrefix(strings.ToLower(c.res.Vault.Hex()))
	params := map[string]string{
		"vault":   vault,
		"network": c.res.EthNetworkName,
	}
	code, response, err := submitRequest_v1[DepositDataData](c, ctx, true, http.MethodGet, nil, params, devPath, depositDataPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}
	if err != nil {
		return 0, nil, fmt.Errorf("error getting deposit data: %w", err)
	}
	return response.Data.Version, response.Data.DepositData, nil
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node
func (c *NodeSetClient_v1) GetRegisteredValidators(ctx context.Context) ([]ValidatorStatus, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	queryParams := map[string]string{
		"network": c.res.EthNetworkName,
	}
	code, response, err := submitRequest_v1[ValidatorsData](c, ctx, true, http.MethodGet, nil, queryParams, devPath, validatorsPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, response.Message)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting registered validators: %w", err)
	}
	return response.Data.Validators, nil
}

// Send a request to the server and read the response
func submitRequest_v1[DataType any](c *NodeSetClient_v1, ctx context.Context, requireAuth bool, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (int, NodeSetResponse[DataType], error) {
	var defaultVal NodeSetResponse[DataType]

	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Make the request
	path, err := url.JoinPath(c.res.NodesetApiUrl, subroutes...)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error joining path [%v]: %w", subroutes, err)
	}
	request, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error generating request to [%s]: %w", path, err)
	}
	query := request.URL.Query()
	for name, value := range queryParams {
		query.Add(name, value)
	}
	request.URL.RawQuery = query.Encode()

	// Set the headers
	if requireAuth {
		// Make sure the auth token exists
		if c.token == "" {
			err = c.loginImpl(ctx)
			if err != nil {
				return 0, defaultVal, fmt.Errorf("error logging in before submitting request: %w", err)
			}
		}
		request.Header.Set(authHeader, fmt.Sprintf(authHeaderFormat, c.token))
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Upload it to the server
	logger.Debug("Sending NodeSet server request", slog.String(log.QueryKey, request.URL.String()))

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error submitting request to nodeset server: %w", err)
	}

	// Read the body
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("nodeset server responded to request with code %s but reading the response body failed: %w", resp.Status, err)
	}

	// Unmarshal the response
	var response NodeSetResponse[DataType]
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("nodeset server responded to request with code %s and unmarshalling the response failed: [%w]... original body: [%s]", resp.Status, err, string(bytes))
	}

	// Check for auth issues
	if resp.StatusCode == http.StatusUnauthorized {
		logger.Debug("NodeSet responded with 401 Unauthorized", slog.String(log.ErrorKey, response.Error))
		switch response.Error {
		case unregisteredAddressKey:
			c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unregistered
			return 0, defaultVal, ErrUnregisteredNode
		case invalidSessionKey:
			// Try logging in again
			err = c.loginImpl(ctx)
			if err != nil {
				return 0, defaultVal, fmt.Errorf("error logging in after token expired: %w", err)
			}

			// Try the request again
			return submitRequest_v1[DataType](c, ctx, requireAuth, method, body, queryParams, subroutes...)
		}
	}

	// Debug log
	logger.Debug("NodeSet response:", slog.String(log.CodeKey, resp.Status), slog.String(keys.MessageKey, response.Message))
	return resp.StatusCode, response, nil
}

// Logs into the NodeSet API server, grabbing a new authentication token
func (c *NodeSetClient_v1) Login(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.loginImpl(ctx)
}

// Implementation for logging in
func (c *NodeSetClient_v1) loginImpl(ctx context.Context) error {
	// Get the logger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a logger!")
	}

	// Get the node wallet
	hd := c.sp.GetHyperdriveClient()
	walletStatusResponse, err := hd.Wallet.Status()
	if err != nil {
		return fmt.Errorf("error getting wallet status for login: %w", err)
	}
	walletStatus := walletStatusResponse.Data.WalletStatus
	err = c.sp.RequireWalletReady(walletStatus)
	if err != nil {
		c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_NoWallet
		return fmt.Errorf("can't log into nodeset, hyperdrive wallet not initialized yet")
	}

	// Log the login attempt
	logger.Info("Not authenticated with the NodeSet server, logging in")

	// Get the nonce
	code, nonceResponse, err := submitRequest_v1[NonceData](c, ctx, false, http.MethodGet, nil, nil, devPath, noncePath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, nonceResponse.Message)
	}
	if err != nil {
		// If the status was no-wallet, set it back to unknown
		if c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_NoWallet {
			c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unknown
		}
		return fmt.Errorf("error getting nonce for login: %w", err)
	}
	nonceData := nonceResponse.Data
	logger.Debug("Got nonce for login",
		slog.String(keys.NonceKey, nonceData.Nonce),
	)
	c.token = nonceData.Token // Store this as a temp token for the login request

	// Create the signature
	nodeAddress := walletStatus.Wallet.WalletAddress
	message := fmt.Sprintf(loginMessageFormat, nonceData.Nonce, nodeAddress.Hex())
	signResponse, err := hd.Wallet.SignMessage([]byte(message))
	if err != nil {
		// If the status was no-wallet, set it back to unknown
		if c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_NoWallet {
			c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unknown
		}
		return fmt.Errorf("error signing login message: %w", err)
	}

	// Create the request
	signature := utils.EncodeHexWithPrefix(signResponse.Data.SignedMessage)
	request := LoginRequest{
		Nonce:     nonceData.Nonce,
		Address:   nodeAddress.Hex(),
		Signature: signature,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		// If the status was no-wallet, set it back to unknown
		if c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_NoWallet {
			c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unknown
		}
		return fmt.Errorf("error marshalling login request: %w", err)
	}

	// Submit the request
	code, loginResponse, err := submitRequest_v1[LoginData](c, ctx, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, devPath, loginPath)
	if err == nil && code != http.StatusOK {
		err = fmt.Errorf("nodeset server responded to request with code %d: [%s]", code, loginResponse.Message)
	}
	if err != nil {
		// If the status was no-wallet, set it back to unknown
		if c.nodeRegistrationStatus == swapi.NodesetRegistrationStatus_NoWallet {
			c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Unknown
		}
		return fmt.Errorf("error logging in: %w", err)
	}
	loginData := loginResponse.Data
	c.token = loginData.Token // Save this as the persistent token for all other requests
	logger.Debug("Got nonce for session",
		slog.String(keys.NonceKey, nonceData.Nonce),
	)

	// Log the successful login
	logger.Info("Logged into NodeSet server")
	c.nodeRegistrationStatus = swapi.NodesetRegistrationStatus_Registered

	return nil
}

func IsUploadedToNodeset(pubKey beacon.ValidatorPubkey, registeredPubkeys []beacon.ValidatorPubkey) bool {
	for _, registeredPubKey := range registeredPubkeys {
		if registeredPubKey == pubKey {
			return true
		}
	}
	return false
}

func IsRegisteredToStakewise(pubKey beacon.ValidatorPubkey, statuses map[beacon.ValidatorPubkey]beacon.ValidatorStatus) bool {
	// TODO: Implement
	return false
}

func IsUploadedStakewise(pubKey beacon.ValidatorPubkey, statuses map[beacon.ValidatorPubkey]beacon.ValidatorStatus) bool {
	// TODO: Implement
	return false
}
