package queue

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the queue module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/queue/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/queue/can-process", canProcessHandler(c)).RegisterTo(router)
	snroute.Write("/api/queue/process", processHandler(c)).RegisterTo(router)
	snroute.Read("/api/queue/get-queue-details", getQueueDetailsHandler(c)).RegisterTo(router)
	snroute.Read("/api/queue/can-assign-deposits", canAssignDepositsHandler(c)).RegisterTo(router)
	snroute.Write("/api/queue/assign-deposits", assignDepositsHandler(c)).RegisterTo(router)
}

func parseUint32Param(r *http.Request, name string) (uint32, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
