package debug

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/debug/rewards-event", func(w http.ResponseWriter, r *http.Request) {
		raw := r.URL.Query().Get("interval")
		if raw == "" {
			apiutils.WriteErrorResponse(w, &apiutils.BadRequestError{Err: fmt.Errorf("missing required query parameter: interval")})
			return
		}
		interval, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			apiutils.WriteErrorResponse(w, &apiutils.BadRequestError{Err: fmt.Errorf("invalid interval: %w", err)})
			return
		}
		resp, err := getRewardsEvent(c, interval)
		apiutils.WriteResponse(w, resp, err)
	})
}
