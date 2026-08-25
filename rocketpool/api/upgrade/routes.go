package upgrade

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the upgrade module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/upgrade/get-upgrade-proposals", getUpgradeProposalsHandler(c)).RegisterTo(router)
	snroute.Read("/api/upgrade/can-execute-upgrade", canExecuteUpgradeHandler(c)).RegisterTo(router)
	snroute.Write("/api/upgrade/execute-upgrade", executeUpgradeHandler(c)).RegisterTo(router)
}
