package security

import (
	"net/http"
	"strconv"

	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the security module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router) {
	snroute.Read("/api/security/status", statusHandler).RegisterTo(router)
	snroute.Read("/api/security/members", membersHandler).RegisterTo(router)
	snroute.Read("/api/security/proposals", proposalsHandler).RegisterTo(router)
	snroute.Read("/api/security/proposal-details", proposalDetailsHandler).RegisterTo(router)
	snroute.Read("/api/security/can-propose-leave", canProposeLeaveHandler).RegisterTo(router)
	snroute.Write("/api/security/propose-leave", proposeLeaveHandler).RegisterTo(router)
	snroute.Read("/api/security/can-propose-setting", canProposeSettingHandler).RegisterTo(router)
	snroute.Write("/api/security/propose-setting", proposeSettingHandler).RegisterTo(router)
	snroute.Read("/api/security/can-cancel-proposal", canCancelProposalHandler).RegisterTo(router)
	snroute.Write("/api/security/cancel-proposal", cancelProposalHandler).RegisterTo(router)
	snroute.Read("/api/security/can-vote-proposal", canVoteProposalHandler).RegisterTo(router)
	snroute.Write("/api/security/vote-proposal", voteProposalHandler).RegisterTo(router)
	snroute.Read("/api/security/can-execute-proposal", canExecuteProposalHandler).RegisterTo(router)
	snroute.Write("/api/security/execute-proposal", executeProposalHandler).RegisterTo(router)
	snroute.Read("/api/security/can-join", canJoinHandler).RegisterTo(router)
	snroute.Write("/api/security/join", joinHandler).RegisterTo(router)
	snroute.Read("/api/security/can-leave", canLeaveHandler).RegisterTo(router)
	snroute.Write("/api/security/leave", leaveHandler).RegisterTo(router)
}

func parseUint64(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
