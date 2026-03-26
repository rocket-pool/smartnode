package node

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/node/routes"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

type httpServer struct {
	server *http.Server
	mux    *http.ServeMux
}

// startHTTP starts the node's HTTP API server and returns immediately.
// The server runs in the background for the lifetime of the process.
func startHTTP(ctx context.Context, c *cli.Command, cfg *config.RocketPoolConfig) {
	port, ok := cfg.Smartnode.APIPort.Value.(uint16)
	if !ok || port == 0 {
		log.Println("Warning: APIPort not configured, HTTP API server will not start.")
		return
	}

	var host string
	if !cfg.IsNativeMode {
		// In Docker mode the server must bind to 0.0.0.0, so other containers can reach it.
		host = "0.0.0.0"
	} else {
		portMode, _ := cfg.Smartnode.OpenAPIPort.Value.(cfgtypes.RPCMode)
		if portMode == cfgtypes.RPC_OpenLocalhost {
			host = "127.0.0.1"
		} else {
			host = "0.0.0.0"
		}
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, c)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	go func() {
		log.Printf("Node HTTP API server listening on %s:%d\n", host, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Node HTTP API server error: %v\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
}
