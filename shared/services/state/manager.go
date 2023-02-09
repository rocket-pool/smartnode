package state

import (
	"fmt"
	"sync"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type NetworkStateManager struct {
	cfg          *config.RocketPoolConfig
	rp           *rocketpool.RocketPool
	ec           rocketpool.ExecutionClient
	bc           beacon.Client
	log          *log.ColorLogger
	Config       *config.RocketPoolConfig
	Network      cfgtypes.Network
	ChainID      uint
	BeaconConfig beacon.Eth2Config
	latestState  *NetworkState
	updateLock   sync.Mutex
}

// Create a new manager for the network state
func NewNetworkStateManager(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, ec rocketpool.ExecutionClient, bc beacon.Client, log *log.ColorLogger) (*NetworkStateManager, error) {

	// Create the manager
	m := &NetworkStateManager{
		cfg:     cfg,
		rp:      rp,
		ec:      ec,
		bc:      bc,
		log:     log,
		Config:  cfg,
		Network: cfg.Smartnode.Network.Value.(cfgtypes.Network),
		ChainID: cfg.Smartnode.GetChainID(),
	}

	// Get the Beacon config info
	var err error
	m.BeaconConfig, err = m.bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	return m, nil

}

// Logs a line if the logger is specified
func (m *NetworkStateManager) logLine(format string, v ...interface{}) {
	if m.log != nil {
		m.log.Printlnf(format, v)
	}
}

// Get the state of the network at the provided Beacon slot
func (m *NetworkStateManager) UpdateState(slotNumber *uint64) (*NetworkState, error) {
	m.updateLock.Lock()
	defer m.updateLock.Unlock()

	if slotNumber == nil {
		// Get the latest finalized slot
		head, err := m.bc.GetBeaconHead()
		if err != nil {
			return nil, fmt.Errorf("error getting latest finalized slot: %w", err)
		}
		targetSlot := head.FinalizedEpoch*m.BeaconConfig.SlotsPerEpoch + (m.BeaconConfig.SlotsPerEpoch - 1)

		// If that slot is missing, get the latest one that isn't
		for {
			// Try to get the current block
			_, exists, err := m.bc.GetBeaconBlock(fmt.Sprint(targetSlot))
			if err != nil {
				return nil, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
			}

			// If the block was missing, try the previous one
			if !exists {
				m.logLine("Slot %d was missing, trying the previous one...", targetSlot)
				targetSlot--
			} else {
				slotNumber = &targetSlot
				break
			}
		}
	}

	state, err := CreateNetworkState(m.cfg, m.rp, m.ec, m.bc, m.log, *slotNumber, m.BeaconConfig)
	if err != nil {
		return nil, err
	}

	m.latestState = state
	return state, nil
}

// Gets the latest state in a thread-safe manner
func (m *NetworkStateManager) GetLatestState() *NetworkState {
	m.updateLock.Lock()
	defer m.updateLock.Unlock()
	return m.latestState
}
