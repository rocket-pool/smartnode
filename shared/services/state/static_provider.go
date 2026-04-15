package state

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// Compile-time check that StaticNetworkStateProvider satisfies NetworkStateProvider
var _ NetworkStateProvider = (*StaticNetworkStateProvider)(nil)

// StaticNetworkStateProvider serves a pre-loaded NetworkState without any
// EC/CC connections. Useful for deterministic tests driven by previously
// serialized state snapshots
type StaticNetworkStateProvider struct {
	state *NetworkState
}

func NewStaticNetworkStateProvider(ns *NetworkState) *StaticNetworkStateProvider {
	return &StaticNetworkStateProvider{state: ns}
}

func NewStaticNetworkStateProviderFromJSON(r io.Reader) (*StaticNetworkStateProvider, error) {
	var ns NetworkState
	if err := json.NewDecoder(r).Decode(&ns); err != nil {
		return nil, err
	}
	return NewStaticNetworkStateProvider(&ns), nil
}

// NewStaticNetworkStateProviderFromFile loads a NetworkState from a JSON file.
// If the path ends in ".gz", the file is transparently decompressed with gzip.
func NewStaticNetworkStateProviderFromFile(path string) (*StaticNetworkStateProvider, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening state file %q: %w", path, err)
	}
	defer f.Close()

	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader for %q: %w", path, err)
		}
		defer gz.Close()
		r = gz
	}
	return NewStaticNetworkStateProviderFromJSON(r)
}

func (p *StaticNetworkStateProvider) GetHeadState() (*NetworkState, error) {
	return p.state, nil
}

func (p *StaticNetworkStateProvider) GetHeadStateForNode(_ common.Address) (*NetworkState, error) {
	return p.state, nil
}

func (p *StaticNetworkStateProvider) GetStateForSlot(_ uint64) (*NetworkState, error) {
	return p.state, nil
}

func (p *StaticNetworkStateProvider) GetLatestBeaconBlock() (beacon.BeaconBlock, error) {
	return beacon.BeaconBlock{
		Slot:                 p.state.BeaconSlotNumber,
		HasExecutionPayload:  true,
		ExecutionBlockNumber: p.state.ElBlockNumber,
	}, nil
}

func (p *StaticNetworkStateProvider) GetLatestFinalizedBeaconBlock() (beacon.BeaconBlock, error) {
	return p.GetLatestBeaconBlock()
}
