package debug

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/debug/rewards-event", rewardsEventHandler).RegisterTo(router)
}
