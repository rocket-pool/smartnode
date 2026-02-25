package network

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the network module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/network/node-fee", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getNodeFee(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/rpl-price", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRplPrice(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/stats", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStats(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/timezone-map", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getTimezones(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/can-generate-rewards-tree", func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canGenerateRewardsTree(c, index)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/generate-rewards-tree", func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := generateRewardsTree(c, index)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/dao-proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getActiveDAOProposals(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/download-rewards-file", func(w http.ResponseWriter, r *http.Request) {
		interval, err := parseUint64Param(r, "interval")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := downloadRewardsFile(c, interval)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/network/latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getLatestDelegate(c)
		apiutils.WriteResponse(w, resp, err)
	})
}

func parseUint64Param(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
