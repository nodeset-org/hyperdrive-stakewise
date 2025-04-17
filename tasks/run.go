package swtasks

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	swcommon "github.com/nodeset-org/hyperdrive-stakewise/common"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
)

// Config
const (
	// Time to wait after finishing tasks before starting the next iteration
	tasksInterval time.Duration = time.Minute * 5

	// Time between individual tasks
	//taskCooldown time.Duration = time.Second

	// Time to wait if the tasks loop isn't ready before checking again
	notReadySleepTime time.Duration = time.Second * 15

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
	sp     swcommon.IStakeWiseServiceProvider
	wg     *sync.WaitGroup

	// Tasks

	// Internal
	wasExecutionClientSynced bool
	wasBeaconClientSynced    bool
}

func NewTaskLoop(sp swcommon.IStakeWiseServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	logger := sp.GetTasksLogger()
	ctx := logger.CreateContextWithLogger(sp.GetBaseContext())
	taskLoop := &TaskLoop{
		sp:     sp,
		logger: logger,
		ctx:    ctx,
		wg:     wg,

		wasExecutionClientSynced: true,
		wasBeaconClientSynced:    true,
	}
	return taskLoop
}

// Run daemon
func (t *TaskLoop) Run() error {
	// Wait until the HD daemon has tried logging into the NodeSet server to check registration status
	//t.getNodeSetRegistrationStatus()

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

// Get the NodeSet server registration status
/*
func (t *TaskLoop) getNodeSetRegistrationStatus() {
	hd := t.sp.GetHyperdriveClient()
	attempts := 3
	for i := 0; i < attempts; i++ {
		response, err := hd.NodeSet.GetRegistrationStatus()
		if err != nil {
			// Error was because of a comms failure, so try again after 1 second
			t.logger.Warn(
				"Getting node registration status during NodeSet login attempt failed",
				slog.String(log.ErrorKey, err.Error()),
				slog.Int(keys.AttemptKey, i+1),
			)
			if utils.SleepWithCancel(t.ctx, time.Second) {
				return
			}
			continue
		}

		switch response.Data.Status {
		case api.NodeSetRegistrationStatus_Registered:
			// Successful login
			return
		case api.NodeSetRegistrationStatus_NoWallet:
			// Error was because the wallet isn't ready yet, so just return since logging in won't work yet
			t.logger.Info("Can't log into NodeSet, node doesn't have a wallet yet")
			return
		case api.NodeSetRegistrationStatus_Unregistered:
			// Node's not registered yet, this isn't an actual error to report
			t.logger.Info("Node is not registered with NodeSet yet")
			return
		default:
			// Error occurred on the remote side, so try again after 1 second
			t.logger.Warn(
				"NodeSet registration status is unknown",
				slog.String(log.ErrorKey, response.Data.ErrorMessage),
				slog.Int(keys.AttemptKey, i+1),
			)
			if utils.SleepWithCancel(t.ctx, time.Second) {
				return
			}
		}
	}
	t.logger.Error("Max login attempts reached")
}
*/

// Wait until the chains and other resources are ready to be queried
// Returns true if the owning loop needs to exit, false if it can continue
func (t *TaskLoop) waitUntilReady() waitUntilReadyResult {
	// Check the EC status
	// NOTE: Not needed with StakeWise v2 now
	/*
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
	*/

	// Check the BC status
	// NOTE: Not needed with StakeWise v2 now
	/*
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
	*/

	// Wait until the Stakewise wallet has been initialized
	err := t.sp.WaitForStakewiseWallet(t.ctx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		t.logger.Error("Error waiting for Stakewise wallet initialization", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	// Wait for NodeSet registration
	// Not needed with StakeWise v2 now
	/*
		if t.sp.WaitForNodeSetRegistration(t.ctx) {
			return waitUntilReadyExit
		}
	*/

	return waitUntilReadySuccess
}

// Sleep on the context for the not-ready sleep time, and return either exit or continue
// based on whether the context was cancelled.
func (t *TaskLoop) sleepAndReturnReadyResult() waitUntilReadyResult {
	if utils.SleepWithCancel(t.ctx, notReadySleepTime) {
		return waitUntilReadyExit
	} else {
		return waitUntilReadyContinue
	}
}

// Runs an iteration of the node tasks.
// Returns true if the task loop should exit, false if it should continue.
func (t *TaskLoop) runTasks() bool {
	return utils.SleepWithCancel(t.ctx, tasksInterval)
}
