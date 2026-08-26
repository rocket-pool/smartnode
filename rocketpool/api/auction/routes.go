package auction

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the auction module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/auction/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/auction/lots", lotsHandler).RegisterTo(router)
	snroute.Read("/api/auction/can-create-lot", canCreateLotHandler).RegisterTo(router)
	snroute.Write("/api/auction/create-lot", createLotHandler).RegisterTo(router)
	snroute.Read("/api/auction/can-bid-lot", canBidLotHandler).RegisterTo(router)
	snroute.Write("/api/auction/bid-lot", bidLotHandler).RegisterTo(router)
	snroute.Read("/api/auction/can-claim-lot", canClaimLotHandler).RegisterTo(router)
	snroute.Write("/api/auction/claim-lot", claimLotHandler).RegisterTo(router)
	snroute.Read("/api/auction/can-recover-lot", canRecoverLotHandler).RegisterTo(router)
	snroute.Write("/api/auction/recover-lot", recoverLotHandler).RegisterTo(router)
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
