package pdao

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	bindtypes "github.com/rocket-pool/smartnode/bindings/types"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// RegisterRoutes registers the pdao module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/pdao/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getProposals(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/proposal-details", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, voteDir, err := parseProposalVoteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canVoteOnProposal(c, id, voteDir)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, voteDir, err := parseProposalVoteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := voteOnProposal(c, id, voteDir)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-override-vote", func(w http.ResponseWriter, r *http.Request) {
		id, voteDir, err := parseProposalVoteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canOverrideVote(c, id, voteDir)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/override-vote", func(w http.ResponseWriter, r *http.Request) {
		id, voteDir, err := parseProposalVoteParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := overrideVote(c, id, voteDir)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExecuteProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := executeProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/get-settings", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getSettings(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-setting", func(w http.ResponseWriter, r *http.Request) {
		contract := paramVal(r, "contract")
		setting := paramVal(r, "setting")
		value := paramVal(r, "value")
		resp, err := canProposeSetting(c, contract, setting, value)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-setting", func(w http.ResponseWriter, r *http.Request) {
		contract := paramVal(r, "contract")
		setting := paramVal(r, "setting")
		value := paramVal(r, "value")
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSetting(c, contract, setting, value, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/get-rewards-percentages", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRewardsPercentages(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-rewards-percentages", func(w http.ResponseWriter, r *http.Request) {
		node, odaoAmt, pdaoAmt, err := parseRewardPercentages(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeRewardsPercentages(c, node, odaoAmt, pdaoAmt)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-rewards-percentages", func(w http.ResponseWriter, r *http.Request) {
		node, odaoAmt, pdaoAmt, err := parseRewardPercentages(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeRewardsPercentages(c, node, odaoAmt, pdaoAmt, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-one-time-spend", func(w http.ResponseWriter, r *http.Request) {
		invoiceID, recipient, amount, customMessage, err := parseOneTimeSpendParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeOneTimeSpend(c, invoiceID, recipient, amount, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-one-time-spend", func(w http.ResponseWriter, r *http.Request) {
		invoiceID, recipient, amount, customMessage, err := parseOneTimeSpendParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeOneTimeSpend(c, invoiceID, recipient, amount, blockNumber, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-recurring-spend", func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, false)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeRecurringSpend(c, contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-recurring-spend", func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, false)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeRecurringSpend(c, contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, blockNumber, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-recurring-spend-update", func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, _, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, true)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeRecurringSpendUpdate(c, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-recurring-spend-update", func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, _, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, true)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeRecurringSpendUpdate(c, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, blockNumber, customMessage)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-invite-to-security-council", func(w http.ResponseWriter, r *http.Request) {
		id := paramVal(r, "id")
		addr := common.HexToAddress(paramVal(r, "address"))
		resp, err := canProposeInviteToSecurityCouncil(c, id, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-invite-to-security-council", func(w http.ResponseWriter, r *http.Request) {
		id := paramVal(r, "id")
		addr := common.HexToAddress(paramVal(r, "address"))
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeInviteToSecurityCouncil(c, id, addr, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-kick-from-security-council", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		resp, err := canProposeKickFromSecurityCouncil(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-kick-from-security-council", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeKickFromSecurityCouncil(c, addr, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-kick-multi-from-security-council", func(w http.ResponseWriter, r *http.Request) {
		addresses, err := parseAddressList(r, "addresses")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeKickMultiFromSecurityCouncil(c, addresses)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-kick-multi-from-security-council", func(w http.ResponseWriter, r *http.Request) {
		addresses, err := parseAddressList(r, "addresses")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeKickMultiFromSecurityCouncil(c, addresses, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-replace-member-of-security-council", func(w http.ResponseWriter, r *http.Request) {
		existing := common.HexToAddress(paramVal(r, "existingAddress"))
		newID := paramVal(r, "newId")
		newAddr := common.HexToAddress(paramVal(r, "newAddress"))
		resp, err := canProposeReplaceMemberOfSecurityCouncil(c, existing, newID, newAddr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-replace-member-of-security-council", func(w http.ResponseWriter, r *http.Request) {
		existing := common.HexToAddress(paramVal(r, "existingAddress"))
		newID := paramVal(r, "newId")
		newAddr := common.HexToAddress(paramVal(r, "newAddress"))
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeReplaceMemberOfSecurityCouncil(c, existing, newID, newAddr, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/get-claimable-bonds", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getClaimableBonds(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-claim-bonds", func(w http.ResponseWriter, r *http.Request) {
		proposalID, indices, err := parseClaimBondsParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canClaimBonds(c, proposalID, indices)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/claim-bonds", func(w http.ResponseWriter, r *http.Request) {
		proposalID, indices, err := parseClaimBondsParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		isProposer := paramVal(r, "isProposer") == "true"
		resp, err := claimBonds(c, isProposer, proposalID, indices)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-defeat-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		index, err := parseUint64Param(r, "index")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDefeatProposal(c, id, index)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/defeat-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		index, err := parseUint64Param(r, "index")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := defeatProposal(c, id, index)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-finalize-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canFinalizeProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/finalize-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64Param(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := finalizeProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/estimate-set-voting-delegate-gas", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		resp, err := estimateSetVotingDelegateGas(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/set-voting-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		resp, err := setVotingDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/get-current-voting-delegate", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getCurrentVotingDelegate(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-set-signalling-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		sig := paramVal(r, "signature")
		resp, err := canSetSignallingAddress(c, addr, sig)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/set-signalling-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(paramVal(r, "address"))
		sig := paramVal(r, "signature")
		resp, err := setSignallingAddress(c, addr, sig)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-clear-signalling-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canClearSignallingAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/clear-signalling-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := clearSignallingAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/can-propose-allow-listed-controllers", func(w http.ResponseWriter, r *http.Request) {
		addressList := paramVal(r, "addressList")
		addresses, err := parseAddressList(r, "addressList")
		if err != nil {
			// Fall back to the raw comma-separated string if address parsing fails
			addresses = parseRawAddressList(addressList)
		}
		resp, err := canProposeAllowListedControllers(c, addresses)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/pdao/propose-allow-listed-controllers", func(w http.ResponseWriter, r *http.Request) {
		addressList := paramVal(r, "addressList")
		addresses, err := parseAddressList(r, "addressList")
		if err != nil {
			addresses = parseRawAddressList(addressList)
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeAllowListedControllers(c, addresses, blockNumber)
		apiutils.WriteResponse(w, resp, err)
	})
}

func paramVal(r *http.Request, name string) string {
	v := r.URL.Query().Get(name)
	if v == "" {
		v = r.FormValue(name)
	}
	return v
}

func parseUint64Param(r *http.Request, name string) (uint64, error) {
	raw := paramVal(r, name)
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return val, nil
}

func parseUint32Param(r *http.Request, name string) (uint32, error) {
	raw := paramVal(r, name)
	val, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return uint32(val), nil
}

func parseProposalVoteParams(r *http.Request) (uint64, bindtypes.VoteDirection, error) {
	id, err := parseUint64Param(r, "id")
	if err != nil {
		return 0, 0, err
	}
	dirStr := paramVal(r, "voteDirection")
	dir, err := cliutils.ValidateVoteDirection("voteDirection", dirStr)
	if err != nil {
		return 0, 0, err
	}
	return id, dir, nil
}

func parseRewardPercentages(r *http.Request) (*big.Int, *big.Int, *big.Int, error) {
	nodeStr := paramVal(r, "node")
	odaoStr := paramVal(r, "odao")
	pdaoStr := paramVal(r, "pdao")

	node, ok := new(big.Int).SetString(nodeStr, 10)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid node percentage: %s", nodeStr)
	}
	odaoAmt, ok := new(big.Int).SetString(odaoStr, 10)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid odao percentage: %s", odaoStr)
	}
	pdaoAmt, ok := new(big.Int).SetString(pdaoStr, 10)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid pdao percentage: %s", pdaoStr)
	}
	return node, odaoAmt, pdaoAmt, nil
}

func parseOneTimeSpendParams(r *http.Request) (string, common.Address, *big.Int, string, error) {
	invoiceID := paramVal(r, "invoiceId")
	recipient := common.HexToAddress(paramVal(r, "recipient"))
	amountStr := paramVal(r, "amount")
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return "", common.Address{}, nil, "", fmt.Errorf("invalid amount: %s", amountStr)
	}
	customMessage := paramVal(r, "customMessage")
	return invoiceID, recipient, amount, customMessage, nil
}

// parseRecurringSpendParams parses recurring spend parameters.
// If skipStartTime is true, the startTime is omitted (for update operations).
func parseRecurringSpendParams(r *http.Request, skipStartTime bool) (string, common.Address, *big.Int, time.Duration, time.Time, uint64, string, error) {
	contractName := paramVal(r, "contractName")
	recipient := common.HexToAddress(paramVal(r, "recipient"))

	amountStr := paramVal(r, "amountPerPeriod")
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return "", common.Address{}, nil, 0, time.Time{}, 0, "", fmt.Errorf("invalid amountPerPeriod: %s", amountStr)
	}

	periodLengthStr := paramVal(r, "periodLength")
	periodLength, err := time.ParseDuration(periodLengthStr)
	if err != nil {
		return "", common.Address{}, nil, 0, time.Time{}, 0, "", fmt.Errorf("invalid periodLength: %s", periodLengthStr)
	}

	var startTime time.Time
	if !skipStartTime {
		startTimeStr := paramVal(r, "startTime")
		startTimeUnix, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			return "", common.Address{}, nil, 0, time.Time{}, 0, "", fmt.Errorf("invalid startTime: %s", startTimeStr)
		}
		startTime = time.Unix(startTimeUnix, 0)
	}

	numberOfPeriodsStr := paramVal(r, "numberOfPeriods")
	numberOfPeriods, err := strconv.ParseUint(numberOfPeriodsStr, 10, 64)
	if err != nil {
		return "", common.Address{}, nil, 0, time.Time{}, 0, "", fmt.Errorf("invalid numberOfPeriods: %s", numberOfPeriodsStr)
	}

	customMessage := paramVal(r, "customMessage")
	return contractName, recipient, amount, periodLength, startTime, numberOfPeriods, customMessage, nil
}

func parseAddressList(r *http.Request, name string) ([]common.Address, error) {
	raw := paramVal(r, name)
	if raw == "" {
		return nil, fmt.Errorf("missing required parameter: %s", name)
	}
	return parseRawAddressList(raw), nil
}

func parseRawAddressList(raw string) []common.Address {
	parts := strings.Split(raw, ",")
	addresses := make([]common.Address, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			addresses = append(addresses, common.HexToAddress(p))
		}
	}
	return addresses
}

func parseClaimBondsParams(r *http.Request) (uint64, []uint64, error) {
	proposalID, err := parseUint64Param(r, "proposalId")
	if err != nil {
		return 0, nil, err
	}
	indicesStr := paramVal(r, "indices")
	parts := strings.Split(indicesStr, ",")
	indices := make([]uint64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid index: %s", p)
		}
		indices = append(indices, idx)
	}
	return proposalID, indices, nil
}
