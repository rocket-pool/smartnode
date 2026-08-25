package api

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	apitypes "github.com/rocket-pool/smartnode/shared/types/api"
)

// RegisterWaitRoute registers the /api/wait endpoint on router.
// It waits for a transaction hash to be mined.
func RegisterWaitRoute(router *snroute.Router) {
	snroute.Read("/api/wait", waitHandler).RegisterTo(router)
}

func waitHandler(ctx snroute.Context) {
	hash := common.HexToHash(ctx.Request.URL.Query().Get("txHash"))
	rp, err := services.GetRocketPool(ctx.Command())
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp := apitypes.APIResponse{}
	_, err = utils.WaitForTransactionWithContext(ctx.Request.Context(), rp.Client, hash)
	response.WriteResponse(ctx.Writer, &resp, err)
}
