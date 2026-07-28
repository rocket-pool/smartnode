package service

import (
	"net/http"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/urfave/cli/v3"
)

// RegisterRoutes registers the service module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/service/get-client-status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getClientStatus(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/service/restart-vc", func(w http.ResponseWriter, r *http.Request) {
		resp, err := restartVc(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/service/terminate-data-folder", func(w http.ResponseWriter, r *http.Request) {
		resp, err := terminateDataFolder(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/service/get-gas-price-from-latest-block", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getGasPriceFromLatestBlock(c)
		response.WriteResponse(w, resp, err)
	})
}
