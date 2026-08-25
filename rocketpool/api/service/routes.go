package service

import (
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the service module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	router.Handle(snroute.Read("/api/service/get-client-status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getClientStatus(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/service/restart-vc", func(w http.ResponseWriter, r *http.Request) {
		resp, err := restartVc(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/service/terminate-data-folder", func(w http.ResponseWriter, r *http.Request) {
		resp, err := terminateDataFolder(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/service/get-gas-price-from-latest-block", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getGasPriceFromLatestBlock(c)
		response.WriteResponse(w, resp, err)
	}))
}
