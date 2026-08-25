package api

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared"
)

type VersionResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Version string `json:"version"`
}

// RegisterVersionRoute registers the /api/version endpoint on router.
func RegisterVersionRoute(router *snroute.Router) {
	snroute.Read("/api/version", versionHandler).RegisterTo(router)
}

func versionHandler(ctx snroute.Context) {
	resp := VersionResponse{Version: shared.RocketPoolVersion()}
	response.WriteResponse(ctx.Writer, &resp, nil)
}
