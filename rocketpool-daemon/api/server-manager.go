package api

import (
	"fmt"
	"sync"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/api/auction"
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

// ServerManager manages all of the daemon sockets and servers run by the main Smart Node daemon
type ServerManager struct {
	// The server for clients to interact with
	apiServer *server.NetworkSocketApiServer
}

// Creates a new server manager
func NewServerManager(sp *services.ServiceProvider, ip string, port uint16, stopWg *sync.WaitGroup) (*ServerManager, error) {
	// Start the API server
	apiServer, err := createServer(sp, ip, port)
	if err != nil {
		return nil, fmt.Errorf("error creating API server: %w", err)
	}
	err = apiServer.Start(stopWg)
	if err != nil {
		return nil, fmt.Errorf("error starting API server: %w", err)
	}
	fmt.Printf("API server started on %s:%d\n", ip, port)

	// Create the manager
	mgr := &ServerManager{
		apiServer: apiServer,
	}
	return mgr, nil
}

// Stops and shuts down the servers
func (m *ServerManager) Stop() {
	err := m.apiServer.Stop()
	if err != nil {
		fmt.Printf("WARNING: API server didn't shutdown cleanly: %s\n", err.Error())
	}
}

// Creates a new Smart Node API server
func createServer(sp *services.ServiceProvider, ip string, port uint16) (*server.NetworkSocketApiServer, error) {
	apiLogger := sp.GetApiLogger()
	ctx := apiLogger.CreateContextWithLogger(sp.GetBaseContext())

	handlers := []server.IHandler{
		auction.NewAuctionHandler(apiLogger, ctx, sp),
		minipool.NewMinipoolHandler(apiLogger, ctx, sp),
		network.NewNetworkHandler(apiLogger, ctx, sp),
		node.NewNodeHandler(apiLogger, ctx, sp),
		odao.NewOracleDaoHandler(apiLogger, ctx, sp),
		pdao.NewProtocolDaoHandler(apiLogger, ctx, sp),
		queue.NewQueueHandler(apiLogger, ctx, sp),
		security.NewSecurityCouncilHandler(apiLogger, ctx, sp),
		service.NewServiceHandler(apiLogger, ctx, sp),
		tx.NewTxHandler(apiLogger, ctx, sp),
		wallet.NewWalletHandler(apiLogger, ctx, sp),
	}

	server, err := server.NewNetworkSocketApiServer(apiLogger.Logger, ip, port, handlers, config.SmartNodeDaemonBaseRoute, config.SmartNodeApiVersion)
	if err != nil {
		return nil, err
	}
	return server, nil
}
