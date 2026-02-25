package node

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/urfave/cli"

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
func startHTTP(ctx context.Context, c *cli.Context, cfg *config.RocketPoolConfig) {
	port, ok := cfg.Smartnode.APIPort.Value.(uint16)
	if !ok || port == 0 {
		log.Println("Warning: APIPort not configured, HTTP API server will not start.")
		return
	}

	// In Docker mode the server must bind to 0.0.0.0 so the port mapping in
	// node.tmpl makes it accessible from the host.  In native mode we respect
	// the OpenAPIPort setting: Closed → 127.0.0.1, OpenLocalhost → 0.0.0.0.
	var host string
	if !cfg.IsNativeMode {
		host = "0.0.0.0"
	} else {
		portMode, _ := cfg.Smartnode.OpenAPIPort.Value.(cfgtypes.RPCMode)
		if portMode == cfgtypes.RPC_OpenLocalhost {
			host = "0.0.0.0"
		} else {
			host = "127.0.0.1"
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
