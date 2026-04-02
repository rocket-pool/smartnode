package routes

import (
	"net/http"

	"github.com/urfave/cli/v3"

	apiroutes "github.com/rocket-pool/smartnode/rocketpool/api"
	auctionroutes "github.com/rocket-pool/smartnode/rocketpool/api/auction"
	debugroutes "github.com/rocket-pool/smartnode/rocketpool/api/debug"
	megapoolroutes "github.com/rocket-pool/smartnode/rocketpool/api/megapool"
	minipoolroutes "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
	networkroutes "github.com/rocket-pool/smartnode/rocketpool/api/network"
	noderoutes "github.com/rocket-pool/smartnode/rocketpool/api/node"
	odaoroutes "github.com/rocket-pool/smartnode/rocketpool/api/odao"
	pdaoroutes "github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	queueroutes "github.com/rocket-pool/smartnode/rocketpool/api/queue"
	securityroutes "github.com/rocket-pool/smartnode/rocketpool/api/security"
	serviceroutes "github.com/rocket-pool/smartnode/rocketpool/api/service"
	upgraderoutes "github.com/rocket-pool/smartnode/rocketpool/api/upgrade"
	walletroutes "github.com/rocket-pool/smartnode/rocketpool/api/wallet"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers all HTTP API routes onto mux.
// Each migration branch adds additional module registrations here.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	apiroutes.RegisterVersionRoute(mux)
	apiroutes.RegisterWaitRoute(mux, c)
	auctionroutes.RegisterRoutes(mux, c)
	debugroutes.RegisterRoutes(mux, c)
	megapoolroutes.RegisterRoutes(mux, c)
	minipoolroutes.RegisterRoutes(mux, c)
	networkroutes.RegisterRoutes(mux, c)
	noderoutes.RegisterRoutes(mux, c)
	odaoroutes.RegisterRoutes(mux, c)
	pdaoroutes.RegisterRoutes(mux, c)
	queueroutes.RegisterRoutes(mux, c)
	securityroutes.RegisterRoutes(mux, c)
	serviceroutes.RegisterRoutes(mux, c)
	upgraderoutes.RegisterRoutes(mux, c)
	walletroutes.RegisterRoutes(mux, c)

	// Catch-all: any path not matched by a specific route gets a JSON 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		apiutils.WriteErrorResponse(w, &apiutils.NotFoundError{Path: r.URL.Path})
	})

}
