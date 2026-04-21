package services

import (
	"fmt"
	"sync"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services/state"
)

// Memoized static snapshot / provider so that every service accessor
// consults the same in-memory NetworkState without re-reading the file.
var (
	staticState          *state.NetworkState
	staticStateErr       error
	initStaticState      sync.Once
	networkStateProv     state.NetworkStateProvider
	networkStateProvErr  error
	initNetworkStateProv sync.Once
)

// GetStaticStatePath returns the --network-state flag value on the root
// command, or the empty string if not set.
func GetStaticStatePath(c *cli.Command) string {
	if c == nil {
		return ""
	}
	return c.Root().String("network-state")
}

// IsStaticStateMode reports whether the daemon has been asked to serve
// requests from a saved NetworkState snapshot instead of live EC/CC clients.
func IsStaticStateMode(c *cli.Command) bool {
	return GetStaticStatePath(c) != ""
}

// getStaticState loads the NetworkState snapshot pointed at by --network-state,
// memoizing the result on success. It is safe to call concurrently.
func getStaticState(c *cli.Command) (*state.NetworkState, error) {
	path := GetStaticStatePath(c)
	if path == "" {
		return nil, fmt.Errorf("static state mode is not enabled (--network-state is unset)")
	}
	initStaticState.Do(func() {
		provider, err := state.NewStaticNetworkStateProviderFromFile(path)
		if err != nil {
			staticStateErr = fmt.Errorf("loading static network state from %q: %w", path, err)
			return
		}
		ns, err := provider.GetHeadState()
		if err != nil {
			staticStateErr = fmt.Errorf("reading head state from %q: %w", path, err)
			return
		}
		staticState = ns
	})
	return staticState, staticStateErr
}

// GetNetworkStateProvider returns a state.NetworkStateProvider backed by
// either the live NetworkStateManager (dialling the configured EC / CC) or
// by a StaticNetworkStateProvider loaded from --network-state, depending on
// the command-line flag.
//
// In live mode, the returned provider reuses the memoized execution and
// consensus clients from GetRocketPool / GetBeaconClient so the daemon
// doesn't establish extra connections.
func GetNetworkStateProvider(c *cli.Command) (state.NetworkStateProvider, error) {
	initNetworkStateProv.Do(func() {
		if IsStaticStateMode(c) {
			ns, err := getStaticState(c)
			if err != nil {
				networkStateProvErr = err
				return
			}
			networkStateProv = state.NewStaticNetworkStateProvider(ns)
			return
		}

		cfg, err := GetConfig(c)
		if err != nil {
			networkStateProvErr = err
			return
		}
		rp, err := GetRocketPool(c)
		if err != nil {
			networkStateProvErr = err
			return
		}
		bc, err := GetBeaconClient(c)
		if err != nil {
			networkStateProvErr = err
			return
		}
		networkStateProv = state.NewNetworkStateManager(rp, cfg.Smartnode.GetStateManagerContracts(), bc, nil)
	})
	return networkStateProv, networkStateProvErr
}
