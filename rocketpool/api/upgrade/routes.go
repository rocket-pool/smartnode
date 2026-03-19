package upgrade

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// RegisterRoutes registers the upgrade module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/upgrade/get-upgrade-proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getUpgradeProposals(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/upgrade/can-execute-upgrade", func(w http.ResponseWriter, r *http.Request) {
		id, err := cliutils.ValidatePositiveUint("upgrade proposal ID", r.URL.Query().Get("id"))
		if err != nil {
			apiutils.WriteResponse(w, nil, err)
			return
		}
		resp, err := canExecuteUpgrade(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/upgrade/execute-upgrade", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
		if err != nil {
			apiutils.WriteResponse(w, nil, err)
			return
		}
		resp, err := executeUpgrade(c, id)
		apiutils.WriteResponse(w, resp, err)
	})
}
