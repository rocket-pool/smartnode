package routes

import (
	"net/http"

	"github.com/urfave/cli"

	queueroutes "github.com/rocket-pool/smartnode/rocketpool/api/queue"
	serviceroutes "github.com/rocket-pool/smartnode/rocketpool/api/service"
)

// RegisterRoutes registers all HTTP API routes onto mux.
// Each migration branch adds additional module registrations here.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	queueroutes.RegisterRoutes(mux, c)
	serviceroutes.RegisterRoutes(mux, c)
}
