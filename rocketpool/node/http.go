package node

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"google.golang.org/genproto/googleapis/maps/routes/v1"
)

type httpServer struct {
	server *http.Server
}

func startHTTP(ctx context.Context, cfg *config.RocketPoolConfig) {
	host := "127.0.0.1"
	if cfg.Smartnode.OpenAPIPort.Value == cfgtypes.RPC_OpenLocalhost {
		host = "0.0.0.0"
	}

	port, ok := cfg.Smartnode.APIPort.Value.(uint16)
	if !ok {
		log.Fatalf("Error getting API port: %v", err)
	}

	httpServer := &httpServer{}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: httpServer,
	}

	httpServer.server = server
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func (s *httpServer) addRoute(r *routes.Route) {
}
