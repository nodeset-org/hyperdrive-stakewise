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
	sp     *swcommon.StakewiseServiceProvider
	wg     *sync.WaitGroup

	// Tasks
	updateDepositData *UpdateDepositDataTask
	sendExitData      *SendExitDataTask

	// Internal
	wasExecutionClientSynced bool
	wasBeaconClientSynced    bool
}

func NewTaskLoop(sp *swcommon.StakewiseServiceProvider, wg *sync.WaitGroup) *TaskLoop {
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
