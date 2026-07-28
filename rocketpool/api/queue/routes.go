package queue

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
)

// RegisterRoutes registers the queue module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/queue/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/can-process", func(w http.ResponseWriter, r *http.Request) {
		m, err := parseUint32Param(r, "max")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProcessQueue(c, int64(m))
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/process", func(w http.ResponseWriter, r *http.Request) {
		m, err := parseUint32Param(r, "max")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := processQueue(c, int64(m), opts)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/get-queue-details", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getQueueDetails(c)
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/can-assign-deposits", func(w http.ResponseWriter, r *http.Request) {
		m, err := parseUint32Param(r, "max")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canAssignDeposits(c, int64(m))
		response.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/assign-deposits", func(w http.ResponseWriter, r *http.Request) {
		m, err := parseUint32Param(r, "max")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := assignDeposits(c, int64(m), opts)
		response.WriteResponse(w, resp, err)
	})
}

func parseUint32Param(r *http.Request, name string) (uint32, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
