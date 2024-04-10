package task

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
)

var (
	// ErrAlreadyRunning is returned when a background task is kicked off, but it is already in progress.
	ErrAlreadyRunning = errors.New("task is already running")
)

// TaskContext is passed to the Task's Callback function when the invoker wishes the task
// to be kicked off.
//
// Its fields are things that are variable and may change between invokations of a task.
type BackgroundTaskContext struct {
	// A context provided by the invoker of this task.
	// May be nil, and cancellations should be respected.
	Ctx context.Context
	// Whether or not the node is on the oDAO at the time the task was invoked
	IsOnOdao bool
	// A recent network state so each task need not query it redundantly
	State *state.NetworkState
}

type BackgroundTask interface {
	// Returns a function to call that starts the task in the background
	Run(*BackgroundTaskContext) error
}

type LockingBackgroundTask struct {
	// Done should be called on successful completion of the task to release the lock
	Done func()

	logger      *slog.Logger
	description string
	run         func(*BackgroundTaskContext) error

	lock      sync.Mutex
	isRunning bool
}

// NewLockingBackgroundTask creates a background task that only allows one instance of itself to run
// logger and description arguments will be used to log whether the task was kicked off or blocked
// by a concurrnet instance.
// f is the task function that will be called, and shoud take care to either return an error or call Done() when finished.
func NewLockingBackgroundTask(logger *slog.Logger, description string, f func(*BackgroundTaskContext) error) *LockingBackgroundTask {
	out := &LockingBackgroundTask{
		description: description,
		logger:      logger,
		run:         f,
	}
	out.Done = func() {
		out.lock.Lock()
		defer out.lock.Unlock()
		out.isRunning = false
	}
	return out
}

func (lbt *LockingBackgroundTask) Run(taskContext *BackgroundTaskContext) error {
	lbt.lock.Lock()
	defer lbt.lock.Unlock()

	lbt.logger.Info("Starting task", "description", lbt.description)
	if lbt.isRunning {
		lbt.logger.Info("Task is already running", "description", lbt.description)
		return ErrAlreadyRunning
	}

	lbt.isRunning = true
	err := lbt.run(taskContext)
	if err != nil {
		// Done is safe to repeat, so we can make it optional when run returns an error
		lbt.Done()
	}
	return err
}
