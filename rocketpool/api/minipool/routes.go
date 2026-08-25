package minipool

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the minipool module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/minipool/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-refund", canRefundHandler).RegisterTo(router)
	snroute.Write("/api/minipool/refund", refundHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-stake", canStakeHandler).RegisterTo(router)
	snroute.Write("/api/minipool/stake", stakeHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-promote", canPromoteHandler).RegisterTo(router)
	snroute.Write("/api/minipool/promote", promoteHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-dissolve", canDissolveHandler).RegisterTo(router)
	snroute.Write("/api/minipool/dissolve", dissolveHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-exit", canExitHandler).RegisterTo(router)
	snroute.Write("/api/minipool/exit", exitHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-minipool-close-details-for-node", getMinipoolCloseDetailsForNodeHandler).RegisterTo(router)
	snroute.Write("/api/minipool/close", closeHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-delegate-upgrade", canDelegateUpgradeHandler).RegisterTo(router)
	snroute.Write("/api/minipool/delegate-upgrade", delegateUpgradeHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-set-use-latest-delegate", canSetUseLatestDelegateHandler).RegisterTo(router)
	snroute.Write("/api/minipool/set-use-latest-delegate", setUseLatestDelegateHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-use-latest-delegate", getUseLatestDelegateHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-delegate", getDelegateHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-effective-delegate", getEffectiveDelegateHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-previous-delegate", getPreviousDelegateHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-vanity-artifacts", getVanityArtifactsHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-distribute-balance-details", getDistributeBalanceDetailsHandler).RegisterTo(router)
	snroute.Write("/api/minipool/distribute-balance", distributeBalanceHandler).RegisterTo(router)
	snroute.Write("/api/minipool/import-key", importKeyHandler).RegisterTo(router)
	snroute.Read("/api/minipool/can-change-withdrawal-creds", canChangeWithdrawalCredsHandler).RegisterTo(router)
	snroute.Write("/api/minipool/change-withdrawal-creds", changeWithdrawalCredsHandler).RegisterTo(router)
	snroute.Read("/api/minipool/get-rescue-dissolved-details-for-node", getRescueDissolvedDetailsForNodeHandler).RegisterTo(router)
	snroute.Write("/api/minipool/rescue-dissolved", rescueDissolvedHandler).RegisterTo(router)

}

func parseAddress(r *http.Request, name string) (common.Address, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	if raw == "" {
		return common.Address{}, fmt.Errorf("missing required parameter: %s", name)
	}
	return common.HexToAddress(raw), nil
}
