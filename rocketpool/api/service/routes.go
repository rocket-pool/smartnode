package service

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the service module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/service/get-client-status", getClientStatusHandler).RegisterTo(router)
	snroute.Write("/api/service/restart-vc", restartVcHandler).RegisterTo(router)
	snroute.Write("/api/service/terminate-data-folder", terminateDataFolderHandler).RegisterTo(router)
	snroute.Read("/api/service/get-gas-price-from-latest-block", getGasPriceFromLatestBlockHandler).RegisterTo(router)
}
