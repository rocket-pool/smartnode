package debug

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	router.Handle(snroute.Read("/api/debug/rewards-event", func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("interval")
		if raw == "" {
			response.WriteErrorResponse(w, &response.BadRequestError{Err: fmt.Errorf("missing required query parameter: interval")})
			return
		}
		interval, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.WriteErrorResponse(w, &response.BadRequestError{Err: fmt.Errorf("invalid interval: %w", err)})
			return
		}
		resp, err := getRewardsEvent(c, interval)
		response.WriteResponse(w, resp, err)
	}))
}
