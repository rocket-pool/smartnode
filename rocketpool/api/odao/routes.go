package odao

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the odao module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/odao/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/members", membersHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/proposals", proposalsHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/proposal-details", proposalDetailsHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-invite", canProposeInviteHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-invite", proposeInviteHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-leave", canProposeLeaveHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-leave", proposeLeaveHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-kick", canProposeKickHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-kick", proposeKickHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-cancel-proposal", canCancelProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/cancel-proposal", cancelProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-vote-proposal", canVoteProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/vote-proposal", voteProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-execute-proposal", canExecuteProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/execute-proposal", executeProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-join", canJoinHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/join-approve-rpl", joinApproveRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/join", joinHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-leave", canLeaveHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/leave", leaveHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/get-member-settings", getMemberSettingsHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/get-proposal-settings", getProposalSettingsHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/get-minipool-settings", getMinipoolSettingsHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-penalise-megapool", canPenaliseMegapoolHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/penalise-megapool", penaliseMegapoolHandler(c)).RegisterTo(router)

	// propose-settings endpoints
	snroute.Read("/api/odao/can-propose-members-quorum", canProposeMembersQuorumHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-members-quorum", proposeMembersQuorumHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-members-rplbond", canProposeMembersRplbondHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-members-rplbond", proposeMembersRplbondHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-cooldown", canProposeProposalCooldownHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-cooldown", proposeProposalCooldownHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-vote-timespan", canProposeProposalVoteTimespanHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-vote-timespan", proposeProposalVoteTimespanHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-vote-delay-timespan", canProposeProposalVoteDelayTimespanHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-vote-delay-timespan", proposeProposalVoteDelayTimespanHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-execute-timespan", canProposeProposalExecuteTimespanHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-execute-timespan", proposeProposalExecuteTimespanHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-proposal-action-timespan", canProposeProposalActionTimespanHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-proposal-action-timespan", proposeProposalActionTimespanHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-scrub-period", canProposeScrubPeriodHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-scrub-period", proposeScrubPeriodHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-promotion-scrub-period", canProposePromotionScrubPeriodHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-promotion-scrub-period", proposePromotionScrubPeriodHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-scrub-penalty-enabled", canProposeScrubPenaltyEnabledHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-scrub-penalty-enabled", proposeScrubPenaltyEnabledHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-bond-reduction-window-start", canProposeBondReductionWindowStartHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-bond-reduction-window-start", proposeBondReductionWindowStartHandler(c)).RegisterTo(router)
	snroute.Read("/api/odao/can-propose-bond-reduction-window-length", canProposeBondReductionWindowLengthHandler(c)).RegisterTo(router)
	snroute.Write("/api/odao/propose-bond-reduction-window-length", proposeBondReductionWindowLengthHandler(c)).RegisterTo(router)
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
