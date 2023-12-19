package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/api/security"
	"github.com/rocket-pool/smartnode/rocketpool/api/service"
	"github.com/rocket-pool/smartnode/rocketpool/api/tx"
	"github.com/rocket-pool/smartnode/rocketpool/common/log"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"

	"github.com/rocket-pool/smartnode/rocketpool/api/auction"
	"github.com/rocket-pool/smartnode/rocketpool/api/faucet"
	"github.com/rocket-pool/smartnode/rocketpool/api/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/api/network"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/odao"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/rocketpool/api/queue"
	"github.com/rocket-pool/smartnode/rocketpool/api/wallet"
)

const (
	ApiLogColor color.Attribute = color.FgHiBlue
)

type IHandler interface {
	RegisterRoutes(router *mux.Router)
}

type ApiManager struct {
	log        log.ColorLogger
	handlers   []IHandler
	socketPath string
	socket     net.Listener
	server     http.Server
	router     *mux.Router
}

func NewApiManager(sp *services.ServiceProvider) *ApiManager {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	cfg := sp.GetConfig()
	mgr := &ApiManager{
		log: log.NewColorLogger(ApiLogColor),
		handlers: []IHandler{
			auction.NewAuctionHandler(sp),
			faucet.NewFaucetHandler(sp),
			minipool.NewMinipoolHandler(sp),
			network.NewNetworkHandler(sp),
			node.NewNodeHandler(sp),
			odao.NewOracleDaoHandler(sp),
			pdao.NewProtocolDaoHandler(sp),
			queue.NewQueueHandler(sp),
			security.NewSecurityCouncilHandler(sp),
			service.NewServiceHandler(sp),
			tx.NewTxHandler(sp),
			wallet.NewWalletHandler(sp),
		},
		socketPath: cfg.Smartnode.GetSocketPath(),
		router:     router,
		server: http.Server{
			Handler: router,
		},
	}

	// Register each route
	smartnodeRouter := router.Host("rocketpool").Subrouter()
	for _, handler := range mgr.handlers {
		handler.RegisterRoutes(smartnodeRouter)
	}

	return mgr
}

// Starts listening for incoming HTTP requests
func (m *ApiManager) Start() error {
	// Create the socket
	socket, err := net.Listen("unix", m.socketPath)
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	m.socket = socket

	// Start listening
	go func() {
		err := m.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			m.log.Printlnf("error while listening for HTTP requests: %s", err.Error())
		}
	}()

	return nil
}

// Stops the HTTP listener
func (m *ApiManager) Stop() error {
	// Shutdown the listener
	err := m.server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("error stopping listener: %w", err)
	}

	// Remove the socket file
	err = os.Remove(m.socketPath)
	if err != nil {
		return fmt.Errorf("error removing socket file: %w", err)
	}

	return nil
}
