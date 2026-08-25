package network

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the network module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/network/node-fee", nodeFeeHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/rpl-price", rplPriceHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/stats", statsHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/timezone-map", timezoneMapHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/can-generate-rewards-tree", canGenerateRewardsTreeHandler(c)).RegisterTo(router)
	snroute.Write("/api/network/generate-rewards-tree", generateRewardsTreeHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/dao-proposals", daoProposalsHandler(c)).RegisterTo(router)
	snroute.Write("/api/network/download-rewards-file", downloadRewardsFileHandler(c)).RegisterTo(router)
	snroute.Read("/api/network/latest-delegate", latestDelegateHandler(c)).RegisterTo(router)
}

func parseUint64Param(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
