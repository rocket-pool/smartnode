package queue

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli"

	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the queue module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/queue/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/can-process", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint32Param(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProcessQueue(c, int64(max))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/process", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint32Param(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := processQueue(c, int64(max))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/get-queue-details", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getQueueDetails(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/can-assign-deposits", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint32Param(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canAssignDeposits(c, int64(max))
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/queue/assign-deposits", func(w http.ResponseWriter, r *http.Request) {
		max, err := parseUint32Param(r, "max")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := assignDeposits(c, int64(max))
		apiutils.WriteResponse(w, resp, err)
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
