package services

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	watcher "github.com/rocket-pool/smartnode/shared/services/performancewatcher"
	"github.com/urfave/cli/v3"
)

var performanceWatcher atomic.Pointer[watcher.Worker]

// GetPerformanceWatcher only returns the daemon-owned worker; API reads never
// start a worker or open a second copy of its database.
func GetPerformanceWatcher() *watcher.Worker { return performanceWatcher.Load() }

func StartPerformanceWatcher(ctx context.Context, c *cli.Command) (func(), error) {
	cfg, err := GetConfig(c)
	if err != nil {
		return nil, err
	}
	noop := func() {}
	if cfg.Smartnode.RunPerformanceWatcher.Value != true || IsStaticStateMode(c) {
		return noop, nil
	}
	days := cfg.Smartnode.PerformanceWatcherBootstrapRetentionDays.Value.(uint64)
	if days == 0 {
		return nil, fmt.Errorf("Pre-Saturn 2 Retention Days must be positive")
	}
	rp, err := GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	clients := []watcher.Requester{}
	budget := make(watcherBudget, 4)
	for _, client := range []interface{}{bc.primaryBc, bc.fallbackBc} {
		if r, ok := client.(watcher.Requester); ok {
			clients = append(clients, watcherRequester{r, budget})
		}
	}
	if bc.preferFallback && len(clients) == 2 {
		clients[0], clients[1] = clients[1], clients[0]
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("configured beacon clients do not support watcher reads")
	}
	ec, err := GetEthClient(c)
	if err != nil {
		return nil, err
	}
	// Use the configured transports without sharing the task loop's mutable
	// primaryReady/fallbackReady flags with the independent worker.
	var primary rocketpool.ExecutionClient = ec.primaryEc
	var fallback rocketpool.ExecutionClient
	if ec.fallbackEc != nil {
		fallback = ec.fallbackEc
	}
	if ec.preferFallback && fallback != nil {
		primary, fallback = fallback, primary
	}
	source := &watcherSource{ec: watcherRPCClient{ExecutionClient: primary, fallback: fallback, budget: budget, chainID: uint64(cfg.Smartnode.GetChainID()), verified: new(sync.Map)}, storage: *rp.RocketStorageContract.Address, multicall: common.HexToAddress(cfg.Smartnode.GetMulticallAddress()), chainID: uint64(cfg.Smartnode.GetChainID())}
	path := filepath.Join(cfg.Smartnode.GetRecordsPath(), "performance-watcher", fmt.Sprintf("%d-%s", cfg.Smartnode.GetChainID(), source.storage.Hex()))
	b := watcher.NewBeacon(clients)
	b.ExpectedChainID = source.chainID
	w := watcher.New(path, days, cfg.Smartnode.PerformanceWatcherBackfill.Value == true, b, source, func(format string, args ...any) { fmt.Printf(format+"\n", args...) })
	performanceWatcher.Store(w)
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); w.Run(workerCtx) }()
	return func() { cancel(); <-done; performanceWatcher.CompareAndSwap(w, nil) }, nil
}
