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
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	securityroutes "github.com/rocket-pool/smartnode/rocketpool/api/security"
	serviceroutes "github.com/rocket-pool/smartnode/rocketpool/api/service"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	upgraderoutes "github.com/rocket-pool/smartnode/rocketpool/api/upgrade"
	walletroutes "github.com/rocket-pool/smartnode/rocketpool/api/wallet"
)

// RegisterRoutes registers all HTTP API routes onto router.
// Each migration branch adds additional module registrations here.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Open("/healthz", healthzHandler).RegisterTo(router)

	apiroutes.RegisterVersionRoute(router)
	apiroutes.RegisterWaitRoute(router, c)
	auctionroutes.RegisterRoutes(router, c)
	debugroutes.RegisterRoutes(router, c)
	megapoolroutes.RegisterRoutes(router, c)
	minipoolroutes.RegisterRoutes(router, c)
	networkroutes.RegisterRoutes(router, c)
	noderoutes.RegisterRoutes(router, c)
	odaoroutes.RegisterRoutes(router, c)
	pdaoroutes.RegisterRoutes(router, c)
	queueroutes.RegisterRoutes(router, c)
	securityroutes.RegisterRoutes(router, c)
	serviceroutes.RegisterRoutes(router, c)
	upgraderoutes.RegisterRoutes(router, c)
	walletroutes.RegisterRoutes(router, c)

	// Catch-all: any path not matched by a specific route gets a JSON 404.
	snroute.Read("/", notFoundHandler).RegisterTo(router)

}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteErrorResponse(w, &response.NotFoundError{Path: r.URL.Path})
}
