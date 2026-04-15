package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// NetworkStateProvider abstracts the retrieval of network state snapshots
// NetworkStateManager satisfies this interface using live EC/CC connections
// StaticNetworkStateProvider satisfies it using a pre-loaded NetworkState
type NetworkStateProvider interface {
	GetHeadState() (*NetworkState, error)
	GetHeadStateForNode(nodeAddress common.Address) (*NetworkState, error)
	GetStateForSlot(slotNumber uint64) (*NetworkState, error)
	GetLatestBeaconBlock() (beacon.BeaconBlock, error)
	GetLatestFinalizedBeaconBlock() (beacon.BeaconBlock, error)
}
