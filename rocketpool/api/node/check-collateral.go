package node

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

type nodeCollateralHandler struct {
}

func (h *nodeCollateralHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *nodeCollateralHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
}

func (h *nodeCollateralHandler) PrepareResponse(ctx *callContext, response *api.NodeCheckCollateralResponse) error {
	rp := ctx.rp
	node := ctx.node

	// Check collateral
	collateral, err := rputils.CheckCollateral(rp, node.Address, nil)
	if err != nil {
		return fmt.Errorf("error checking node collateral: %w", err)
	}
	response.EthMatched = collateral.EthMatched
	response.EthMatchedLimit = collateral.EthMatchedLimit
	response.PendingMatchAmount = collateral.PendingMatchAmount

	// Check if there's sufficient collateral including pending bond reductions
	remainingMatch := big.NewInt(0).Sub(response.EthMatchedLimit, response.EthMatched)
	remainingMatch.Sub(remainingMatch, response.PendingMatchAmount)
	response.InsufficientCollateral = (remainingMatch.Cmp(big.NewInt(0)) < 0)
	return nil
}
