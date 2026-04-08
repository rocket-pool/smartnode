package api

import (
	"net/http"

	"github.com/rocket-pool/smartnode/shared"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

type VersionResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Version string `json:"version"`
}

// RegisterVersionRoute registers the /api/version endpoint on mux.
func RegisterVersionRoute(mux *http.ServeMux) {
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		response := VersionResponse{Version: shared.RocketPoolVersion()}
		apiutils.WriteResponse(w, &response, nil)
	})
}
