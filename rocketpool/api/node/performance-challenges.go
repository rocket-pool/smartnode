package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

func readPerformanceChallenges(c *cli.Command, id *big.Int) (*api.PerformanceChallengesResponse, error) {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	header, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	opts := &bind.CallOpts{BlockNumber: header.Number}
	result := &api.PerformanceChallengesResponse{}
	result.Saturn2Deployed, err = state.IsSaturn2Deployed(rp, opts)
	if err != nil {
		return nil, err
	}
	if !result.Saturn2Deployed {
		return result, nil
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	interval, err := cfg.GetEventLogInterval()
	if err != nil {
		return nil, err
	}
	challenges, err := megapool.GetPerformanceChallenges(rp, id, big.NewInt(int64(interval)), nil, header.Number, opts)
	if err != nil {
		return nil, err
	}
	for _, challenge := range challenges {
		status, err := megapool.GetPerformanceChallengeStatus(rp, challenge.ChallengeId, opts)
		if err != nil {
			return nil, err
		}
		if id == nil && (status.Responded || status.Finalised) && status.BondSettled {
			continue
		}
		result.Challenges = append(result.Challenges, api.PerformanceChallengeDetails{
			Challenge: challenge,
			State:     status,
		})
	}
	return result, nil
}

func performanceChallengesHandler(ctx snroute.Context) {
	var id *big.Int
	var err error
	if ctx.Request.FormValue("challengeId") != "" {
		id, err = parseChallengeBondID(ctx.Request)
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
	}
	result, err := readPerformanceChallenges(ctx.Command(), id)
	response.WriteResponse(ctx.Writer, result, err)
}

// Recheck eligibility and simulate on POST as well, including when the user
// supplies a gas limit. Finalisation and bond settlement are independent.
func preflightPerformanceChallenge(c *cli.Command, id *big.Int, defend bool, opts *bind.TransactOpts) (*api.CanActOnPerformanceChallengeResponse, megapool.PerformanceChallengeDefense, error) {
	result := &api.CanActOnPerformanceChallengeResponse{}
	var proof megapool.PerformanceChallengeDefense
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, proof, err
	}
	header, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, proof, err
	}
	callOpts := &bind.CallOpts{BlockNumber: header.Number}
	deployed, err := state.IsSaturn2Deployed(rp, callOpts)
	if err != nil {
		return nil, proof, err
	}
	if !deployed {
		result.Reason = "Performance challenges require Saturn 2."
		return result, proof, nil
	}
	status, err := megapool.GetPerformanceChallengeStatus(rp, id, callOpts)
	if err != nil {
		return nil, proof, err
	}
	opts.Value = nil
	if defend {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, proof, err
		}
		if !status.CanDefend(opts.From, header.Time) {
			result.Reason = "The challenge must be unresolved, within its response deadline, and proposed by another node."
			return result, proof, nil
		}
		events, err := readPerformanceChallenges(c, id)
		if err != nil {
			return nil, proof, err
		}
		if len(events.Challenges) != 1 {
			return nil, proof, fmt.Errorf("expected one event for challenge %s, found %d", id, len(events.Challenges))
		}
		proof, err = services.BuildPerformanceChallengeDefense(c, events.Challenges[0].Challenge)
		if err != nil {
			return nil, proof, err
		}
		result.GasLimits, err = proof.EstimateGas(rp, opts)
		if err != nil {
			return nil, proof, err
		}
	} else {
		if !status.CanFinalise(header.Time) {
			result.Reason = "The challenge must be unresolved and past its response deadline."
			return result, proof, nil
		}
		result.GasLimits, err = megapool.EstimateFinaliseChallengeGas(rp, id, opts)
		if err != nil {
			return nil, proof, err
		}
	}
	result.CanAct = true
	return result, proof, nil
}

func canActOnPerformanceChallengeHandler(defend bool) func(snroute.Context) {
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
		result, _, err := preflightPerformanceChallenge(ctx.Command(), id, defend, opts)
		response.WriteResponse(ctx.Writer, result, err)
	}
}

func actOnPerformanceChallengeHandler(defend bool) func(snroute.WriteContext) {
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
		preflight, proof, err := preflightPerformanceChallenge(ctx.Command(), id, defend, transactor.Opts())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		if !preflight.CanAct {
			response.WriteErrorResponse(ctx.Writer, fmt.Errorf("%s", preflight.Reason))
			return
		}
		rp, err := services.GetRocketPool(ctx.Command())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		result := &api.PerformanceChallengeResponse{}
		if defend {
			result.TxHash, err = proof.Submit(rp, transactor.Opts())
		} else {
			result.TxHash, err = megapool.FinaliseChallenge(rp, id, transactor.Opts())
		}
		response.WriteResponse(ctx.Writer, result, err)
	}
}
