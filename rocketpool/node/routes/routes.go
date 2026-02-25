package routes

import (
	"net/http"

	"github.com/urfave/cli"
)

// RegisterRoutes registers all HTTP API routes onto mux.
// Module-specific routes are registered by successive migration branches.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
