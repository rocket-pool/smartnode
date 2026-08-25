package upgrade

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the upgrade module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/upgrade/get-upgrade-proposals", getUpgradeProposalsHandler).RegisterTo(router)
	snroute.Read("/api/upgrade/can-execute-upgrade", canExecuteUpgradeHandler).RegisterTo(router)
	snroute.Write("/api/upgrade/execute-upgrade", executeUpgradeHandler).RegisterTo(router)
}
