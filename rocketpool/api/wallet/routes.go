package wallet

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the wallet module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/wallet/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/set-password", func(w http.ResponseWriter, r *http.Request) {
		password := r.FormValue("password")
		resp, err := setPassword(c, password)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/init", func(w http.ResponseWriter, r *http.Request) {
		derivationPath := r.URL.Query().Get("derivationPath")
		if derivationPath == "" {
			derivationPath = r.FormValue("derivationPath")
		}
		resp, err := initWalletWithPath(c, derivationPath)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		derivationPath := r.FormValue("derivationPath")
		walletIndex, _ := strconv.ParseUint(r.FormValue("walletIndex"), 10, 64)
		resp, err := recoverWalletWithParams(c, mnemonic, skipRecovery, derivationPath, uint(walletIndex))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/search-and-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		address := common.HexToAddress(r.FormValue("address"))
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		resp, err := searchAndRecoverWalletWithParams(c, mnemonic, address, skipRecovery)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/test-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		derivationPath := r.FormValue("derivationPath")
		walletIndex, _ := strconv.ParseUint(r.FormValue("walletIndex"), 10, 64)
		resp, err := testRecoverWalletWithParams(c, mnemonic, skipRecovery, derivationPath, uint(walletIndex))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/test-search-and-recover", func(w http.ResponseWriter, r *http.Request) {
		mnemonic := r.FormValue("mnemonic")
		address := common.HexToAddress(r.FormValue("address"))
		skipRecovery := r.FormValue("skipValidatorKeyRecovery") == "true"
		resp, err := testSearchAndRecoverWalletWithParams(c, mnemonic, address, skipRecovery)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/rebuild", func(w http.ResponseWriter, r *http.Request) {
		resp, err := rebuildWallet(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/export", func(w http.ResponseWriter, r *http.Request) {
		resp, err := exportWallet(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/masquerade", func(w http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.FormValue("address"))
		resp, err := masquerade(c, address)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/end-masquerade", func(w http.ResponseWriter, r *http.Request) {
		resp, err := endMasquerade(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/estimate-gas-set-ens-name", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = r.FormValue("name")
		}
		resp, err := setEnsName(c, name, true)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/wallet/set-ens-name", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		resp, err := setEnsName(c, name, false)
		apiutils.WriteResponse(w, resp, err)
	})
}
