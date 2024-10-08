package swtasks

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/nodeset-org/hyperdrive-daemon/client"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

const (
	PubkeyKey string = "pubkey"
)

// Send exit data task
type SendExitDataTask struct {
	logger *log.Logger
	ctx    context.Context
	sp     swcommon.IStakeWiseServiceProvider
	w      *swcommon.Wallet
	hd     *client.ApiClient
	bc     beacon.IBeaconClient
	res    *swconfig.MergedResources
}

// Create Exit data task
func NewSendExitDataTask(ctx context.Context, sp swcommon.IStakeWiseServiceProvider, logger *log.Logger) *SendExitDataTask {
	return &SendExitDataTask{
		logger: logger,
		ctx:    ctx,
		sp:     sp,
		w:      sp.GetWallet(),
		hd:     sp.GetHyperdriveClient(),
		bc:     sp.GetBeaconClient(),
		res:    sp.GetResources(),
	}
}

// Update Exit data
func (t *SendExitDataTask) Run() error {
	t.logger.Info("Checking for missing signed exit data...")

	// Get registered validators
	resp, err := t.hd.NodeSet_StakeWise.GetRegisteredValidators(t.res.DeploymentName, t.res.Vault)
	if err != nil {
		return fmt.Errorf("error getting registered validators: %w", err)
	}
	if resp.Data.NotRegistered {
		t.logger.Warn("Node is not registered with the NodeSet server yet.")
		return nil
	}
	for _, status := range resp.Data.Validators {
		t.logger.Debug(
			"Retrieved registered validator",
			slog.String(PubkeyKey, status.Pubkey.HexWithPrefix()),
			slog.Bool("uploaded", status.ExitMessageUploaded),
		)
	}

	// Check for any that are missing signed exits
	missingExitPubkeys := []beacon.ValidatorPubkey{}
	for _, v := range resp.Data.Validators {
		if v.ExitMessageUploaded {
			continue
		}
		missingExitPubkeys = append(missingExitPubkeys, v.Pubkey)
	}
	if len(missingExitPubkeys) == 0 {
		return nil
	}

	// Get statuses for validators with missing exits
	statuses, err := t.bc.GetValidatorStatuses(t.ctx, missingExitPubkeys, nil)
	if err != nil {
		return fmt.Errorf("error getting validator statuses: %w", err)
	}

	// Get beacon head and domain data
	head, err := t.bc.GetBeaconHead(t.ctx)
	if err != nil {
		return fmt.Errorf("error getting beacon head: %w", err)
	}
	epoch := head.FinalizedEpoch
	signatureDomain, err := t.bc.GetDomainData(t.ctx, eth2types.DomainVoluntaryExit[:], epoch, false)
	if err != nil {
		return fmt.Errorf("error getting domain data: %w", err)
	}

	// Get signed exit messages
	exitData := []common.EncryptedExitData{}
	for _, pubkey := range missingExitPubkeys {
		// Check if it's been seen on Beacon
		index := statuses[pubkey].Index
		if index == "" {
			t.logger.Debug("Validator doesn't have an index yet", slog.String(PubkeyKey, pubkey.HexWithPrefix()))
			continue
		}

		// Make a signed exit message
		t.logger.Info("Validator has been added to the Beacon queue but is missing a signed exit message.", slog.String(PubkeyKey, pubkey.HexWithPrefix()))
		key, err := t.w.GetPrivateKeyForPubkey(pubkey)
		if err != nil {
			// Print message and continue because we don't want to stop the loop
			t.logger.Warn("Error getting private key", slog.String(PubkeyKey, pubkey.HexWithPrefix()), log.Err(err))
			continue
		}
		if key == nil {
			t.logger.Warn("Private key not found", slog.String(PubkeyKey, pubkey.HexWithPrefix()))
			continue
		}
		signature, err := validator.GetSignedExitMessage(key, index, epoch, signatureDomain)
		if err != nil {
			// Print message and continue because we don't want to stop the loop
			// Index might not be ready
			t.logger.Debug("Error getting signed exit message", slog.String(PubkeyKey, pubkey.HexWithPrefix()), log.Err(err))
			continue
		}
		exitMessage := common.ExitMessage{
			Message: common.ExitMessageDetails{
				Epoch:          strconv.FormatUint(epoch, 10),
				ValidatorIndex: index,
			},
			Signature: signature.HexWithPrefix(),
		}

		// Encrypt it
		encryptedMessage, err := common.EncryptSignedExitMessage(exitMessage, t.res.EncryptionPubkey)
		if err != nil {
			t.logger.Warn("Error encrypting signed exit message",
				slog.String("pubkey", pubkey.HexWithPrefix()),
				log.Err(err),
			)
			continue
		}

		exitData = append(exitData, common.EncryptedExitData{
			Pubkey:      pubkey.HexWithPrefix(),
			ExitMessage: encryptedMessage,
		})
	}

	// Upload the messages to Nodeset
	if len(exitData) > 0 {
		_, err := t.hd.NodeSet_StakeWise.UploadSignedExits(t.res.DeploymentName, t.res.Vault, exitData)
		if err != nil {
			return fmt.Errorf("error uploading signed exit messages to NodeSet: %w", err)
		}

		pubkeys := []any{}
		for _, d := range exitData {
			pubkeys = append(pubkeys, slog.String(PubkeyKey, d.Pubkey))
		}
		t.logger.Info("Uploaded exit messages", pubkeys...)
	}

	return nil
}
