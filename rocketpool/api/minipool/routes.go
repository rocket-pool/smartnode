package minipool

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the minipool module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/minipool/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-refund", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canRefundMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/refund", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := refundMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-stake", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canStakeMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/stake", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := stakeMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-promote", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canPromoteMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/promote", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := promoteMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-dissolve", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDissolveMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/dissolve", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := dissolveMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-exit", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExitMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/exit", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := exitMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-minipool-close-details-for-node", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMinipoolCloseDetailsForNode(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/close", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := closeMinipool(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDelegateUpgrade(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := delegateUpgrade(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canSetUseLatestDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := setUseLatestDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getUseLatestDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-effective-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getEffectiveDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-previous-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getPreviousDelegate(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-vanity-artifacts", func(w http.ResponseWriter, r *http.Request) {
		depositAmountStr := r.URL.Query().Get("depositAmount")
		depositAmount, ok := new(big.Int).SetString(depositAmountStr, 10)
		if !ok {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid depositAmount: %s", depositAmountStr))
			return
		}
		nodeAddressStr := r.URL.Query().Get("nodeAddress")
		resp, err := getVanityArtifacts(c, depositAmount, nodeAddressStr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-begin-reduce-bond-amount", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		amountStr := r.URL.Query().Get("newBondAmountWei")
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid newBondAmountWei: %s", amountStr))
			return
		}
		resp, err := canBeginReduceBondAmount(c, addr, amount)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/begin-reduce-bond-amount", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		amountStr := r.FormValue("newBondAmountWei")
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid newBondAmountWei: %s", amountStr))
			return
		}
		resp, err := beginReduceBondAmount(c, addr, amount)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-reduce-bond-amount", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canReduceBondAmount(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/reduce-bond-amount", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := reduceBondAmount(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-distribute-balance-details", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getDistributeBalanceDetails(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/distribute-balance", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := distributeBalance(c, addr)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/import-key", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.FormValue("mnemonic")
		resp, err := importKey(c, addr, mnemonic)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-change-withdrawal-creds", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.URL.Query().Get("mnemonic")
		resp, err := canChangeWithdrawalCreds(c, addr, mnemonic)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/change-withdrawal-creds", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.FormValue("mnemonic")
		resp, err := changeWithdrawalCreds(c, addr, mnemonic)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-rescue-dissolved-details-for-node", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMinipoolRescueDissolvedDetailsForNode(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/rescue-dissolved", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		amountStr := r.FormValue("amount")
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			apiutils.WriteErrorResponse(w, fmt.Errorf("invalid amount: %s", amountStr))
			return
		}
		submit := r.FormValue("submit") == "true"
		resp, err := rescueDissolvedMinipool(c, addr, amount, submit)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-bond-reduction-enabled", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getBondReductionEnabled(c)
		apiutils.WriteResponse(w, resp, err)
	})
}

func parseAddress(r *http.Request, name string) (common.Address, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	if raw == "" {
		return common.Address{}, fmt.Errorf("missing required parameter: %s", name)
	}
	return common.HexToAddress(raw), nil
}
