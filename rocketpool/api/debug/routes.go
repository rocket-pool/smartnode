package debug

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/debug/rewards-event", rewardsEventHandler(c)).RegisterTo(router)
}
