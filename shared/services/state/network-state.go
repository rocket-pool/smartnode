package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type NetworkStateManager struct {
	rp           *rocketpool.RocketPool
	ec           rocketpool.ExecutionClient
	bc           beacon.Client
	log          *log.ColorLogger
	Config       *config.RocketPoolConfig
	Network      cfgtypes.Network
	ChainID      uint
	BeaconConfig beacon.Eth2Config
}

type NetworkState struct {
	ElBlockNumber    uint64
	BeaconSlotNumber uint64

	// Insert various state variables as required

	// Node details
	NodeDetails          []node.NativeNodeDetails
	NodeDetailsByAddress map[common.Address]*node.NativeNodeDetails

	// Minipool details
	MinipoolDetails          []minipool.NativeMinipoolDetails
	MinipoolDetailsByAddress map[common.Address]*minipool.NativeMinipoolDetails
	MinipoolDetailsByNode    map[common.Address][]*minipool.NativeMinipoolDetails

	// Validator details
	ValidatorDetails map[types.ValidatorPubkey]beacon.ValidatorStatus
}

// Create a new manager for the network state
func NewNetworkStateManager(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, ec rocketpool.ExecutionClient, bc beacon.Client, log *log.ColorLogger) (*NetworkStateManager, error) {

	// Create the manager
	m := &NetworkStateManager{
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
func (m *NetworkStateManager) GetTotalState(slotNumber uint64) (*NetworkState, error) {

	// Get the execution block for the given slot
	beaconBlock, exists, err := m.bc.GetBeaconBlock(fmt.Sprintf("%d", slotNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon block for slot %s: %w", slotNumber, err)
	}
	if !exists {
		return nil, fmt.Errorf("slot %d did not have a Beacon block", slotNumber)
	}

	// Get the corresponding block on the EL
	elBlockNumber := beaconBlock.ExecutionBlockNumber
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlockNumber),
	}
	m.logLine("Getting network state for EL block %d, Beacon slot %d", elBlockNumber, slotNumber)
	start := time.Now()

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*node.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*minipool.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*minipool.NativeMinipoolDetails{},
	}

	// Node details
	nodeDetails, err := node.GetAllNativeNodeDetails(m.rp, opts)
	if err != nil {
		return nil, err
	}
	state.NodeDetails = nodeDetails
	m.logLine("1/4 - Retrieved node details (%s so far)", time.Since(start))

	// Minipool details
	minipoolDetails, err := minipool.GetAllNativeMinipoolDetails(m.rp, opts)
	if err != nil {
		return nil, err
	}
	m.logLine("2/4 - Retrieved minipool details (%s so far)", time.Since(start))

	// Create the node lookup
	for _, details := range nodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &details
	}

	// Create the minipool lookups
	pubkeys := make([]types.ValidatorPubkey, 0, len(minipoolDetails))
	emptyPubkey := types.ValidatorPubkey{}
	for _, details := range minipoolDetails {
		state.MinipoolDetailsByAddress[details.MinipoolAddress] = &details
		if details.Pubkey != emptyPubkey {
			pubkeys = append(pubkeys, details.Pubkey)
		}

		// The map of nodes to minipools
		nodeList, exists := state.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			nodeList = []*minipool.NativeMinipoolDetails{}
		}
		nodeList = append(nodeList, &details)
		state.MinipoolDetailsByNode[details.NodeAddress] = nodeList
	}
	m.logLine("3/4 - Created lookups (%s so far)", time.Since(start))

	// Get the validator stats from Beacon
	statusMap, err := m.bc.GetValidatorStatuses(pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.ValidatorDetails = statusMap
	m.logLine("4/4 - Retrieved validator details (total time: %s)", time.Since(start))

	return state, nil

}

//func (m *NetworkStateManager)
