package auction

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
)

// RegisterRoutes registers the auction module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	router.Handle(snroute.Read("/api/auction/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/auction/lots", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getLots(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/auction/can-create-lot", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canCreateLot(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/auction/create-lot", func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := createLot(c, opts)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/auction/can-bid-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, amountWei, err := parseLotIndexAndAmount(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canBidOnLot(c, lotIndex, amountWei)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/auction/bid-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, amountWei, err := parseLotIndexAndAmount(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := bidOnLot(c, lotIndex, amountWei, opts)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/auction/can-claim-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canClaimFromLot(c, lotIndex)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/auction/claim-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := claimFromLot(c, lotIndex, opts)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/auction/can-recover-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canRecoverRplFromLot(c, lotIndex)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/auction/recover-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := recoverRplFromLot(c, lotIndex, opts)
		response.WriteResponse(w, resp, err)
	}))
}

func parseLotIndex(r *http.Request) (uint64, error) {
	raw := r.URL.Query().Get("lotIndex")
	if raw == "" {
		raw = r.FormValue("lotIndex")
	}
	return strconv.ParseUint(raw, 10, 64)
}

func parseLotIndexAndAmount(r *http.Request) (uint64, *big.Int, error) {
	lotIndex, err := parseLotIndex(r)
	if err != nil {
		return 0, nil, err
	}
	raw := r.URL.Query().Get("amountWei")
	if raw == "" {
		raw = r.FormValue("amountWei")
	}
	amountWei, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return 0, nil, fmt.Errorf("invalid amountWei: %s", raw)
	}
	return lotIndex, amountWei, nil
}
