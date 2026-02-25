package odao

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the odao module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/odao/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/members", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMembers(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getProposals(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/proposal-details", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-invite", func(w http.ResponseWriter, r *http.Request) {
		addr, memberId, memberUrl, err := parseInviteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeInvite(c, addr, memberId, memberUrl)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-invite", func(w http.ResponseWriter, r *http.Request) {
		addr, memberId, memberUrl, err := parseInviteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeInvite(c, addr, memberId, memberUrl)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-leave", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canProposeLeave(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-leave", func(w http.ResponseWriter, r *http.Request) {
		resp, err := proposeLeave(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-kick", func(w http.ResponseWriter, r *http.Request) {
		addr, fine, err := parseKickParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeKick(c, addr, fine)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-kick", func(w http.ResponseWriter, r *http.Request) {
		addr, fine, err := parseKickParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeKick(c, addr, fine)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-cancel-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canCancelProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/cancel-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := cancelProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canVoteOnProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		supportStr := r.FormValue("support")
		support := supportStr == "true"
		resp, err := voteOnProposal(c, id, support)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExecuteProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := executeProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-join", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canJoin(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/join-approve-rpl", func(w http.ResponseWriter, r *http.Request) {
		resp, err := approveRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/join", func(w http.ResponseWriter, r *http.Request) {
		hashStr := r.FormValue("approvalTxHash")
		if hashStr == "" {
			apiutils.WriteErrorResponse(w, fmt.Errorf("missing required parameter: approvalTxHash"))
			return
		}
		hash := common.HexToHash(hashStr)
		resp, err := waitForApprovalAndJoin(c, hash)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-leave", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canLeave(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/leave", func(w http.ResponseWriter, r *http.Request) {
		bondRefundStr := r.FormValue("bondRefundAddress")
		if bondRefundStr == "" {
			apiutils.WriteErrorResponse(w, fmt.Errorf("missing required parameter: bondRefundAddress"))
			return
		}
		resp, err := leave(c, common.HexToAddress(bondRefundStr))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/get-member-settings", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMemberSettings(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/get-proposal-settings", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getProposalSettings(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/get-minipool-settings", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMinipoolSettings(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-penalise-megapool", func(w http.ResponseWriter, r *http.Request) {
		megapool, block, amount, err := parsePenaliseParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canPenaliseMegapool(c, megapool, block, amount)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/penalise-megapool", func(w http.ResponseWriter, r *http.Request) {
		megapool, block, amount, err := parsePenaliseParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := penaliseMegapool(c, megapool, block, amount)
		apiutils.WriteResponse(w, resp, err)
	})

	// propose-settings endpoints
	mux.HandleFunc("/api/odao/can-propose-members-quorum", func(w http.ResponseWriter, r *http.Request) {
		quorum, err := parseFloat64(r, "quorum")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingMembersQuorum(c, quorum)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-members-quorum", func(w http.ResponseWriter, r *http.Request) {
		quorum, err := parseFloat64(r, "quorum")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingMembersQuorum(c, quorum)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-members-rplbond", func(w http.ResponseWriter, r *http.Request) {
		bond, err := parseBigInt(r, "bondAmountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingMembersRplBond(c, bond)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-members-rplbond", func(w http.ResponseWriter, r *http.Request) {
		bond, err := parseBigInt(r, "bondAmountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingMembersRplBond(c, bond)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-members-minipool-unbonded-max", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint64(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingMinipoolUnbondedMax(c, max)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-members-minipool-unbonded-max", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint64(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingMinipoolUnbondedMax(c, max)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-proposal-cooldown", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalCooldown(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-proposal-cooldown", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalCooldown(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-proposal-vote-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalVoteTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-proposal-vote-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalVoteTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-proposal-vote-delay-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalVoteDelayTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-proposal-vote-delay-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalVoteDelayTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-proposal-execute-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalExecuteTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-proposal-execute-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalExecuteTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-proposal-action-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalActionTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-proposal-action-timespan", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalActionTimespan(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-scrub-period", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingScrubPeriod(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-scrub-period", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingScrubPeriod(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-promotion-scrub-period", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingPromotionScrubPeriod(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-promotion-scrub-period", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingPromotionScrubPeriod(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-scrub-penalty-enabled", func(w http.ResponseWriter, r *http.Request) {
		enabledStr := r.URL.Query().Get("enabled")
		resp, err := canProposeSettingScrubPenaltyEnabled(c, enabledStr == "true")
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-scrub-penalty-enabled", func(w http.ResponseWriter, r *http.Request) {
		enabledStr := r.FormValue("enabled")
		resp, err := proposeSettingScrubPenaltyEnabled(c, enabledStr == "true")
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-bond-reduction-window-start", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingBondReductionWindowStart(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-bond-reduction-window-start", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingBondReductionWindowStart(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/can-propose-bond-reduction-window-length", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingBondReductionWindowLength(c, val)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/odao/propose-bond-reduction-window-length", func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingBondReductionWindowLength(c, val)
		apiutils.WriteResponse(w, resp, err)
	})
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
