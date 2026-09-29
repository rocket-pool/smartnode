package megapool

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/performance"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

func canChallengePerformanceHandler(ctx snroute.Context)   { performance.CanChallengeHandler(false)(ctx) }
func challengePerformanceHandler(ctx snroute.WriteContext) { performance.ChallengeHandler(false)(ctx) }
