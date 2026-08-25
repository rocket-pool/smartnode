package network

import (
	"net/http"
	"strconv"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the network module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/network/node-fee", nodeFeeHandler).RegisterTo(router)
	snroute.Read("/api/network/rpl-price", rplPriceHandler).RegisterTo(router)
	snroute.Read("/api/network/stats", statsHandler).RegisterTo(router)
	snroute.Read("/api/network/timezone-map", timezoneMapHandler).RegisterTo(router)
	snroute.Read("/api/network/can-generate-rewards-tree", canGenerateRewardsTreeHandler).RegisterTo(router)
	snroute.Write("/api/network/generate-rewards-tree", generateRewardsTreeHandler).RegisterTo(router)
	snroute.Read("/api/network/dao-proposals", daoProposalsHandler).RegisterTo(router)
	snroute.Write("/api/network/download-rewards-file", downloadRewardsFileHandler).RegisterTo(router)
	snroute.Read("/api/network/latest-delegate", latestDelegateHandler).RegisterTo(router)
}

func parseUint64Param(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
