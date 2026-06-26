package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/minipool"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// verifyPerformance computes a minipool validator's RPIP-73 target-vote
// performance over the inclusive epoch range [startEpoch, endEpoch].
func verifyPerformance(
	c *cli.Command,
	minipoolAddress common.Address,
	startEpoch uint64,
	endEpoch uint64,
) (*api.VerifyPerformanceResponse, error) {
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool %s pubkey: %w", minipoolAddress.Hex(), err)
	}

	return performance.VerifyPerformance(rp, bc, pubkey, startEpoch, endEpoch)
}

func verifyPerformanceHandler(ctx snroute.Context) {
	addr, err := parseAddress(ctx.Request, "address")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	startEpoch, err := parseUint64Param(ctx.Request, "startEpoch")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	endEpoch, err := parseUint64Param(ctx.Request, "endEpoch")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := verifyPerformance(ctx.Command(), addr, startEpoch, endEpoch)
	response.WriteResponse(ctx.Writer, resp, err)
}
