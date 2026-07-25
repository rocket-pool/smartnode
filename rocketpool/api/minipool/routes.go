package minipool

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
)

// RegisterRoutes registers the minipool module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/minipool/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-refund", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canRefundMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/refund", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := refundMinipool(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-stake", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canStakeMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/stake", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := stakeMinipool(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-promote", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canPromoteMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/promote", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := promoteMinipool(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-dissolve", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDissolveMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/dissolve", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := dissolveMinipool(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-exit", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExitMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/exit", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := exitMinipool(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-minipool-close-details-for-node", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMinipoolCloseDetailsForNode(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/close", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		bundle := r.FormValue("bundle") == "true"
		resp, err := closeMinipool(c, addr, opts, bundle)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canDelegateUpgrade(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/delegate-upgrade", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := delegateUpgrade(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canSetUseLatestDelegate(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/set-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setUseLatestDelegate(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-use-latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := getUseLatestDelegate(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := getDelegate(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-effective-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := getEffectiveDelegate(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-previous-delegate", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := getPreviousDelegate(c, addr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-vanity-artifacts", func(w http.ResponseWriter, r *http.Request) {
		depositAmountStr := r.URL.Query().Get("depositAmount")
		depositAmount, ok := new(big.Int).SetString(depositAmountStr, 10)
		if !ok {
			response.WriteErrorResponse(w, fmt.Errorf("invalid depositAmount: %s", depositAmountStr))
			return
		}
		nodeAddressStr := r.URL.Query().Get("nodeAddress")
		resp, err := getVanityArtifacts(c, depositAmount, nodeAddressStr)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-distribute-balance-details", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getDistributeBalanceDetails(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/distribute-balance", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := distributeBalance(c, addr, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/import-key", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.FormValue("mnemonic")
		resp, err := importKey(c, addr, mnemonic)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/can-change-withdrawal-creds", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.URL.Query().Get("mnemonic")
		resp, err := canChangeWithdrawalCreds(c, addr, mnemonic)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/change-withdrawal-creds", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		mnemonic := r.FormValue("mnemonic")
		resp, err := changeWithdrawalCreds(c, addr, mnemonic)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/get-rescue-dissolved-details-for-node", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMinipoolRescueDissolvedDetailsForNode(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/minipool/rescue-dissolved", func(w http.ResponseWriter, r *http.Request) {
		addr, err := parseAddress(r, "address")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		amountStr := r.FormValue("amount")
		amount, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			response.WriteErrorResponse(w, fmt.Errorf("invalid amount: %s", amountStr))
			return
		}
		submit := r.FormValue("submit") == "true"
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := rescueDissolvedMinipool(c, addr, amount, submit, opts)
		response.WriteResponse(w, resp, err)
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
