package api

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/auction"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/faucet"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/minipool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/network"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/node"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/odao"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/pdao"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/queue"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/security"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/service"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/api/wallet"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/shared/config"
)

// ServerManager manages all of the daemon sockets and servers run by the main Smart Node daemon
type ServerManager struct {
	// The server for the CLI to interact with
	cliServer *server.ApiServer

	// The daemon's main closing waitgroup
	stopWg *sync.WaitGroup

	// Context for gracefully stopping API requests during shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// Creates a new server manager
func NewServerManager(sp *services.ServiceProvider, cfgPath string, stopWg *sync.WaitGroup) (*ServerManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	mgr := &ServerManager{
		stopWg: stopWg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Get the owner of the config file
	var cfgFileStat syscall.Stat_t
	err := syscall.Stat(cfgPath, &cfgFileStat)
	if err != nil {
		return nil, fmt.Errorf("error getting config file [%s] info: %w", cfgPath, err)
	}

	// Start the CLI server
	cliSocketPath := filepath.Join(sp.GetUserDir(), config.SmartNodeSocketFilename)
	cliServer, err := createServer(sp, cliSocketPath, ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating CLI server: %w", err)
	}
	err = cliServer.Start(stopWg, cfgFileStat.Uid, cfgFileStat.Gid)
	if err != nil {
		return nil, fmt.Errorf("error starting CLI server: %w", err)
	}
	mgr.cliServer = cliServer
	fmt.Printf("CLI daemon started on %s\n", cliSocketPath)

	return mgr, nil
}

// Stops and shuts down the servers
func (m *ServerManager) Stop() {
	m.cancel()
	err := m.cliServer.Stop()
	if err != nil {
		fmt.Printf("WARNING: CLI server didn't shutdown cleanly: %s\n", err.Error())
		m.stopWg.Done()
	}
}

// Creates a new Smart Node API server
func createServer(sp *services.ServiceProvider, socketPath string, ctx context.Context) (*server.ApiServer, error) {
	handlers := []server.IHandler{
		auction.NewAuctionHandler(ctx, sp),
		faucet.NewFaucetHandler(ctx, sp),
		minipool.NewMinipoolHandler(ctx, sp),
		network.NewNetworkHandler(ctx, sp),
		node.NewNodeHandler(ctx, sp),
		odao.NewOracleDaoHandler(ctx, sp),
		pdao.NewProtocolDaoHandler(ctx, sp),
		queue.NewQueueHandler(ctx, sp),
		security.NewSecurityCouncilHandler(ctx, sp),
		service.NewServiceHandler(ctx, sp),
		tx.NewTxHandler(ctx, sp),
		wallet.NewWalletHandler(ctx, sp),
	}

	server, err := server.NewApiServer(socketPath, handlers, config.SmartNodeDaemonRoute)
	if err != nil {
		return nil, err
	}
	return server, nil
}
