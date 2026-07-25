package api

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	apitypes "github.com/rocket-pool/smartnode/shared/types/api"
)

// RegisterWaitRoute registers the /api/wait endpoint on mux.
// It waits for a transaction hash to be mined.
func RegisterWaitRoute(mux *http.ServeMux, c *cli.Command) {
	mux.HandleFunc("/api/wait", func(w http.ResponseWriter, r *http.Request) {
		hash := common.HexToHash(r.URL.Query().Get("txHash"))
		rp, err := services.GetRocketPool(c)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp := apitypes.APIResponse{}
		_, err = utils.WaitForTransactionWithContext(r.Context(), rp.Client, hash)
		response.WriteResponse(w, &resp, err)
	})
}
