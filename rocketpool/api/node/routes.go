package node

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the node module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/node/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/node/alerts", alertsHandler).RegisterTo(router)
	snroute.Read("/api/node/sync", syncHandler).RegisterTo(router)
	snroute.Read("/api/node/get-eth-balance", getEthBalanceHandler).RegisterTo(router)
	snroute.Read("/api/node/check-collateral", checkCollateralHandler).RegisterTo(router)
	snroute.Read("/api/node/rewards", rewardsHandler).RegisterTo(router)
	snroute.Read("/api/node/deposit-contract-info", depositContractInfoHandler).RegisterTo(router)

	// --- Register ---

	snroute.Read("/api/node/can-register", canRegisterHandler).RegisterTo(router)
	snroute.Write("/api/node/register", registerHandler).RegisterTo(router)

	// --- Timezone ---

	snroute.Read("/api/node/can-set-timezone", canSetTimezoneHandler).RegisterTo(router)
	snroute.Write("/api/node/set-timezone", setTimezoneHandler).RegisterTo(router)

	// --- Primary withdrawal address ---

	snroute.Read("/api/node/can-set-primary-withdrawal-address", canSetPrimaryWithdrawalAddressHandler).RegisterTo(router)
	snroute.Write("/api/node/set-primary-withdrawal-address", setPrimaryWithdrawalAddressHandler).RegisterTo(router)
	snroute.Read("/api/node/can-confirm-primary-withdrawal-address", canConfirmPrimaryWithdrawalAddressHandler).RegisterTo(router)
	snroute.Write("/api/node/confirm-primary-withdrawal-address", confirmPrimaryWithdrawalAddressHandler).RegisterTo(router)

	// --- RPL withdrawal address ---

	snroute.Read("/api/node/can-set-rpl-withdrawal-address", canSetRplWithdrawalAddressHandler).RegisterTo(router)
	snroute.Write("/api/node/set-rpl-withdrawal-address", setRplWithdrawalAddressHandler).RegisterTo(router)
	snroute.Read("/api/node/can-confirm-rpl-withdrawal-address", canConfirmRplWithdrawalAddressHandler).RegisterTo(router)
	snroute.Write("/api/node/confirm-rpl-withdrawal-address", confirmRplWithdrawalAddressHandler).RegisterTo(router)

	// --- Swap RPL ---

	snroute.Read("/api/node/swap-rpl-allowance", swapRplAllowanceHandler).RegisterTo(router)
	snroute.Read("/api/node/can-swap-rpl", canSwapRplHandler).RegisterTo(router)
	snroute.Read("/api/node/get-swap-rpl-approval-gas", getSwapRplApprovalGasHandler).RegisterTo(router)
	snroute.Write("/api/node/swap-rpl-approve-rpl", swapRplApproveRplHandler).RegisterTo(router)
	snroute.Write("/api/node/wait-and-swap-rpl", waitAndSwapRplHandler).RegisterTo(router)
	snroute.Write("/api/node/swap-rpl", swapRplHandler).RegisterTo(router)

	// --- Stake RPL ---

	snroute.Read("/api/node/stake-rpl-allowance", stakeRplAllowanceHandler).RegisterTo(router)
	snroute.Read("/api/node/can-stake-rpl", canStakeRplHandler).RegisterTo(router)
	snroute.Read("/api/node/get-stake-rpl-approval-gas", getStakeRplApprovalGasHandler).RegisterTo(router)
	snroute.Write("/api/node/stake-rpl-approve-rpl", stakeRplApproveRplHandler).RegisterTo(router)
	snroute.Write("/api/node/wait-and-stake-rpl", waitAndStakeRplHandler).RegisterTo(router)
	snroute.Write("/api/node/stake-rpl", stakeRplHandler).RegisterTo(router)

	// --- RPL locking ---

	snroute.Read("/api/node/can-set-rpl-locking-allowed", canSetRplLockingAllowedHandler).RegisterTo(router)
	snroute.Write("/api/node/set-rpl-locking-allowed", setRplLockingAllowedHandler).RegisterTo(router)

	// --- Stake RPL for allowed ---

	snroute.Read("/api/node/can-set-stake-rpl-for-allowed", canSetStakeRplForAllowedHandler).RegisterTo(router)
	snroute.Write("/api/node/set-stake-rpl-for-allowed", setStakeRplForAllowedHandler).RegisterTo(router)

	// --- Withdraw RPL ---

	snroute.Read("/api/node/can-withdraw-rpl", canWithdrawRplHandler).RegisterTo(router)
	snroute.Write("/api/node/withdraw-rpl", withdrawRplHandler).RegisterTo(router)
	snroute.Read("/api/node/can-unstake-legacy-rpl", canUnstakeLegacyRplHandler).RegisterTo(router)
	snroute.Write("/api/node/unstake-legacy-rpl", unstakeLegacyRplHandler).RegisterTo(router)
	snroute.Read("/api/node/can-withdraw-rpl-v131", canWithdrawRplV131Handler).RegisterTo(router)
	snroute.Write("/api/node/withdraw-rpl-v131", withdrawRplV131Handler).RegisterTo(router)
	snroute.Read("/api/node/can-unstake-rpl", canUnstakeRplHandler).RegisterTo(router)
	snroute.Write("/api/node/unstake-rpl", unstakeRplHandler).RegisterTo(router)

	// --- Withdraw ETH / credit ---

	snroute.Read("/api/node/can-withdraw-eth", canWithdrawEthHandler).RegisterTo(router)
	snroute.Write("/api/node/withdraw-eth", withdrawEthHandler).RegisterTo(router)
	snroute.Read("/api/node/can-withdraw-credit", canWithdrawCreditHandler).RegisterTo(router)
	snroute.Write("/api/node/withdraw-credit", withdrawCreditHandler).RegisterTo(router)

	// --- Deposit ---

	snroute.Read("/api/node/can-deposit", canDepositHandler).RegisterTo(router)
	snroute.Write("/api/node/deposit", depositHandler).RegisterTo(router)

	// --- Send / burn ---

	snroute.Read("/api/node/can-send", canSendHandler).RegisterTo(router)
	snroute.Write("/api/node/send", sendHandler).RegisterTo(router)
	snroute.Write("/api/node/send-all", sendAllHandler).RegisterTo(router)
	snroute.Read("/api/node/can-burn", canBurnHandler).RegisterTo(router)
	snroute.Write("/api/node/burn", burnHandler).RegisterTo(router)

	// --- RPL claim ---

	snroute.Read("/api/node/can-claim-rpl-rewards", canClaimRplRewardsHandler).RegisterTo(router)
	snroute.Write("/api/node/claim-rpl-rewards", claimRplRewardsHandler).RegisterTo(router)

	// --- Fee distributor ---

	snroute.Read("/api/node/is-fee-distributor-initialized", isFeeDistributorInitializedHandler).RegisterTo(router)
	snroute.Read("/api/node/get-initialize-fee-distributor-gas", getInitializeFeeDistributorGasHandler).RegisterTo(router)
	snroute.Write("/api/node/initialize-fee-distributor", initializeFeeDistributorHandler).RegisterTo(router)
	snroute.Read("/api/node/can-distribute", canDistributeHandler).RegisterTo(router)
	snroute.Write("/api/node/distribute", distributeHandler).RegisterTo(router)

	// --- Interval rewards ---

	snroute.Read("/api/node/get-rewards-info", getRewardsInfoHandler).RegisterTo(router)
	snroute.Read("/api/node/can-claim-rewards", canClaimRewardsHandler).RegisterTo(router)
	snroute.Write("/api/node/claim-rewards", claimRewardsHandler).RegisterTo(router)
	snroute.Read("/api/node/can-claim-and-stake-rewards", canClaimAndStakeRewardsHandler).RegisterTo(router)
	snroute.Write("/api/node/claim-and-stake-rewards", claimAndStakeRewardsHandler).RegisterTo(router)

	// --- Smoothing pool ---

	snroute.Read("/api/node/get-smoothing-pool-registration-status", getSmoothingPoolRegistrationStatusHandler).RegisterTo(router)
	snroute.Read("/api/node/can-set-smoothing-pool-status", canSetSmoothingPoolStatusHandler).RegisterTo(router)
	snroute.Write("/api/node/set-smoothing-pool-status", setSmoothingPoolStatusHandler).RegisterTo(router)

	// --- ENS ---

	snroute.Read("/api/node/resolve-ens-name", resolveEnsNameHandler).RegisterTo(router)
	snroute.Read("/api/node/reverse-resolve-ens-name", reverseResolveEnsNameHandler).RegisterTo(router)

	// --- Sign ---

	snroute.Write("/api/node/sign-message", signMessageHandler).RegisterTo(router)
	snroute.Write("/api/node/sign", signHandler).RegisterTo(router)

	// --- Vacant minipool ---

	snroute.Read("/api/node/can-create-vacant-minipool", canCreateVacantMinipoolHandler).RegisterTo(router)
	snroute.Write("/api/node/create-vacant-minipool", createVacantMinipoolHandler).RegisterTo(router)

	// --- Send message ---

	snroute.Read("/api/node/can-send-message", canSendMessageHandler).RegisterTo(router)
	snroute.Write("/api/node/send-message", sendMessageHandler).RegisterTo(router)

	// --- Express tickets ---

	snroute.Read("/api/node/get-express-ticket-count", getExpressTicketCountHandler).RegisterTo(router)
	snroute.Read("/api/node/get-express-tickets-provisioned", getExpressTicketsProvisionedHandler).RegisterTo(router)
	snroute.Read("/api/node/can-provision-express-tickets", canProvisionExpressTicketsHandler).RegisterTo(router)
	snroute.Write("/api/node/provision-express-tickets", provisionExpressTicketsHandler).RegisterTo(router)

	// --- Unclaimed rewards ---

	snroute.Read("/api/node/can-claim-unclaimed-rewards", canClaimUnclaimedRewardsHandler).RegisterTo(router)
	snroute.Write("/api/node/claim-unclaimed-rewards", claimUnclaimedRewardsHandler).RegisterTo(router)

	// --- Bond requirement ---

	snroute.Read("/api/node/get-bond-requirement", getBondRequirementHandler).RegisterTo(router)
}

// --- Helper types and functions ---

type depositParams struct {
	count            uint64
	amountWei        *big.Int
	minFee           float64
	salt             *big.Int
	expressTickets   int64
	useCreditBalance bool
	submit           bool
}

func parseDepositParams(r *http.Request, includeExecuteParams bool) (depositParams, error) {
	var p depositParams
	var err error

	p.amountWei, err = parseNodeBigInt(r, "amountWei")
	if err != nil {
		return p, fmt.Errorf("invalid amountWei: %w", err)
	}

	minFeeStr := r.URL.Query().Get("minFee")
	if minFeeStr == "" {
		minFeeStr = r.FormValue("minFee")
	}
	p.minFee, err = strconv.ParseFloat(minFeeStr, 64)
	if err != nil {
		return p, fmt.Errorf("invalid minFee: %w", err)
	}

	p.salt, err = parseNodeBigInt(r, "salt")
	if err != nil {
		return p, fmt.Errorf("invalid salt: %w", err)
	}

	expressStr := r.URL.Query().Get("expressTickets")
	if expressStr == "" {
		expressStr = r.FormValue("expressTickets")
	}
	p.expressTickets, err = strconv.ParseInt(expressStr, 10, 64)
	if err != nil {
		return p, fmt.Errorf("invalid expressTickets: %w", err)
	}

	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		countStr = r.FormValue("count")
	}
	p.count, err = strconv.ParseUint(countStr, 10, 64)
	if err != nil {
		return p, fmt.Errorf("invalid count: %w", err)
	}

	if includeExecuteParams {
		p.useCreditBalance = r.FormValue("useCreditBalance") == "true"
		p.submit = r.FormValue("submit") == "true"
	}

	return p, nil
}

type vacantMinipoolParams struct {
	amountWei *big.Int
	minFee    float64
	salt      *big.Int
	pubkey    rptypes.ValidatorPubkey
}

func parseVacantMinipoolParams(r *http.Request) (vacantMinipoolParams, error) {
	var p vacantMinipoolParams
	var err error

	raw := r.URL.Query().Get("amountWei")
	if raw == "" {
		raw = r.FormValue("amountWei")
	}
	p.amountWei, _ = new(big.Int).SetString(raw, 10)
	if p.amountWei == nil {
		return p, fmt.Errorf("invalid amountWei: %s", raw)
	}

	minFeeStr := r.URL.Query().Get("minFee")
	if minFeeStr == "" {
		minFeeStr = r.FormValue("minFee")
	}
	p.minFee, err = strconv.ParseFloat(minFeeStr, 64)
	if err != nil {
		return p, fmt.Errorf("invalid minFee: %w", err)
	}

	saltStr := r.URL.Query().Get("salt")
	if saltStr == "" {
		saltStr = r.FormValue("salt")
	}
	p.salt, _ = new(big.Int).SetString(saltStr, 10)
	if p.salt == nil {
		return p, fmt.Errorf("invalid salt: %s", saltStr)
	}

	pubkeyStr := r.URL.Query().Get("pubkey")
	if pubkeyStr == "" {
		pubkeyStr = r.FormValue("pubkey")
	}
	pubkeyBytes, err := hex.DecodeString(pubkeyStr)
	if err != nil {
		return p, fmt.Errorf("invalid pubkey hex: %w", err)
	}
	if len(pubkeyBytes) != len(p.pubkey) {
		return p, fmt.Errorf("pubkey must be %d bytes, got %d", len(p.pubkey), len(pubkeyBytes))
	}
	copy(p.pubkey[:], pubkeyBytes)

	return p, nil
}

func parseNodeBigInt(r *http.Request, name string) (*big.Int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	v, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid %s: %s", name, raw)
	}
	return v, nil
}

func parseNodeFloat64(r *http.Request, name string) (float64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseFloat(raw, 64)
}
