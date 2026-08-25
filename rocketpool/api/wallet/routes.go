package wallet

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the wallet module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/wallet/status", statusHandler).RegisterTo(router)
	snroute.Write("/api/wallet/set-password", setPasswordHandler).RegisterTo(router)
	snroute.Write("/api/wallet/init", initHandler).RegisterTo(router)

	// Reports on the recovery currently holding the lock below, so the CLI can
	// explain a rejected command rather than just failing
	snroute.Read("/api/wallet/recovery-status", recoveryStatusHandler).RegisterTo(router)
	snroute.Write("/api/wallet/recover", recoverHandler).RegisterTo(router)
	snroute.Write("/api/wallet/search-and-recover", searchAndRecoverHandler).RegisterTo(router)
	snroute.Read("/api/wallet/test-recover", testRecoverHandler).RegisterTo(router)
	snroute.Read("/api/wallet/test-search-and-recover", testSearchAndRecoverHandler).RegisterTo(router)
	snroute.Write("/api/wallet/rebuild", rebuildHandler).RegisterTo(router)
	snroute.Write("/api/wallet/export", exportHandler).RegisterTo(router)
	snroute.Write("/api/wallet/masquerade", masqueradeHandler).RegisterTo(router)
	snroute.Write("/api/wallet/end-masquerade", endMasqueradeHandler).RegisterTo(router)
	snroute.Read("/api/wallet/estimate-gas-set-ens-name", estimateGasSetEnsNameHandler).RegisterTo(router)
	snroute.Write("/api/wallet/set-ens-name", setEnsNameHandler).RegisterTo(router)
}
