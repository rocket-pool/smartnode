package minipool

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the minipool module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/minipool/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-refund", canRefundHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/refund", refundHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-stake", canStakeHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/stake", stakeHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-promote", canPromoteHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/promote", promoteHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-dissolve", canDissolveHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/dissolve", dissolveHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-exit", canExitHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/exit", exitHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-minipool-close-details-for-node", getMinipoolCloseDetailsForNodeHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/close", closeHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-delegate-upgrade", canDelegateUpgradeHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/delegate-upgrade", delegateUpgradeHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-set-use-latest-delegate", canSetUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/set-use-latest-delegate", setUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-use-latest-delegate", getUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-delegate", getDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-effective-delegate", getEffectiveDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-previous-delegate", getPreviousDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-vanity-artifacts", getVanityArtifactsHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-distribute-balance-details", getDistributeBalanceDetailsHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/distribute-balance", distributeBalanceHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/import-key", importKeyHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/can-change-withdrawal-creds", canChangeWithdrawalCredsHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/change-withdrawal-creds", changeWithdrawalCredsHandler(c)).RegisterTo(router)
	snroute.Read("/api/minipool/get-rescue-dissolved-details-for-node", getRescueDissolvedDetailsForNodeHandler(c)).RegisterTo(router)
	snroute.Write("/api/minipool/rescue-dissolved", rescueDissolvedHandler(c)).RegisterTo(router)

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
