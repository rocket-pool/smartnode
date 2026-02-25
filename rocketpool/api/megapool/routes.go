package megapool

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the megapool module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/megapool/status", func(w http.ResponseWriter, r *http.Request) {
		finalizedState := r.URL.Query().Get("finalizedState") == "true"
		resp, err := getStatus(c, finalizedState)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/validator-map-and-balances", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getValidatorMapAndBalances(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-claim-refund", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canClaimRefund(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/claim-refund", func(w http.ResponseWriter, r *http.Request) {
		resp, err := claimRefund(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-repay-debt", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canRepayDebt(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/repay-debt", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := repayDebt(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-reduce-bond", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canReduceBond(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/reduce-bond", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := reduceBond(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-stake", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint64(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canStake(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/stake", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint64(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := stake(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-dissolve-validator", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDissolveValidator(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/dissolve-validator", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := dissolveValidator(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-dissolve-with-proof", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDissolveWithProof(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/dissolve-with-proof", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := dissolveWithProof(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-exit-validator", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExitValidator(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/exit-validator", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := exitValidator(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-notify-validator-exit", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNotifyValidatorExit(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/notify-validator-exit", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := notifyValidatorExit(c, validatorId)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-notify-final-balance", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		slot, err := parseUint64(r, "slot")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canNotifyFinalBalance(c, validatorId, slot)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/notify-final-balance", func(w http.ResponseWriter, r *http.Request) {
		validatorId, err := parseUint32(r, "validatorId")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		slot, err := parseUint64(r, "slot")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := notifyFinalBalance(c, validatorId, slot)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-exit-queue", func(w http.ResponseWriter, r *http.Request) {
		validatorIndex, err := parseUint32(r, "validatorIndex")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExitQueue(c, validatorIndex)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/exit-queue", func(w http.ResponseWriter, r *http.Request) {
		validatorIndex, err := parseUint32(r, "validatorIndex")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := exitQueue(c, validatorIndex)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-distribute", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canDistributeMegapool(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/distribute", func(w http.ResponseWriter, r *http.Request) {
		resp, err := distributeMegapool(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/get-new-validator-bond-requirement", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getNewValidatorBondRequirement(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/pending-rewards", func(w http.ResponseWriter, r *http.Request) {
		resp, err := calculatePendingRewards(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/calculate-rewards", func(w http.ResponseWriter, r *http.Request) {
		amountWei, err := parseBigInt(r, "amountWei")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := calculateRewards(c, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/get-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := getUseLatestDelegate(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := canDelegateUpgrade(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.FormValue("address"))
		resp, err := delegateUpgrade(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/can-set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := canSetUseLatestDelegate(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.FormValue("address"))
		resp, err := setUseLatestDelegate(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/get-delegate", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := getDelegate(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/megapool/get-effective-delegate", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := getEffectiveDelegate(c, address)
		apiutils.WriteResponse(w, resp, err)
	})
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

func parseBigInt(r *http.Request, name string) (*big.Int, error) {
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
