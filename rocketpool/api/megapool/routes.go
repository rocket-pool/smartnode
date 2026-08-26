package megapool

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the megapool module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/megapool/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/megapool/validator-map-and-balances", validatorMapAndBalancesHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-claim-refund", canClaimRefundHandler).RegisterTo(router)
	snroute.Write("/api/megapool/claim-refund", claimRefundHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-repay-debt", canRepayDebtHandler).RegisterTo(router)
	snroute.Write("/api/megapool/repay-debt", repayDebtHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-reduce-bond", canReduceBondHandler).RegisterTo(router)
	snroute.Write("/api/megapool/reduce-bond", reduceBondHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-stake", canStakeHandler).RegisterTo(router)
	snroute.Write("/api/megapool/stake", stakeHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-dissolve-validator", canDissolveValidatorHandler).RegisterTo(router)
	snroute.Write("/api/megapool/dissolve-validator", dissolveValidatorHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-dissolve-with-proof", canDissolveWithProofHandler).RegisterTo(router)
	snroute.Write("/api/megapool/dissolve-with-proof", dissolveWithProofHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-exit-validator", canExitValidatorHandler).RegisterTo(router)
	snroute.Write("/api/megapool/exit-validator", exitValidatorHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-notify-validator-exit", canNotifyValidatorExitHandler).RegisterTo(router)
	snroute.Write("/api/megapool/notify-validator-exit", notifyValidatorExitHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-notify-final-balance", canNotifyFinalBalanceHandler).RegisterTo(router)
	snroute.Write("/api/megapool/notify-final-balance", notifyFinalBalanceHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-exit-queue", canExitQueueHandler).RegisterTo(router)
	snroute.Write("/api/megapool/exit-queue", exitQueueHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-distribute", canDistributeHandler).RegisterTo(router)
	snroute.Write("/api/megapool/distribute", distributeHandler).RegisterTo(router)
	snroute.Read("/api/megapool/get-new-validator-bond-requirement", getNewValidatorBondRequirementHandler).RegisterTo(router)
	snroute.Read("/api/megapool/pending-rewards", pendingRewardsHandler).RegisterTo(router)
	snroute.Read("/api/megapool/calculate-rewards", calculateRewardsHandler).RegisterTo(router)
	snroute.Read("/api/megapool/get-use-latest-delegate", getUseLatestDelegateHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-delegate-upgrade", canDelegateUpgradeHandler).RegisterTo(router)
	snroute.Write("/api/megapool/delegate-upgrade", delegateUpgradeHandler).RegisterTo(router)
	snroute.Read("/api/megapool/can-set-use-latest-delegate", canSetUseLatestDelegateHandler).RegisterTo(router)
	snroute.Write("/api/megapool/set-use-latest-delegate", setUseLatestDelegateHandler).RegisterTo(router)
	snroute.Read("/api/megapool/get-delegate", getDelegateHandler).RegisterTo(router)
	snroute.Read("/api/megapool/get-effective-delegate", getEffectiveDelegateHandler).RegisterTo(router)
	snroute.Read("/api/megapool/latest-block-withdrawals", latestBlockWithdrawalsHandler).RegisterTo(router)
	snroute.Read("/api/megapool/beacon-withdrawal-queue-estimate", beaconWithdrawalQueueEstimateHandler).RegisterTo(router)
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
