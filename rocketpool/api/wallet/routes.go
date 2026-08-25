package wallet

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the wallet module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/wallet/status", statusHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/set-password", setPasswordHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/init", initHandler(c)).RegisterTo(router)

	// Reports on the recovery currently holding the lock below, so the CLI can
	// explain a rejected command rather than just failing
	snroute.Read("/api/wallet/recovery-status", recoveryStatusHandler).RegisterTo(router)
	snroute.Write("/api/wallet/recover", recoverHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/search-and-recover", searchAndRecoverHandler(c)).RegisterTo(router)
	snroute.Read("/api/wallet/test-recover", testRecoverHandler(c)).RegisterTo(router)
	snroute.Read("/api/wallet/test-search-and-recover", testSearchAndRecoverHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/rebuild", rebuildHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/export", exportHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/masquerade", masqueradeHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/end-masquerade", endMasqueradeHandler(c)).RegisterTo(router)
	snroute.Read("/api/wallet/estimate-gas-set-ens-name", estimateGasSetEnsNameHandler(c)).RegisterTo(router)
	snroute.Write("/api/wallet/set-ens-name", setEnsNameHandler(c)).RegisterTo(router)
}
