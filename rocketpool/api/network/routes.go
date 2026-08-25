package network

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
)

// RegisterRoutes registers the network module's HTTP routes onto router.
func RegisterRoutes(router *snroute.Router, c *cli.Command) {
	router.Handle(snroute.Read("/api/network/node-fee", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getNodeFee(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/rpl-price", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRplPrice(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/stats", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStats(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/timezone-map", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getTimezones(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/can-generate-rewards-tree", func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canGenerateRewardsTree(c, index)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/network/generate-rewards-tree", func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := generateRewardsTree(c, index)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/dao-proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getActiveDAOProposals(c)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Write("/api/network/download-rewards-file", func(w http.ResponseWriter, r *http.Request) {
		interval, err := parseUint64Param(r, "interval")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := downloadRewardsFile(c, interval)
		response.WriteResponse(w, resp, err)
	}))

	router.Handle(snroute.Read("/api/network/latest-delegate", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getLatestDelegate(c)
		response.WriteResponse(w, resp, err)
	}))
}

func parseUint64Param(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
