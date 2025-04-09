package relay

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	nscommon "github.com/nodeset-org/nodeset-client-go/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Request from the StakeWise Operator to get validators
// See https://docs.stakewise.io/for-operators/operator-service/running-as-api-service
type ValidatorsRequest struct {
	// The address of the vault that validators will be created for
	Vault common.Address `json:"vault"`

	// The validator index to start from when creating signed exit messages
	ValidatorsStartIndex int `json:"validators_start_index"`

	// Number of validators in current batch. Maximum batch size is determined by prococol config, currently 10.
	ValidatorsBatchSize int `json:"validators_batch_size"`

	// Total number of validators supplied by vault assets.
	// validators_total should be more than or equal to validators_count.
	// Relayer may use this value to allocate larger portions of validators in background.
	ValidatorsTotal int `json:"validators_total"`
}

// Info about a validator that can be deposited
// See https://docs.stakewise.io/for-operators/operator-service/running-as-api-service
type ValidatorInfo struct {
	// The public key of the validator
	PublicKey beacon.ValidatorPubkey `json:"public_key"`

	// The signature of the validator's deposit data
	DepositSignature string `json:"deposit_signature"`

	// The amount of Gwei to send for the deposit
	AmountGwei uint64 `json:"amount_gwei"`

	// The signature of the validator's exit message
	ExitSignature string `json:"exit_signature"`
}

// Response to the StakeWise Operator for a request to get validators
// See https://docs.stakewise.io/for-operators/operator-service/running-as-api-service
type ValidatorsResponse struct {
	// List of validator deposit info that can be used
	Validators []ValidatorInfo `json:"validators"`

	// The signature from NodeSet for the validators
	ValidatorsManagerSignature string `json:"validators_manager_signature"`
}

// Handle a request to get validators from the StakeWise Operator
func (h *baseHandler) getValidators(w http.ResponseWriter, r *http.Request) {
	// Check if the request is already in progress
	h.validatorsLock.Lock()
	if h.validatorsBusy {
		h.validatorsLock.Unlock()
		HandleError(w, h.logger, http.StatusTooManyRequests, fmt.Errorf("validators request already in progress"))
		return
	}

	// Mark the handler as busy to prevent requests until it's done
	h.validatorsBusy = true
	h.validatorsLock.Unlock()
	defer func() {
		h.validatorsLock.Lock()
		h.validatorsBusy = false
		h.validatorsLock.Unlock()
	}()

	// Get the services
	logger := h.logger
	ctx := h.ctx
	sp := h.sp
	res := sp.GetResources()
	hd := sp.GetHyperdriveClient()
	qMgr := sp.GetQueryManager()
	keyMgr := sp.GetAvailableKeyManager()
	bn := sp.GetBeaconClient()
	start := time.Now()

	// Parse the body
	var request ValidatorsRequest
	pathArgs, queryArgs := ProcessApiRequest(logger, w, r, &request)
	if pathArgs == nil && queryArgs == nil {
		return
	}
	logger.Debug("Parsed request", "elapsed", time.Since(start))

	// Short-circuit if the private keys haven't been loaded yet
	if !keyMgr.HasLoadedKeys() {
		logger.Debug("Private keys need to be loaded, loading now")
		HandleError(w, logger, http.StatusServiceUnavailable, fmt.Errorf("still loading private keys"))
		keyMgr.LoadPrivateKeys(logger)
		return
	}

	// Short-circuit if there aren't any validators to get
	if !keyMgr.HasKeyCandidates() {
		logger.Debug("No candidate keys present")
		HandleSuccess(w, logger, ValidatorsResponse{
			Validators: []ValidatorInfo{},
		})
		return
	}

	// Check the wallet
	walletResponse, err := hd.Wallet.Status()
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting wallet status: %w", err))
		return
	}
	logger.Debug("Verified wallet status", "elapsed", time.Since(start))
	walletStatus := walletResponse.Data.WalletStatus
	err = sp.RequireStakewiseWalletReady(ctx, walletStatus)
	if err != nil {
		HandleError(w, logger, http.StatusUnprocessableEntity, fmt.Errorf("error checking wallet status: %w", err))
		return
	}
	logger.Debug("StakeWise wallet ready", "elapsed", time.Since(start))

	// Check if NodeSet can support more validators
	validatorsInfo, err := hd.NodeSet_StakeWise.GetValidatorsInfo(res.DeploymentName, res.Vault)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting meta info from nodeset: %w", err))
		return
	}
	if validatorsInfo.Data.NotRegistered {
		HandleError(w, logger, http.StatusUnprocessableEntity, fmt.Errorf("node is not registered with nodeset"))
		return
	}
	availableForNodeSet := validatorsInfo.Data.Available
	if availableForNodeSet == 0 {
		// Return an empty response
		HandleSuccess(w, logger, ValidatorsResponse{
			Validators: []ValidatorInfo{},
		})
		return
	}
	logger.Debug("Got meta info from NodeSet", "elapsed", time.Since(start), "available", availableForNodeSet)

	// Get the current Beacon deposit root
	err = sp.RequireEthClientSynced(ctx)
	if err != nil {
		if errors.Is(err, services.ErrExecutionClientNotSynced) {
			HandleError(w, logger, http.StatusUnprocessableEntity, err)
			return
		}
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error checking eth client status: %w", err))
		return
	}
	var depositRoot common.Hash
	err = qMgr.Query(func(mc *batch.MultiCaller) error {
		sp.GetBeaconDepositContract().GetDepositRoot(mc, &depositRoot)
		return nil
	}, nil)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting latest Beacon deposit root: %w", err))
		return
	}
	logger.Debug("Got deposit root", "elapsed", time.Since(start), "root", depositRoot.Hex())

	// Short-circuit if lookback scanning is required because it will take too long
	if keyMgr.RequiresLookbackScan() {
		logger.Info("Lookback scan required for new keys, starting scan")
		HandleError(w, logger, http.StatusServiceUnavailable, fmt.Errorf("lookback scan required for new keys, try again later"))
		_, _, err := keyMgr.GetAvailableKeys(ctx, logger, depositRoot, swcommon.GetAvailableKeyOptions{
			SkipSyncCheck:  true,
			DoLookbackScan: true,
		})
		if err != nil {
			logger.Error("Error during lookback scan", "error", err)
		}
		return
	}

	// Get the available keys, clamping to the number of validators requested
	scanOpts := swcommon.GetAvailableKeyOptions{
		SkipSyncCheck:  true,
		DoLookbackScan: false,
	}
	availableKeys, ineligibleKeys, err := keyMgr.GetAvailableKeys(ctx, logger, depositRoot, scanOpts)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting available keys: %w", err))
		return
	}

	// Log the ineligible keys
	if len(ineligibleKeys) > 0 {
		logger.Info("Ineligible keys found", "count", len(ineligibleKeys))
		for key, reason := range ineligibleKeys {
			switch reason {
			case swcommon.IneligibleReason_NoPrivateKey:
				logger.Info("No private key found", "key", key.PublicKey.HexWithPrefix())
			case swcommon.IneligibleReason_LookbackScanRequired:
				logger.Info("Key requires lookback scan still?", "key", key.PublicKey.HexWithPrefix())
			case swcommon.IneligibleReason_OnBeacon:
				logger.Info("Key already seen on Beacon", "key", key.PublicKey.HexWithPrefix())
			case swcommon.IneligibleReason_HasDepositEvent:
				logger.Info("Key has a deposit event already", "key", key.PublicKey.HexWithPrefix())
			case swcommon.IneligibleReason_AlreadyUsedDepositRoot:
				logger.Info("Key has already used this deposit root", "key", key.PublicKey.HexWithPrefix())
			default:
				logger.Info("Key is ineligible for unknown reason", "key", key.PublicKey.HexWithPrefix(), "reason", reason)
			}
		}
	}

	// Check if there are any available keys
	if len(availableKeys) == 0 {
		// Return an empty response
		logger.Debug("No available keys", "elapsed", time.Since(start))
		HandleSuccess(w, logger, ValidatorsResponse{
			Validators: []ValidatorInfo{},
		})
		return
	}
	if len(availableKeys) > request.ValidatorsBatchSize {
		logger.Debug("Clamping available keys to requested count", "available", len(availableKeys), "requested", request.ValidatorsBatchSize)
		availableKeys = availableKeys[:request.ValidatorsBatchSize]
	}
	if availableForNodeSet < len(availableKeys) {
		logger.Debug("Clamping available keys to NodeSet limit", "available", len(availableKeys), "nodeSetLimit", availableForNodeSet)
		availableKeys = availableKeys[:availableForNodeSet]
	}
	debugEntries := []any{
		"elapsed",
		time.Since(start),
	}
	for _, key := range availableKeys {
		debugEntries = append(debugEntries, "key", key.PublicKey.HexWithPrefix())
	}
	logger.Info("Got available keys", debugEntries...)

	// Create the deposit data
	privateKeys := make([]*eth2types.BLSPrivateKey, len(availableKeys))
	for i, key := range availableKeys {
		privateKeys[i] = key.PrivateKey
	}
	depositDatas, err := swcommon.GenerateDepositData(logger, res, privateKeys)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error generating deposit data: %w", err))
		return
	}
	logger.Debug("Generated deposit data", "elapsed", time.Since(start))

	// Create signed exits
	signatureDomain, err := bn.GetDomainData(ctx, eth2types.DomainVoluntaryExit[:], res.CapellaForkEpoch, false)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting voluntary exit domain data: %w", err))
	}
	exitMessages := make([]nscommon.ExitMessage, len(availableKeys))
	currentIndex := uint64(request.ValidatorsStartIndex)
	for i, key := range privateKeys {
		exitMessage, err := createSignedExitMessage(key, currentIndex, res.CapellaForkEpoch, signatureDomain)
		if err != nil {
			HandleError(w, logger, http.StatusInternalServerError, err)
			return
		}
		exitMessages[i] = exitMessage
		currentIndex++
	}
	logger.Debug("Generated exit messages", "elapsed", time.Since(start))

	// Encrypt the exits
	encryptedExits := make([]string, len(availableKeys))
	for i, exitMessage := range exitMessages {
		encryptedMessage, err := nscommon.EncryptSignedExitMessage(exitMessage, res.EncryptionPubkey)
		if err != nil {
			HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf(
				"error encrypting signed exit message for [%s]: %w", availableKeys[i].PublicKey.HexWithPrefix(), err),
			)
			return
		}
		encryptedExits[i] = encryptedMessage
	}
	logger.Debug("Encrypted exit messages", "elapsed", time.Since(start))

	// Get a signature from NodeSet
	signatureResponse, err := hd.NodeSet_StakeWise.GetValidatorManagerSignature(res.DeploymentName, res.Vault, depositRoot, depositDatas, encryptedExits)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error getting validators signature from nodeset: %w", err))
		return
	}
	if signatureResponse.Data.NotRegistered {
		HandleError(w, logger, http.StatusUnprocessableEntity, fmt.Errorf("node is not registered with nodeset"))
		return
	}
	if signatureResponse.Data.InvalidPermissions {
		HandleError(w, logger, http.StatusUnauthorized, fmt.Errorf("node does not have permission to register validators with this deployment"))
		return
	}
	if signatureResponse.Data.VaultNotFound {
		HandleError(w, logger, http.StatusUnprocessableEntity, fmt.Errorf("nodeset cannot find vault [%s] on deployment [%s]", res.Vault.Hex(), res.DeploymentName))
		return
	}
	signature := signatureResponse.Data.Signature
	logger.Debug("Got validators signature from NodeSet", "elapsed", time.Since(start), "signature", signature)

	// Set the last deposit root for those keys
	err = keyMgr.SetLastDepositRoot(availableKeys, depositRoot)
	if err != nil {
		HandleError(w, logger, http.StatusInternalServerError, fmt.Errorf("error setting last deposit root: %w", err))
		return
	}
	logger.Debug("Updated available keys with last deposit root", "elapsed", time.Since(start))

	// Return the validators to SW
	response := ValidatorsResponse{
		Validators:                 make([]ValidatorInfo, len(availableKeys)),
		ValidatorsManagerSignature: signature,
	}
	for i, key := range availableKeys {
		response.Validators[i] = ValidatorInfo{
			PublicKey:        key.PublicKey,
			DepositSignature: beacon.ValidatorSignature(depositDatas[i].Signature[:]).HexWithPrefix(), // Type conversion is annoying, should be standardized somewhere
			AmountGwei:       depositDatas[i].Amount,
			ExitSignature:    exitMessages[i].Signature,
		}
	}
	HandleSuccess(w, logger, response)
	logger.Debug("Relay processing complete", "elapsed", time.Since(start))
}

// Create a signed exit message for a validator
// TODO: This really needs to be baseline in NMC, not just the signature generator
func createSignedExitMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex uint64, epoch uint64, signatureDomain []byte) (nscommon.ExitMessage, error) {
	indexString := strconv.FormatUint(validatorIndex, 10)
	exitMessage := nscommon.ExitMessage{
		Message: nscommon.ExitMessageDetails{
			Epoch:          strconv.FormatUint(epoch, 10),
			ValidatorIndex: indexString,
		},
	}
	exitMessageSignature, err := validator.GetSignedExitMessage(validatorKey, indexString, epoch, signatureDomain)
	if err != nil {
		pubkey := beacon.ValidatorPubkey(validatorKey.PublicKey().Marshal())
		return nscommon.ExitMessage{}, fmt.Errorf("error getting signed exit message for validator [%s]: %w", pubkey.HexWithPrefix(), err)
	}
	exitMessage.Signature = exitMessageSignature.HexWithPrefix()
	return exitMessage, nil
}
