package swtasks

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	swapi "github.com/nodeset-org/hyperdrive-stakewise/shared/api"
	"github.com/nodeset-org/hyperdrive-stakewise/shared/keys"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
)

// Config
const (
	tasksInterval time.Duration = time.Minute * 5
	taskCooldown  time.Duration = time.Second * 10

	ErrorColor             = color.FgRed
	WarningColor           = color.FgYellow
	UpdateDepositDataColor = color.FgHiWhite
	SendExitDataColor      = color.FgGreen
)

type waitUntilReadyResult int

const (
	waitUntilReadyExit waitUntilReadyResult = iota
	waitUntilReadyContinue
	waitUntilReadySuccess
)

type TaskLoop struct {
	// Services
	ctx    context.Context
	logger *log.Logger
	sp     *swcommon.StakeWiseServiceProvider
	wg     *sync.WaitGroup

	// Tasks
	updateDepositData *UpdateDepositDataTask
	sendExitData      *SendExitDataTask

	// Internal
	wasExecutionClientSynced bool
	wasBeaconClientSynced    bool
}

func NewTaskLoop(sp *swcommon.StakeWiseServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	logger := sp.GetTasksLogger()
	ctx := logger.CreateContextWithLogger(sp.GetBaseContext())
	taskLoop := &TaskLoop{
		sp:                sp,
		logger:            logger,
		ctx:               ctx,
		wg:                wg,
		updateDepositData: NewUpdateDepositDataTask(ctx, sp, logger),
		sendExitData:      NewSendExitDataTask(ctx, sp, logger),

		wasExecutionClientSynced: true,
		wasBeaconClientSynced:    true,
	}
	return taskLoop
}

// Run daemon
func (t *TaskLoop) Run() error {
	// Log into the NodeSet server to check registration status
	t.logIntoNodeSet()

	// Run task loop
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		for {
			// Make sure all of the resources are ready for task processing
			readyResult := t.waitUntilReady()
			switch readyResult {
			case waitUntilReadyExit:
				return
			case waitUntilReadyContinue:
				continue
			}

			// === Task execution ===
			if t.runTasks() {
				return
			}
		}
	}()

	/*
		// Run metrics loop
		go func() {
			err := runMetricsServer(sp, log.NewColorLogger(MetricsColor), stateLocker)
			if err != nil {
				errorLog.Println(err)
			}
			wg.Done()
		}()
	*/
	return nil
}

// Log into the NodeSet server to check registration status
func (t *TaskLoop) logIntoNodeSet() {
	ns := t.sp.GetNodesetClient()
	attempts := 3
	for i := 0; i < attempts; i++ {
		status, err := ns.GetNodeRegistrationStatus(t.ctx)
		switch status {
		case swapi.NodesetRegistrationStatus_Registered:
			// Successful login
			return
		case swapi.NodesetRegistrationStatus_NoWallet:
			// Error was because the wallet isn't ready yet, so just return since logging in won't work yet
			t.logger.Info("Can't log in, node doesn't have a wallet yet")
			return
		case swapi.NodesetRegistrationStatus_Unregistered:
			// Node's not registered yet, this isn't an actual error to report
			t.logger.Info("Can't log in, node is not registered with NodeSet yet")
			return
		default:
			// Error was because of a comms failure, so try again after 1 second
			t.logger.Warn(
				"Getting node registration status during NodeSet login attempt failed",
				slog.String(log.ErrorKey, err.Error()),
				slog.Int(keys.AttemptKey, i+1),
			)
			if utils.SleepWithCancel(t.ctx, time.Second) {
				return
			}
		}

	}
	t.logger.Error("Max login attempts reached")
}

// Wait until the chains and other resources are ready to be queried
// Returns true if the owning loop needs to exit, false if it can continue
func (t *TaskLoop) waitUntilReady() waitUntilReadyResult {
	// Check the EC status
	err := t.sp.WaitEthClientSynced(t.ctx, false) // Force refresh the primary / fallback EC status
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		t.wasExecutionClientSynced = false
		t.logger.Error("Execution Client not synced. Waiting for sync...", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	if !t.wasExecutionClientSynced {
		t.logger.Info("Execution Client is now synced.")
		t.wasExecutionClientSynced = true
	}

	// Check the BC status
	err = t.sp.WaitBeaconClientSynced(t.ctx, false) // Force refresh the primary / fallback BC status
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		// NOTE: if not synced, it returns an error - so there isn't necessarily an underlying issue
		t.wasBeaconClientSynced = false
		t.logger.Error("Beacon Node not synced. Waiting for sync...", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	if !t.wasBeaconClientSynced {
		t.logger.Info("Beacon Node is now synced.")
		t.wasBeaconClientSynced = true
	}

	// Wait until the Stakewise wallet has been initialized
	err = t.sp.WaitForStakewiseWallet(t.ctx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		t.logger.Error("Error waiting for Stakewise wallet initialization", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	// Wait for NodeSet registration
	if t.sp.WaitForNodeSetRegistration(t.ctx) {
		return waitUntilReadyExit
	}

	return waitUntilReadySuccess
}

// Sleep on the context for the task cooldown time, and return either exit or continue
// based on whether the context was cancelled.
func (t *TaskLoop) sleepAndReturnReadyResult() waitUntilReadyResult {
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return waitUntilReadyExit
	} else {
		return waitUntilReadyContinue
	}
}

// Runs an iteration of the node tasks.
// Returns true if the task loop should exit, false if it should continue.
func (t *TaskLoop) runTasks() bool {
	// Update deposit data from the NodeSet server
	if err := t.updateDepositData.Run(); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Submit missing exit messages to the NodeSet server
	if err := t.sendExitData.Run(); err != nil {
		t.logger.Error(err.Error())
	}

	return utils.SleepWithCancel(t.ctx, tasksInterval)
}
