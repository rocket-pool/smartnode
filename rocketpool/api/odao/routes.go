package odao

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the odao module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/odao/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/odao/members", membersHandler).RegisterTo(router)
	snroute.Read("/api/odao/proposals", proposalsHandler).RegisterTo(router)
	snroute.Read("/api/odao/proposal-details", proposalDetailsHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-invite", canProposeInviteHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-invite", proposeInviteHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-leave", canProposeLeaveHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-leave", proposeLeaveHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-kick", canProposeKickHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-kick", proposeKickHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-cancel-proposal", canCancelProposalHandler).RegisterTo(router)
	snroute.Write("/api/odao/cancel-proposal", cancelProposalHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-vote-proposal", canVoteProposalHandler).RegisterTo(router)
	snroute.Write("/api/odao/vote-proposal", voteProposalHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-execute-proposal", canExecuteProposalHandler).RegisterTo(router)
	snroute.Write("/api/odao/execute-proposal", executeProposalHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-join", canJoinHandler).RegisterTo(router)
	snroute.Write("/api/odao/join-approve-rpl", joinApproveRplHandler).RegisterTo(router)
	snroute.Write("/api/odao/join", joinHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-leave", canLeaveHandler).RegisterTo(router)
	snroute.Write("/api/odao/leave", leaveHandler).RegisterTo(router)
	snroute.Read("/api/odao/get-member-settings", getMemberSettingsHandler).RegisterTo(router)
	snroute.Read("/api/odao/get-proposal-settings", getProposalSettingsHandler).RegisterTo(router)
	snroute.Read("/api/odao/get-minipool-settings", getMinipoolSettingsHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-penalise-megapool", canPenaliseMegapoolHandler).RegisterTo(router)
	snroute.Write("/api/odao/penalise-megapool", penaliseMegapoolHandler).RegisterTo(router)

	// propose-settings endpoints
	snroute.Read("/api/odao/can-propose-members-quorum", canProposeMembersQuorumHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-members-quorum", proposeMembersQuorumHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-members-rplbond", canProposeMembersRplbondHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-members-rplbond", proposeMembersRplbondHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-cooldown", canProposeProposalCooldownHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-cooldown", proposeProposalCooldownHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-vote-timespan", canProposeProposalVoteTimespanHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-vote-timespan", proposeProposalVoteTimespanHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-vote-delay-timespan", canProposeProposalVoteDelayTimespanHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-vote-delay-timespan", proposeProposalVoteDelayTimespanHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-execute-timespan", canProposeProposalExecuteTimespanHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-execute-timespan", proposeProposalExecuteTimespanHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-action-timespan", canProposeProposalActionTimespanHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-action-timespan", proposeProposalActionTimespanHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-scrub-period", canProposeScrubPeriodHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-scrub-period", proposeScrubPeriodHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-promotion-scrub-period", canProposePromotionScrubPeriodHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-promotion-scrub-period", proposePromotionScrubPeriodHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-scrub-penalty-enabled", canProposeScrubPenaltyEnabledHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-scrub-penalty-enabled", proposeScrubPenaltyEnabledHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-bond-reduction-window-start", canProposeBondReductionWindowStartHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-bond-reduction-window-start", proposeBondReductionWindowStartHandler).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-bond-reduction-window-length", canProposeBondReductionWindowLengthHandler).RegisterTo(router)
	snroute.Write("/api/odao/propose-bond-reduction-window-length", proposeBondReductionWindowLengthHandler).RegisterTo(router)
}

func parseUint64(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return val, nil
}

func parseFloat64(r *http.Request, name string) (float64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return val, nil
}

func parseBigInt(r *http.Request, name string) (*big.Int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	val, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return val, nil
}

func parseInviteParams(r *http.Request) (common.Address, string, string, error) {
	addrStr := r.URL.Query().Get("address")
	if addrStr == "" {
		addrStr = r.FormValue("address")
	}
	if addrStr == "" {
		return common.Address{}, "", "", fmt.Errorf("missing required parameter: address")
	}
	memberId := r.URL.Query().Get("memberId")
	if memberId == "" {
		memberId = r.FormValue("memberId")
	}
	memberUrl := r.URL.Query().Get("memberUrl")
	if memberUrl == "" {
		memberUrl = r.FormValue("memberUrl")
	}
	return common.HexToAddress(addrStr), memberId, memberUrl, nil
}

func parseKickParams(r *http.Request) (common.Address, *big.Int, error) {
	addrStr := r.URL.Query().Get("address")
	if addrStr == "" {
		addrStr = r.FormValue("address")
	}
	if addrStr == "" {
		return common.Address{}, nil, fmt.Errorf("missing required parameter: address")
	}
	fineStr := r.URL.Query().Get("fineAmountWei")
	if fineStr == "" {
		fineStr = r.FormValue("fineAmountWei")
	}
	fine, ok := new(big.Int).SetString(fineStr, 10)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("invalid fineAmountWei: %s", fineStr)
	}
	return common.HexToAddress(addrStr), fine, nil
}

func parsePenaliseParams(r *http.Request) (common.Address, *big.Int, *big.Int, error) {
	addrStr := r.URL.Query().Get("megapoolAddress")
	if addrStr == "" {
		addrStr = r.FormValue("megapoolAddress")
	}
	blockStr := r.URL.Query().Get("block")
	if blockStr == "" {
		blockStr = r.FormValue("block")
	}
	amountStr := r.URL.Query().Get("amountWei")
	if amountStr == "" {
		amountStr = r.FormValue("amountWei")
	}
	block, ok := new(big.Int).SetString(blockStr, 10)
	if !ok {
		return common.Address{}, nil, nil, fmt.Errorf("invalid block: %s", blockStr)
	}
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return common.Address{}, nil, nil, fmt.Errorf("invalid amountWei: %s", amountStr)
	}
	return common.HexToAddress(addrStr), block, amount, nil
}
