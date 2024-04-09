package api

import (
	"fmt"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/auction"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/faucet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/minipool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/network"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/odao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/pdao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/queue"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/security"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/service"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

const (
	cliOrigin string = "cli"
	webOrigin string = "net"
)

// ServerManager manages all of the daemon sockets and servers run by the main Smart Node daemon
type ServerManager struct {
	// The server for the CLI to interact with
	cliServer *server.ApiServer
}

// Creates a new server manager
func NewServerManager(sp *services.ServiceProvider, cfgPath string, stopWg *sync.WaitGroup) (*ServerManager, error) {
	// Get the owner of the config file
	var cfgFileStat syscall.Stat_t
	err := syscall.Stat(cfgPath, &cfgFileStat)
	if err != nil {
		return nil, fmt.Errorf("error getting config file [%s] info: %w", cfgPath, err)
	}

	// Start the CLI server
	cliSocketPath := filepath.Join(sp.GetUserDir(), config.SmartNodeCliSocketFilename)
	cliServer, err := createServer(cliOrigin, sp, cliSocketPath)
	if err != nil {
		return nil, fmt.Errorf("error creating CLI server: %w", err)
	}
	err = cliServer.Start(stopWg, cfgFileStat.Uid, cfgFileStat.Gid)
	if err != nil {
		return nil, fmt.Errorf("error starting CLI server: %w", err)
	}
	fmt.Printf("CLI daemon started on %s\n", cliSocketPath)

	// Create the manager
	mgr := &ServerManager{
		cliServer: cliServer,
	}
	return mgr, nil
}

// Stops and shuts down the servers
func (m *ServerManager) Stop() {
	err := m.cliServer.Stop()
	if err != nil {
		fmt.Printf("WARNING: CLI server didn't shutdown cleanly: %s\n", err.Error())
	}
}

// Creates a new Smart Node API server
func createServer(origin string, sp *services.ServiceProvider, socketPath string) (*server.ApiServer, error) {
	apiLogger := sp.GetApiLogger()
	subLogger := apiLogger.CreateSubLogger(origin)
	ctx := subLogger.CreateContextWithLogger(sp.GetBaseContext())

	handlers := []server.IHandler{
		auction.NewAuctionHandler(subLogger, ctx, sp),
		faucet.NewFaucetHandler(subLogger, ctx, sp),
		minipool.NewMinipoolHandler(subLogger, ctx, sp),
		network.NewNetworkHandler(subLogger, ctx, sp),
		node.NewNodeHandler(subLogger, ctx, sp),
		odao.NewOracleDaoHandler(subLogger, ctx, sp),
		pdao.NewProtocolDaoHandler(subLogger, ctx, sp),
		queue.NewQueueHandler(subLogger, ctx, sp),
		security.NewSecurityCouncilHandler(subLogger, ctx, sp),
		service.NewServiceHandler(subLogger, ctx, sp),
		tx.NewTxHandler(subLogger, ctx, sp),
		wallet.NewWalletHandler(subLogger, ctx, sp),
	}

	server, err := server.NewApiServer(subLogger.Logger, socketPath, handlers, config.SmartNodeDaemonBaseRoute, config.SmartNodeApiVersion)
	if err != nil {
		return nil, err
	}
	return server, nil
}
