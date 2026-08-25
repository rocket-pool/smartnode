package auction

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the auction module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/auction/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/auction/lots", lotsHandler(c)).RegisterTo(router)
	snroute.Read("/api/auction/can-create-lot", canCreateLotHandler(c)).RegisterTo(router)
	snroute.Write("/api/auction/create-lot", createLotHandler(c)).RegisterTo(router)
	snroute.Read("/api/auction/can-bid-lot", canBidLotHandler(c)).RegisterTo(router)
	snroute.Write("/api/auction/bid-lot", bidLotHandler(c)).RegisterTo(router)
	snroute.Read("/api/auction/can-claim-lot", canClaimLotHandler(c)).RegisterTo(router)
	snroute.Write("/api/auction/claim-lot", claimLotHandler(c)).RegisterTo(router)
	snroute.Read("/api/auction/can-recover-lot", canRecoverLotHandler(c)).RegisterTo(router)
	snroute.Write("/api/auction/recover-lot", recoverLotHandler(c)).RegisterTo(router)
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
