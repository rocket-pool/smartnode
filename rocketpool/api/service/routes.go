package service

import (
	"net/http"

	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the service module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/service/get-client-status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getClientStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/service/restart-vc", func(w http.ResponseWriter, r *http.Request) {
		resp, err := restartVc(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/service/terminate-data-folder", func(w http.ResponseWriter, r *http.Request) {
		resp, err := terminateDataFolder(c)
		apiutils.WriteResponse(w, resp, err)
	})
}
