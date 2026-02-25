package node

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the node module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/node/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/alerts", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getAlerts(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/sync", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getSyncProgress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/get-eth-balance", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getNodeEthBalance(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/check-collateral", func(w http.ResponseWriter, r *http.Request) {
		resp, err := checkCollateral(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/rewards", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRewards(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/deposit-contract-info", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getDepositContractInfo(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Register ---

	mux.HandleFunc("/api/node/can-register", func(w http.ResponseWriter, r *http.Request) {
		tz := r.URL.Query().Get("timezoneLocation")
		resp, err := canRegisterNode(c, tz)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/register", func(w http.ResponseWriter, r *http.Request) {
		tz := r.FormValue("timezoneLocation")
		resp, err := registerNode(c, tz)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Timezone ---

	mux.HandleFunc("/api/node/can-set-timezone", func(w http.ResponseWriter, r *http.Request) {
		tz := r.URL.Query().Get("timezoneLocation")
		resp, err := canSetTimezoneLocation(c, tz)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-timezone", func(w http.ResponseWriter, r *http.Request) {
		tz := r.FormValue("timezoneLocation")
		resp, err := setTimezoneLocation(c, tz)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Primary withdrawal address ---

	mux.HandleFunc("/api/node/can-set-primary-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		confirm := r.URL.Query().Get("confirm") == "true"
		resp, err := canSetPrimaryWithdrawalAddress(c, addr, confirm)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-primary-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.FormValue("address"))
		confirm := r.FormValue("confirm") == "true"
		resp, err := setPrimaryWithdrawalAddress(c, addr, confirm)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-confirm-primary-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canConfirmPrimaryWithdrawalAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/confirm-primary-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := confirmPrimaryWithdrawalAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- RPL withdrawal address ---

	mux.HandleFunc("/api/node/can-set-rpl-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		confirm := r.URL.Query().Get("confirm") == "true"
		resp, err := canSetRPLWithdrawalAddress(c, addr, confirm)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-rpl-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.FormValue("address"))
		confirm := r.FormValue("confirm") == "true"
		resp, err := setRPLWithdrawalAddress(c, addr, confirm)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-confirm-rpl-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canConfirmRPLWithdrawalAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/confirm-rpl-withdrawal-address", func(w http.ResponseWriter, r *http.Request) {
		resp, err := confirmRPLWithdrawalAddress(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Swap RPL ---

	mux.HandleFunc("/api/node/swap-rpl-allowance", func(w http.ResponseWriter, r *http.Request) {
		resp, err := allowanceFsRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-swap-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeSwapRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/get-swap-rpl-approval-gas", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getSwapApprovalGas(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/swap-rpl-approve-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := approveFsRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/wait-and-swap-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		hash := common.HexToHash(r.FormValue("approvalTxHash"))
		resp, err := waitForApprovalAndSwapFsRpl(c, amountWei, hash)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/swap-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := swapRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Stake RPL ---

	mux.HandleFunc("/api/node/stake-rpl-allowance", func(w http.ResponseWriter, r *http.Request) {
		resp, err := allowanceRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-stake-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeStakeRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/get-stake-rpl-approval-gas", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getStakeApprovalGas(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/stake-rpl-approve-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := approveRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/wait-and-stake-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		hash := common.HexToHash(r.FormValue("approvalTxHash"))
		resp, err := waitForApprovalAndStakeRpl(c, amountWei, hash)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/stake-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := stakeRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- RPL locking ---

	mux.HandleFunc("/api/node/can-set-rpl-locking-allowed", func(w http.ResponseWriter, r *http.Request) {
		allowed := r.URL.Query().Get("allowed") == "true"
		resp, err := canSetRplLockAllowed(c, allowed)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-rpl-locking-allowed", func(w http.ResponseWriter, r *http.Request) {
		allowed := r.FormValue("allowed") == "true"
		resp, err := setRplLockAllowed(c, allowed)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Stake RPL for allowed ---

	mux.HandleFunc("/api/node/can-set-stake-rpl-for-allowed", func(w http.ResponseWriter, r *http.Request) {
		caller := common.HexToAddress(r.URL.Query().Get("caller"))
		allowed := r.URL.Query().Get("allowed") == "true"
		resp, err := canSetStakeRplForAllowed(c, caller, allowed)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-stake-rpl-for-allowed", func(w http.ResponseWriter, r *http.Request) {
		caller := common.HexToAddress(r.FormValue("caller"))
		allowed := r.FormValue("allowed") == "true"
		resp, err := setStakeRplForAllowed(c, caller, allowed)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Withdraw RPL ---

	mux.HandleFunc("/api/node/can-withdraw-rpl", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canNodeWithdrawRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/withdraw-rpl", func(w http.ResponseWriter, r *http.Request) {
		resp, err := nodeWithdrawRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-unstake-legacy-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeUnstakeLegacyRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/unstake-legacy-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeUnstakeLegacyRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-withdraw-rpl-v131", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeWithdrawRplv1_3_1(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/withdraw-rpl-v131", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeWithdrawRplv1_3_1(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-unstake-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeUnstakeRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/unstake-rpl", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeUnstakeRpl(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Withdraw ETH / credit ---

	mux.HandleFunc("/api/node/can-withdraw-eth", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeWithdrawEth(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/withdraw-eth", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeWithdrawEth(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-withdraw-credit", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeWithdrawCredit(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/withdraw-credit", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeWithdrawCredit(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Deposit ---

	mux.HandleFunc("/api/node/can-deposit", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseDepositParams(r, false)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNodeDeposits(c, params.count, params.amountWei, params.minFee, params.salt, params.expressTickets)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/deposit", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseDepositParams(r, true)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := nodeDeposits(c, params.count, params.amountWei, params.minFee, params.salt, params.useCreditBalance, params.expressTickets, params.submit)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Send / burn ---

	mux.HandleFunc("/api/node/can-send", func(w http.ResponseWriter, r *http.Request) {
		amountRaw, err := parseNodeFloat64(r, "amountRaw")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		token := r.URL.Query().Get("token")
		to := common.HexToAddress(r.URL.Query().Get("to"))
		resp, err := canNodeSend(c, amountRaw, token, to)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/send", func(w http.ResponseWriter, r *http.Request) {
		amountRaw, err := parseNodeFloat64(r, "amountRaw")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		token := r.FormValue("token")
		to := common.HexToAddress(r.FormValue("to"))
		resp, err := nodeSend(c, amountRaw, token, to)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/send-all", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		to := common.HexToAddress(r.FormValue("to"))
		resp, err := nodeSendAllTokens(c, token, to)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-burn", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		token := r.URL.Query().Get("token")
		resp, err := canNodeBurn(c, amountWei, token)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/burn", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseNodeBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		token := r.FormValue("token")
		resp, err := nodeBurn(c, amountWei, token)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- RPL claim ---

	mux.HandleFunc("/api/node/can-claim-rpl-rewards", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canNodeClaimRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/claim-rpl-rewards", func(w http.ResponseWriter, r *http.Request) {
		resp, err := nodeClaimRpl(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Fee distributor ---

	mux.HandleFunc("/api/node/is-fee-distributor-initialized", func(w http.ResponseWriter, r *http.Request) {
		resp, err := isFeeDistributorInitialized(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/get-initialize-fee-distributor-gas", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getInitializeFeeDistributorGas(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/initialize-fee-distributor", func(w http.ResponseWriter, r *http.Request) {
		resp, err := initializeFeeDistributor(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-distribute", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canDistribute(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/distribute", func(w http.ResponseWriter, r *http.Request) {
		resp, err := distribute(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Interval rewards ---

	mux.HandleFunc("/api/node/get-rewards-info", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRewardsInfo(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-claim-rewards", func(w http.ResponseWriter, r *http.Request) {
		indices := r.URL.Query().Get("indices")
		resp, err := canClaimRewards(c, indices)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/claim-rewards", func(w http.ResponseWriter, r *http.Request) {
		indices := r.FormValue("indices")
		resp, err := claimRewards(c, indices)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-claim-and-stake-rewards", func(w http.ResponseWriter, r *http.Request) {
		indices := r.URL.Query().Get("indices")
		stakeAmount, err := parseNodeBigInt(r, "stakeAmount")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canClaimAndStakeRewards(c, indices, stakeAmount)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/claim-and-stake-rewards", func(w http.ResponseWriter, r *http.Request) {
		indices := r.FormValue("indices")
		stakeAmount, err := parseNodeBigInt(r, "stakeAmount")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := claimAndStakeRewards(c, indices, stakeAmount)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Smoothing pool ---

	mux.HandleFunc("/api/node/get-smoothing-pool-registration-status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getSmoothingPoolRegistrationStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-set-smoothing-pool-status", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status") == "true"
		resp, err := canSetSmoothingPoolStatus(c, status)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/set-smoothing-pool-status", func(w http.ResponseWriter, r *http.Request) {
		status := r.FormValue("status") == "true"
		resp, err := setSmoothingPoolStatus(c, status)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- ENS ---

	mux.HandleFunc("/api/node/resolve-ens-name", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		resp, err := resolveEnsName(c, name)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/reverse-resolve-ens-name", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := reverseResolveEnsName(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Sign ---

	mux.HandleFunc("/api/node/sign-message", func(w http.ResponseWriter, r *http.Request) {
		message := r.FormValue("message")
		resp, err := signMessage(c, message)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/sign", func(w http.ResponseWriter, r *http.Request) {
		serializedTx := r.FormValue("serializedTx")
		resp, err := sign(c, serializedTx)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Vacant minipool ---

	mux.HandleFunc("/api/node/can-create-vacant-minipool", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseVacantMinipoolParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canCreateVacantMinipool(c, params.amountWei, params.minFee, params.salt, params.pubkey)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/create-vacant-minipool", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseVacantMinipoolParams(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := createVacantMinipool(c, params.amountWei, params.minFee, params.salt, params.pubkey)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Send message ---

	mux.HandleFunc("/api/node/can-send-message", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		msgBytes, err := hex.DecodeString(r.URL.Query().Get("message"))
		if err != nil {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid message hex: %w", err))
			return
		}
		resp, err := canSendMessage(c, addr, msgBytes)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/send-message", func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.FormValue("address"))
		msgBytes, err := hex.DecodeString(r.FormValue("message"))
		if err != nil {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid message hex: %w", err))
			return
		}
		resp, err := sendMessage(c, addr, msgBytes)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Express tickets ---

	mux.HandleFunc("/api/node/get-express-ticket-count", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getExpressTicketCount(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/get-express-tickets-provisioned", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getExpressTicketsProvisioned(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/can-provision-express-tickets", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canProvisionExpressTickets(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/provision-express-tickets", func(w http.ResponseWriter, r *http.Request) {
		resp, err := provisionExpressTickets(c)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Unclaimed rewards ---

	mux.HandleFunc("/api/node/can-claim-unclaimed-rewards", func(w http.ResponseWriter, r *http.Request) {
		nodeAddr := common.HexToAddress(r.URL.Query().Get("nodeAddress"))
		resp, err := canClaimUnclaimedRewards(c, nodeAddr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/node/claim-unclaimed-rewards", func(w http.ResponseWriter, r *http.Request) {
		nodeAddr := common.HexToAddress(r.FormValue("nodeAddress"))
		resp, err := claimUnclaimedRewards(c, nodeAddr)
		apiutils.WriteResponse(w, resp, err)
	})

	// --- Bond requirement ---

	mux.HandleFunc("/api/node/get-bond-requirement", func(w http.ResponseWriter, r *http.Request) {
		numValidators, err := strconv.ParseUint(r.URL.Query().Get("numValidators"), 10, 64)
		if err != nil {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid numValidators: %w", err))
			return
		}
		resp, err := getBondRequirement(c, numValidators)
		apiutils.WriteResponse(w, resp, err)
	})
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
