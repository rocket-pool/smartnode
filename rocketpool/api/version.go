package api

import (
	"net/http"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared"
)

type VersionResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Version string `json:"version"`
}

// RegisterVersionRoute registers the /api/version endpoint on mux.
func RegisterVersionRoute(mux *http.ServeMux) {
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		resp := VersionResponse{Version: shared.RocketPoolVersion()}
		response.WriteResponse(w, &resp, nil)
	})
}
