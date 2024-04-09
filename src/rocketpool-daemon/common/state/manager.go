package state

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	nmc_config "github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

type NetworkStateManager struct {
	cfg          *config.SmartNodeConfig
	rp           *rocketpool.RocketPool
	ec           eth.IExecutionClient
	bc           beacon.IBeaconClient
	logger       *slog.Logger
	Config       *config.SmartNodeConfig
	Network      nmc_config.Network
	ChainID      uint
	BeaconConfig beacon.Eth2Config
}

// Create a new manager for the network state
func NewNetworkStateManager(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, ec eth.IExecutionClient, bc beacon.IBeaconClient, logger *slog.Logger) (*NetworkStateManager, error) {
	// Make a resource list
	resources := cfg.GetNetworkResources()

	// Create the manager
	m := &NetworkStateManager{
		cfg:     cfg,
		rp:      rp,
		ec:      ec,
		bc:      bc,
		logger:  logger,
		Network: cfg.Network.Value,
		ChainID: resources.ChainID,
	}

	// Get the Beacon config info
	var err error
	m.BeaconConfig, err = m.bc.GetEth2Config(context)
	if err != nil {
		return nil, err
	}

	return m, nil

}

// Get the state of the network using the latest Execution layer block
func (m *NetworkStateManager) GetHeadState(context context.Context) (*NetworkState, error) {
	targetSlot, err := m.GetHeadSlot()
	if err != nil {
		return nil, fmt.Errorf("error getting latest Beacon slot: %w", err)
	}
	return m.getState(context, targetSlot)
}

// Get the state of the network for a single node using the latest Execution layer block, along with the total effective RPL stake for the network
func (m *NetworkStateManager) GetHeadStateForNode(context context.Context, nodeAddress common.Address, calculateTotalEffectiveStake bool) (*NetworkState, *big.Int, error) {
	targetSlot, err := m.GetHeadSlot()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting latest Beacon slot: %w", err)
	}
	return m.getStateForNode(context, nodeAddress, targetSlot, calculateTotalEffectiveStake)
}

// Get the state of the network at the provided Beacon slot
func (m *NetworkStateManager) GetStateForSlot(context context.Context, slotNumber uint64) (*NetworkState, error) {
	return m.getState(context, slotNumber)
}

// Get the state of the network for a single node at the provided Beacon slot, along with the total effective RPL stake for the network
func (m *NetworkStateManager) GetNodeStateForSlot(context context.Context, nodeAddress common.Address, slotNumber uint64, calculateTotalEffectiveStake bool) (*NetworkState, *big.Int, error) {
	return m.getStateForNode(context, nodeAddress, slotNumber, calculateTotalEffectiveStake)
}

// Gets the latest valid block
func (m *NetworkStateManager) GetLatestBeaconBlock(context context.Context) (beacon.BeaconBlock, error) {
	targetSlot, err := m.GetHeadSlot()
	if err != nil {
		return beacon.BeaconBlock{}, fmt.Errorf("error getting head slot: %w", err)
	}
	return m.GetLatestProposedBeaconBlock(context, targetSlot)
}

// Gets the latest valid finalized block
func (m *NetworkStateManager) GetLatestFinalizedBeaconBlock(context context.Context) (beacon.BeaconBlock, error) {
	head, err := m.bc.GetBeaconHead(context)
	if err != nil {
		return beacon.BeaconBlock{}, fmt.Errorf("error getting Beacon chain head: %w", err)
	}
	targetSlot := head.FinalizedEpoch*m.BeaconConfig.SlotsPerEpoch + (m.BeaconConfig.SlotsPerEpoch - 1)
	return m.GetLatestProposedBeaconBlock(context, targetSlot)
}

// Gets the Beacon slot for the latest execution layer block
func (m *NetworkStateManager) GetHeadSlot() (uint64, error) {
	// Get the latest EL block
	latestBlockHeader, err := m.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest EL block: %w", err)
	}

	// Get the corresponding Beacon slot based on the timestamp
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	genesisTime := time.Unix(int64(m.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := uint64(latestBlockTime.Sub(genesisTime).Seconds())
	targetSlot := secondsSinceGenesis / m.BeaconConfig.SecondsPerSlot
	return targetSlot, nil
}

// Gets the target Beacon block, or if it was missing, the first one under it that wasn't missing
func (m *NetworkStateManager) GetLatestProposedBeaconBlock(context context.Context, targetSlot uint64) (beacon.BeaconBlock, error) {
	for {
		// Try to get the current block
		block, exists, err := m.bc.GetBeaconBlock(context, fmt.Sprint(targetSlot))
		if err != nil {
			return beacon.BeaconBlock{}, fmt.Errorf("error getting Beacon block %d: %w", targetSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			m.logger.Info("Slot was missing, trying the previous one...", slog.Uint64(keys.SlotKey, targetSlot))
			targetSlot--
		} else {
			return block, nil
		}
	}
}

// Get the state of the network at the provided Beacon slot
func (m *NetworkStateManager) getState(context context.Context, slotNumber uint64) (*NetworkState, error) {
	state, err := CreateNetworkState(m.cfg, m.rp, m.ec, m.bc, m.logger, slotNumber, m.BeaconConfig, context)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Get the state of the network for a specific node only at the provided Beacon slot
func (m *NetworkStateManager) getStateForNode(context context.Context, nodeAddress common.Address, slotNumber uint64, calculateTotalEffectiveStake bool) (*NetworkState, *big.Int, error) {
	state, totalEffectiveStake, err := CreateNetworkStateForNode(m.cfg, m.rp, m.ec, m.bc, m.logger, slotNumber, m.BeaconConfig, nodeAddress, calculateTotalEffectiveStake, context)
	if err != nil {
		return nil, nil, err
	}
	return state, totalEffectiveStake, nil
}
