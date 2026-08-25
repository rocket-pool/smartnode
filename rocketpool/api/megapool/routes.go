package megapool

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the megapool module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/megapool/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/validator-map-and-balances", validatorMapAndBalancesHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-claim-refund", canClaimRefundHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/claim-refund", claimRefundHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-repay-debt", canRepayDebtHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/repay-debt", repayDebtHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-reduce-bond", canReduceBondHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/reduce-bond", reduceBondHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-stake", canStakeHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/stake", stakeHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-dissolve-validator", canDissolveValidatorHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/dissolve-validator", dissolveValidatorHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-dissolve-with-proof", canDissolveWithProofHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/dissolve-with-proof", dissolveWithProofHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-exit-validator", canExitValidatorHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/exit-validator", exitValidatorHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-notify-validator-exit", canNotifyValidatorExitHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/notify-validator-exit", notifyValidatorExitHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-notify-final-balance", canNotifyFinalBalanceHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/notify-final-balance", notifyFinalBalanceHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-exit-queue", canExitQueueHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/exit-queue", exitQueueHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-distribute", canDistributeHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/distribute", distributeHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/get-new-validator-bond-requirement", getNewValidatorBondRequirementHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/pending-rewards", pendingRewardsHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/calculate-rewards", calculateRewardsHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/get-use-latest-delegate", getUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-delegate-upgrade", canDelegateUpgradeHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/delegate-upgrade", delegateUpgradeHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/can-set-use-latest-delegate", canSetUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Write("/api/megapool/set-use-latest-delegate", setUseLatestDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/get-delegate", getDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/get-effective-delegate", getEffectiveDelegateHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/latest-block-withdrawals", latestBlockWithdrawalsHandler(c)).RegisterTo(router)
	snroute.Read("/api/megapool/beacon-withdrawal-queue-estimate", beaconWithdrawalQueueEstimateHandler(c)).RegisterTo(router)
}

func parseUint64(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}

func parseUint32(r *http.Request, name string) (uint32, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	return uint32(v), err
}

func parseBool(r *http.Request, name string) (bool, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	if raw == "" {
		return false, &response.BadRequestError{Err: fmt.Errorf("missing required parameter '%s'", name)}
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &response.BadRequestError{Err: fmt.Errorf("invalid %s: %s", name, raw)}
	}
	return v, nil
}

func parseBigInt(r *http.Request, name string) (*big.Int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	if raw == "" {
		return nil, &response.BadRequestError{Err: fmt.Errorf("missing required parameter '%s'", name)}
	}
	v, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, &response.BadRequestError{Err: fmt.Errorf("invalid %s: %s", name, raw)}
	}
	return v, nil
}
