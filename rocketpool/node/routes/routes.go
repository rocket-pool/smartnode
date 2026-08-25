package routes

import (
	"net/http"

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
func RegisterRoutes(router *snroute.Router) {
	snroute.Open("/healthz", healthzHandler).RegisterTo(router)

	apiroutes.RegisterVersionRoute(router)
	apiroutes.RegisterWaitRoute(router)
	auctionroutes.RegisterRoutes(router)
	debugroutes.RegisterRoutes(router)
	megapoolroutes.RegisterRoutes(router)
	minipoolroutes.RegisterRoutes(router)
	networkroutes.RegisterRoutes(router)
	noderoutes.RegisterRoutes(router)
	odaoroutes.RegisterRoutes(router)
	pdaoroutes.RegisterRoutes(router)
	queueroutes.RegisterRoutes(router)
	securityroutes.RegisterRoutes(router)
	serviceroutes.RegisterRoutes(router)
	upgraderoutes.RegisterRoutes(router)
	walletroutes.RegisterRoutes(router)

	// Catch-all: any path not matched by a specific route gets a JSON 404.
	snroute.Read("/", notFoundHandler).RegisterTo(router)

}

func healthzHandler(ctx snroute.Context) {
	ctx.Writer.WriteHeader(http.StatusOK)
}

func notFoundHandler(ctx snroute.Context) {
	response.WriteErrorResponse(ctx.Writer, &response.NotFoundError{Path: ctx.Request.URL.Path})
}
