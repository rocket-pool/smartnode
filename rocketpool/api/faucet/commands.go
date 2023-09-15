package faucet

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	types "github.com/rocket-pool/smartnode/shared/types/api"
)

// Register routes
func RegisterRoutes(router *mux.Router, name string) {
	route := "faucet"

	// Status
	router.HandleFunc(fmt.Sprintf("/%s/status", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runFaucetCall[types.FaucetStatusData](&faucetStatusHandler{})
		server.HandleResponse(w, response, err)
	})

	// Withdraw RPL
	router.HandleFunc(fmt.Sprintf("/%s/withdraw-rpl", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runFaucetCall[types.FaucetWithdrawRplData](&faucetWithdrawHandler{})
		server.HandleResponse(w, response, err)
	})
}
