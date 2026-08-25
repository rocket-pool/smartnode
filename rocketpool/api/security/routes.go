package security

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the security module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	snroute.Read("/api/security/status", statusHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/members", membersHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/proposals", proposalsHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/proposal-details", proposalDetailsHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-propose-leave", canProposeLeaveHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/propose-leave", proposeLeaveHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-propose-setting", canProposeSettingHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/propose-setting", proposeSettingHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-cancel-proposal", canCancelProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/cancel-proposal", cancelProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-vote-proposal", canVoteProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/vote-proposal", voteProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-execute-proposal", canExecuteProposalHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/execute-proposal", executeProposalHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-join", canJoinHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/join", joinHandler(c)).RegisterTo(router)
	snroute.Read("/api/security/can-leave", canLeaveHandler(c)).RegisterTo(router)
	snroute.Write("/api/security/leave", leaveHandler(c)).RegisterTo(router)
}

func parseUint64(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
