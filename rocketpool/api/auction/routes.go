package auction

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the auction module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/auction/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/lots", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getLots(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/can-create-lot", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canCreateLot(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/create-lot", func(w http.ResponseWriter, r *http.Request) {
		resp, err := createLot(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/can-bid-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, amountWei, err := parseLotIndexAndAmount(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canBidOnLot(c, lotIndex, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/bid-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, amountWei, err := parseLotIndexAndAmount(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := bidOnLot(c, lotIndex, amountWei)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/can-claim-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canClaimFromLot(c, lotIndex)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/claim-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := claimFromLot(c, lotIndex)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/can-recover-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canRecoverRplFromLot(c, lotIndex)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/auction/recover-lot", func(w http.ResponseWriter, r *http.Request) {
		lotIndex, err := parseLotIndex(r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := recoverRplFromLot(c, lotIndex)
		apiutils.WriteResponse(w, resp, err)
	})
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
