package collectors

import (
	"sync"

	"github.com/rocket-pool/smartnode/shared/services/state"
)

type StateLocker struct {
	state *state.NetworkState

	// Internal fields
	lock *sync.RWMutex
}

func NewStateLocker() *StateLocker {
	return &StateLocker{
		lock: &sync.RWMutex{},
	}
}

func (l *StateLocker) UpdateState(state *state.NetworkState) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.state = state
}

func (l *StateLocker) GetState() *state.NetworkState {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.state
}
