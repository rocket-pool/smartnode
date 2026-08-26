package pdao

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	bindtypes "github.com/rocket-pool/smartnode/bindings/types"
	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// RegisterRoutes registers the pdao module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/pdao/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/pdao/proposals", proposalsHandler).RegisterTo(router)
	snroute.Read("/api/pdao/proposal-details", proposalDetailsHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-vote-proposal", canVoteProposalHandler).RegisterTo(router)
	snroute.Write("/api/pdao/vote-proposal", voteProposalHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-override-vote", canOverrideVoteHandler).RegisterTo(router)
	snroute.Write("/api/pdao/override-vote", overrideVoteHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-execute-proposal", canExecuteProposalHandler).RegisterTo(router)
	snroute.Write("/api/pdao/execute-proposal", executeProposalHandler).RegisterTo(router)
	snroute.Read("/api/pdao/get-settings", getSettingsHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-setting", canProposeSettingHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-setting", proposeSettingHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-setting-multi", canProposeSettingMultiHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-setting-multi", proposeSettingMultiHandler).RegisterTo(router)
	snroute.Read("/api/pdao/get-rewards-percentages", getRewardsPercentagesHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-rewards-percentages", canProposeRewardsPercentagesHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-rewards-percentages", proposeRewardsPercentagesHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-one-time-spend", canProposeOneTimeSpendHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-one-time-spend", proposeOneTimeSpendHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-recurring-spend", canProposeRecurringSpendHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-recurring-spend", proposeRecurringSpendHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-recurring-spend-update", canProposeRecurringSpendUpdateHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-recurring-spend-update", proposeRecurringSpendUpdateHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-invite-to-security-council", canProposeInviteToSecurityCouncilHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-invite-to-security-council", proposeInviteToSecurityCouncilHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-kick-from-security-council", canProposeKickFromSecurityCouncilHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-kick-from-security-council", proposeKickFromSecurityCouncilHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-kick-multi-from-security-council", canProposeKickMultiFromSecurityCouncilHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-kick-multi-from-security-council", proposeKickMultiFromSecurityCouncilHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-replace-member-of-security-council", canProposeReplaceMemberOfSecurityCouncilHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-replace-member-of-security-council", proposeReplaceMemberOfSecurityCouncilHandler).RegisterTo(router)
	snroute.Read("/api/pdao/get-claimable-bonds", getClaimableBondsHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-claim-bonds", canClaimBondsHandler).RegisterTo(router)
	snroute.Write("/api/pdao/claim-bonds", claimBondsHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-defeat-proposal", canDefeatProposalHandler).RegisterTo(router)
	snroute.Write("/api/pdao/defeat-proposal", defeatProposalHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-finalize-proposal", canFinalizeProposalHandler).RegisterTo(router)
	snroute.Write("/api/pdao/finalize-proposal", finalizeProposalHandler).RegisterTo(router)
	snroute.Read("/api/pdao/estimate-set-voting-delegate-gas", estimateSetVotingDelegateGasHandler).RegisterTo(router)
	snroute.Write("/api/pdao/set-voting-delegate", setVotingDelegateHandler).RegisterTo(router)
	snroute.Read("/api/pdao/get-current-voting-delegate", getCurrentVotingDelegateHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-set-signalling-address", canSetSignallingAddressHandler).RegisterTo(router)
	snroute.Write("/api/pdao/set-signalling-address", setSignallingAddressHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-clear-signalling-address", canClearSignallingAddressHandler).RegisterTo(router)
	snroute.Write("/api/pdao/clear-signalling-address", clearSignallingAddressHandler).RegisterTo(router)
	snroute.Read("/api/pdao/can-propose-allow-listed-controllers", canProposeAllowListedControllersHandler).RegisterTo(router)
	snroute.Write("/api/pdao/propose-allow-listed-controllers", proposeAllowListedControllersHandler).RegisterTo(router)
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

func parseBatchSettings(r *http.Request) ([]api.PDAOBatchSetting, string, error) {
	raw := paramVal(r, "settings")
	if raw == "" {
		return nil, "", fmt.Errorf("missing required parameter: settings")
	}
	var settings []api.PDAOBatchSetting
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, "", fmt.Errorf("invalid settings JSON: %w", err)
	}
	return settings, paramVal(r, "customMessage"), nil
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
