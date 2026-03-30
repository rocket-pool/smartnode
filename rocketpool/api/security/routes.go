package security

import (
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterRoutes registers the security module's HTTP routes onto mux.
func RegisterRoutes(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/security/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getStatus(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/members", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMembers(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/proposals", func(w http.ResponseWriter, r *http.Request) {
		resp, err := getProposals(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/proposal-details", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := getProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-propose-leave", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canProposeLeave(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/propose-leave", func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeLeave(c, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-propose-setting", func(w http.ResponseWriter, r *http.Request) {
		contractName := r.URL.Query().Get("contractName")
		settingName := r.URL.Query().Get("settingName")
		value := r.URL.Query().Get("value")
		resp, err := canProposeSetting(c, contractName, settingName, value)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/propose-setting", func(w http.ResponseWriter, r *http.Request) {
		contractName := r.FormValue("contractName")
		settingName := r.FormValue("settingName")
		value := r.FormValue("value")
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSetting(c, contractName, settingName, value, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-cancel-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canCancelProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/cancel-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := cancelProposal(c, id, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canVoteOnProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/vote-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		support := r.FormValue("support") == "true"
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := voteOnProposal(c, id, support, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := canExecuteProposal(c, id)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/execute-proposal", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseUint64(r, "id")
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := executeProposal(c, id, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-join", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canJoin(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/join", func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := join(c, opts)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/can-leave", func(w http.ResponseWriter, r *http.Request) {
		resp, err := canLeave(c)
		apiutils.WriteResponse(w, resp, err)
	})

	mux.HandleFunc("/api/security/leave", func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		resp, err := leave(c, opts)
		apiutils.WriteResponse(w, resp, err)
	})
}

func parseUint64(r *http.Request, name string) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		raw = r.FormValue(name)
	}
	return strconv.ParseUint(raw, 10, 64)
}
