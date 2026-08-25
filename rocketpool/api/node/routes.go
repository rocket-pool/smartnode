package node

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the node module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/node/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/alerts", alertsHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/sync", syncHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/get-eth-balance", getEthBalanceHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/check-collateral", checkCollateralHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/rewards", rewardsHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/deposit-contract-info", depositContractInfoHandler(c)).RegisterTo(router)

	// --- Register ---

	snroute.Read("/api/node/can-register", canRegisterHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/register", registerHandler(c)).RegisterTo(router)

	// --- Timezone ---

	snroute.Read("/api/node/can-set-timezone", canSetTimezoneHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-timezone", setTimezoneHandler(c)).RegisterTo(router)

	// --- Primary withdrawal address ---

	snroute.Read("/api/node/can-set-primary-withdrawal-address", canSetPrimaryWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-primary-withdrawal-address", setPrimaryWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-confirm-primary-withdrawal-address", canConfirmPrimaryWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/confirm-primary-withdrawal-address", confirmPrimaryWithdrawalAddressHandler(c)).RegisterTo(router)

	// --- RPL withdrawal address ---

	snroute.Read("/api/node/can-set-rpl-withdrawal-address", canSetRplWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-rpl-withdrawal-address", setRplWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-confirm-rpl-withdrawal-address", canConfirmRplWithdrawalAddressHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/confirm-rpl-withdrawal-address", confirmRplWithdrawalAddressHandler(c)).RegisterTo(router)

	// --- Swap RPL ---

	snroute.Read("/api/node/swap-rpl-allowance", swapRplAllowanceHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-swap-rpl", canSwapRplHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/get-swap-rpl-approval-gas", getSwapRplApprovalGasHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/swap-rpl-approve-rpl", swapRplApproveRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/wait-and-swap-rpl", waitAndSwapRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/swap-rpl", swapRplHandler(c)).RegisterTo(router)

	// --- Stake RPL ---

	snroute.Read("/api/node/stake-rpl-allowance", stakeRplAllowanceHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-stake-rpl", canStakeRplHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/get-stake-rpl-approval-gas", getStakeRplApprovalGasHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/stake-rpl-approve-rpl", stakeRplApproveRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/wait-and-stake-rpl", waitAndStakeRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/stake-rpl", stakeRplHandler(c)).RegisterTo(router)

	// --- RPL locking ---

	snroute.Read("/api/node/can-set-rpl-locking-allowed", canSetRplLockingAllowedHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-rpl-locking-allowed", setRplLockingAllowedHandler(c)).RegisterTo(router)

	// --- Stake RPL for allowed ---

	snroute.Read("/api/node/can-set-stake-rpl-for-allowed", canSetStakeRplForAllowedHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-stake-rpl-for-allowed", setStakeRplForAllowedHandler(c)).RegisterTo(router)

	// --- Withdraw RPL ---

	snroute.Read("/api/node/can-withdraw-rpl", canWithdrawRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/withdraw-rpl", withdrawRplHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-unstake-legacy-rpl", canUnstakeLegacyRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/unstake-legacy-rpl", unstakeLegacyRplHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-withdraw-rpl-v131", canWithdrawRplV131Handler(c)).RegisterTo(router)
	snroute.Write("/api/node/withdraw-rpl-v131", withdrawRplV131Handler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-unstake-rpl", canUnstakeRplHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/unstake-rpl", unstakeRplHandler(c)).RegisterTo(router)

	// --- Withdraw ETH / credit ---

	snroute.Read("/api/node/can-withdraw-eth", canWithdrawEthHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/withdraw-eth", withdrawEthHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-withdraw-credit", canWithdrawCreditHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/withdraw-credit", withdrawCreditHandler(c)).RegisterTo(router)

	// --- Deposit ---

	snroute.Read("/api/node/can-deposit", canDepositHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/deposit", depositHandler(c)).RegisterTo(router)

	// --- Send / burn ---

	snroute.Read("/api/node/can-send", canSendHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/send", sendHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/send-all", sendAllHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-burn", canBurnHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/burn", burnHandler(c)).RegisterTo(router)

	// --- RPL claim ---

	snroute.Read("/api/node/can-claim-rpl-rewards", canClaimRplRewardsHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/claim-rpl-rewards", claimRplRewardsHandler(c)).RegisterTo(router)

	// --- Fee distributor ---

	snroute.Read("/api/node/is-fee-distributor-initialized", isFeeDistributorInitializedHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/get-initialize-fee-distributor-gas", getInitializeFeeDistributorGasHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/initialize-fee-distributor", initializeFeeDistributorHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-distribute", canDistributeHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/distribute", distributeHandler(c)).RegisterTo(router)

	// --- Interval rewards ---

	snroute.Read("/api/node/get-rewards-info", getRewardsInfoHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-claim-rewards", canClaimRewardsHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/claim-rewards", claimRewardsHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-claim-and-stake-rewards", canClaimAndStakeRewardsHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/claim-and-stake-rewards", claimAndStakeRewardsHandler(c)).RegisterTo(router)

	// --- Smoothing pool ---

	snroute.Read("/api/node/get-smoothing-pool-registration-status", getSmoothingPoolRegistrationStatusHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-set-smoothing-pool-status", canSetSmoothingPoolStatusHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/set-smoothing-pool-status", setSmoothingPoolStatusHandler(c)).RegisterTo(router)

	// --- ENS ---

	snroute.Read("/api/node/resolve-ens-name", resolveEnsNameHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/reverse-resolve-ens-name", reverseResolveEnsNameHandler(c)).RegisterTo(router)

	// --- Sign ---

	snroute.Write("/api/node/sign-message", signMessageHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/sign", signHandler(c)).RegisterTo(router)

	// --- Vacant minipool ---

	snroute.Read("/api/node/can-create-vacant-minipool", canCreateVacantMinipoolHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/create-vacant-minipool", createVacantMinipoolHandler(c)).RegisterTo(router)

	// --- Send message ---

	snroute.Read("/api/node/can-send-message", canSendMessageHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/send-message", sendMessageHandler(c)).RegisterTo(router)

	// --- Express tickets ---

	snroute.Read("/api/node/get-express-ticket-count", getExpressTicketCountHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/get-express-tickets-provisioned", getExpressTicketsProvisionedHandler(c)).RegisterTo(router)
	snroute.Read("/api/node/can-provision-express-tickets", canProvisionExpressTicketsHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/provision-express-tickets", provisionExpressTicketsHandler(c)).RegisterTo(router)

	// --- Unclaimed rewards ---

	snroute.Read("/api/node/can-claim-unclaimed-rewards", canClaimUnclaimedRewardsHandler(c)).RegisterTo(router)
	snroute.Write("/api/node/claim-unclaimed-rewards", claimUnclaimedRewardsHandler(c)).RegisterTo(router)

	// --- Bond requirement ---

	snroute.Read("/api/node/get-bond-requirement", getBondRequirementHandler(c)).RegisterTo(router)
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
