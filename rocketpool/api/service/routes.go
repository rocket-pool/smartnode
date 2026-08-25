package service

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the service module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/service/get-client-status", getClientStatusHandler(c)).RegisterTo(router)
	snroute.Write("/api/service/restart-vc", restartVcHandler(c)).RegisterTo(router)
	snroute.Write("/api/service/terminate-data-folder", terminateDataFolderHandler(c)).RegisterTo(router)
	snroute.Read("/api/service/get-gas-price-from-latest-block", getGasPriceFromLatestBlockHandler(c)).RegisterTo(router)
}
