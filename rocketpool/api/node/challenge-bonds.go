package node

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	rpnode "github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

func parseChallengeBondID(r *http.Request) (*big.Int, error) {
	id, ok := new(big.Int).SetString(r.FormValue("challengeId"), 10)
	if !ok || id.Sign() <= 0 || id.BitLen() > 256 {
		return nil, &response.BadRequestError{Err: fmt.Errorf("challengeId must be a positive uint256")}
	}
	return id, nil
}

// preflightChallengeBond is also called just before submission, even with a
// user-supplied gas limit, so already-settled bonds cannot bypass simulation.
func preflightChallengeBond(c *cli.Command, id *big.Int, claim bool, opts *bind.TransactOpts) (*api.CanSettleChallengeBondResponse, error) {
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	header, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	callOpts := &bind.CallOpts{BlockNumber: header.Number}
	result := &api.CanSettleChallengeBondResponse{}
	result.Saturn2Deployed, err = state.IsSaturn2Deployed(rp, callOpts)
	if err != nil {
		return nil, err
	}
	if !result.Saturn2Deployed {
		result.Reason = "Challenge bond settlement is not available until Saturn 2 is deployed."
		return result, nil
	}
	if claim {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, err
		}
	}
	status, err := megapool.GetPerformanceChallengeStatus(rp, id, callOpts)
	if err != nil {
		return nil, err
	}
	result.Proposer, result.Responder, result.BondAmount = status.Proposer, status.Responder, status.BondAmount
	if status.BondSettled {
		result.Reason = "The challenge bond has already been settled."
		return result, nil
	}
	if claim && !status.CanClaimReward(opts.From) {
		result.Reason = "Only the recorded defender can claim the reward of a defeated challenge."
		return result, nil
	}
	if !claim && !status.CanReleaseBond(header.Time) {
		result.Reason = "The challenge must be undefeated and its response deadline must have passed."
		return result, nil
	}
	// Both contract methods are nonpayable.
	opts.Value = nil
	if claim {
		stake, err := rpnode.GetNodeStakedRPL(rp, status.Proposer, callOpts)
		if err != nil {
			return nil, err
		}
		result.EstimatedReward, result.EstimatedBurn = challengeRewardAmounts(status.BondAmount, stake)
		result.GasLimits, err = megapool.EstimateClaimChallengeRewardGas(rp, id, opts)
		if err != nil {
			return nil, err
		}
	} else {
		result.GasLimits, err = megapool.EstimateReleaseChallengeBondGas(rp, id, opts)
	}
	if err != nil {
		return nil, err
	}
	result.CanSettle = true
	return result, nil
}

func challengeRewardAmounts(bond, stake *big.Int) (*big.Int, *big.Int) {
	recovered := new(big.Int).Set(bond)
	if recovered.Cmp(stake) > 0 {
		recovered.Set(stake)
	}
	burn := new(big.Int).Div(recovered, big.NewInt(5))
	return new(big.Int).Sub(recovered, burn), burn
}

func canSettleChallengeBondHandler(claim bool) func(snroute.Context) {
	return func(ctx snroute.Context) {
		id, err := parseChallengeBondID(ctx.Request)
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		w, err := services.GetWallet(ctx.Command())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		result, err := preflightChallengeBond(ctx.Command(), id, claim, opts)
		response.WriteResponse(ctx.Writer, result, err)
	}
}

func settleChallengeBondHandler(claim bool) func(snroute.WriteContext) {
	return func(ctx snroute.WriteContext) {
		id, err := parseChallengeBondID(ctx.Request)
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		transactor, err := ctx.Transactor()
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		preflight, err := preflightChallengeBond(ctx.Command(), id, claim, transactor.Opts())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		if !preflight.CanSettle {
			response.WriteErrorResponse(ctx.Writer, fmt.Errorf("%s", preflight.Reason))
			return
		}
		rp, err := services.GetRocketPool(ctx.Command())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		result := &api.SettleChallengeBondResponse{}
		if claim {
			result.TxHash, err = megapool.ClaimChallengeReward(rp, id, transactor.Opts())
		} else {
			result.TxHash, err = megapool.ReleaseChallengeBond(rp, id, transactor.Opts())
		}
		response.WriteResponse(ctx.Writer, result, err)
	}
}
