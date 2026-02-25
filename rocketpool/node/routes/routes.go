package routes

import (
	"net/http"

	"github.com/urfave/cli"

	auctionroutes "github.com/rocket-pool/smartnode/rocketpool/api/auction"
	networkroutes "github.com/rocket-pool/smartnode/rocketpool/api/network"
	queueroutes "github.com/rocket-pool/smartnode/rocketpool/api/queue"
	serviceroutes "github.com/rocket-pool/smartnode/rocketpool/api/service"
	walletroutes "github.com/rocket-pool/smartnode/rocketpool/api/wallet"
)

// RegisterRoutes registers all HTTP API routes onto mux.
// Each migration branch adds additional module registrations here.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	auctionroutes.RegisterRoutes(mux, c)
	networkroutes.RegisterRoutes(mux, c)
	queueroutes.RegisterRoutes(mux, c)
	serviceroutes.RegisterRoutes(mux, c)
	walletroutes.RegisterRoutes(mux, c)
}
