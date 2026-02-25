package api

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	apitypes "github.com/rocket-pool/smartnode/shared/types/api"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// RegisterWaitRoute registers the /api/wait endpoint on mux.
// It waits for a transaction hash to be mined.
func RegisterWaitRoute(mux *http.ServeMux, c *cli.Context) {
	mux.HandleFunc("/api/wait", func(w http.ResponseWriter, r *http.Request) {
		hash := common.HexToHash(r.URL.Query().Get("txHash"))
		rp, err := services.GetRocketPool(c)
		if err != nil {
			apiutils.WriteErrorResponse(w, err)
			return
		}
		response := apitypes.APIResponse{}
		_, err = utils.WaitForTransaction(rp.Client, hash)
		apiutils.WriteResponse(w, &response, err)
	})
}
