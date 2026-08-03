package wallet

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// RegisterRoutes registers the wallet module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/wallet/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/set-password", func(w http.ResponseWriter, r *http.Request) {
		password := r.FormValue("password")
		resp, err := setPassword(c, password)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/init", func(w http.ResponseWriter, r *http.Request) {
		derivationPath := r.URL.Query().Get("derivationPath")
		if derivationPath == "" {
			derivationPath = r.FormValue("derivationPath")
		}
		resp, err := initWalletWithPath(c, derivationPath)
		response.WriteResponse(w, resp, err)
	})

	// Reports on the recovery currently holding the lock below, so the CLI can
	// explain a rejected command rather than just failing
	mux.HandleFunc("/api/wallet/recovery-status", func(w http.ResponseWriter, r *http.Request) {
		response.WriteResponse(w, &api.KeyRecoveryStatusResponse{Recovery: activeRecovery.status()}, nil)
	})

	mux.HandleFunc("/api/wallet/recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		derivationPath := r.FormValue("derivationPath")
		walletIndex, _ := strconv.ParseUint(r.FormValue("walletIndex"), 10, 64)
		resp, err := withRecoveryLock("wallet recover", func() (*api.RecoverWalletResponse, error) {
			return recoverWalletWithParams(c, mnemonic, skipRecovery, derivationPath, uint(walletIndex))
		})
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/search-and-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		address := common.HexToAddress(r.FormValue("address"))
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		resp, err := withRecoveryLock("wallet recover --address", func() (*api.SearchAndRecoverWalletResponse, error) {
			return searchAndRecoverWalletWithParams(c, mnemonic, address, skipRecovery)
		})
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/test-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		derivationPath := r.FormValue("derivationPath")
		walletIndex, _ := strconv.ParseUint(r.FormValue("walletIndex"), 10, 64)
		resp, err := withRecoveryLock("wallet test-recovery", func() (*api.RecoverWalletResponse, error) {
			return testRecoverWalletWithParams(c, mnemonic, skipRecovery, derivationPath, uint(walletIndex))
		})
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/test-search-and-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		address := common.HexToAddress(r.FormValue("address"))
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		resp, err := withRecoveryLock("wallet test-recovery --address", func() (*api.SearchAndRecoverWalletResponse, error) {
			return testSearchAndRecoverWalletWithParams(c, mnemonic, address, skipRecovery)
		})
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/rebuild", func(w http.ResponseWriter, r *http.Request) {
		resp, err := withRecoveryLock("wallet rebuild", func() (*api.RebuildWalletResponse, error) {
			return rebuildWallet(c)
		})
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/export", func(w http.ResponseWriter, r *http.Request) {
		resp, err := exportWallet(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/masquerade", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.FormValue("address"))
		observe := r.FormValue("observe") == "true"
		resp, err := masquerade(c, address, observe)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/end-masquerade", func(w http.ResponseWriter, r *http.Request) {
		resp, err := endMasquerade(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/estimate-gas-set-ens-name", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = r.FormValue("name")
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setEnsName(c, name, true, opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/set-ens-name", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setEnsName(c, name, false, opts)
		response.WriteResponse(w, resp, err)
	})
}
