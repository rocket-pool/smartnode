package performance

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func parseRequest(r *http.Request, minipools bool) (services.PerformanceChallengeRequest, error) {
	var out services.PerformanceChallengeRequest
	bad := func(err error) (services.PerformanceChallengeRequest, error) {
		return out, &response.BadRequestError{Err: err}
	}
	var err error
	out.StartEpoch, err = strconv.ParseUint(r.FormValue("startEpoch"), 10, 64)
	if err != nil {
		return bad(fmt.Errorf("invalid startEpoch: %w", err))
	}
	if minipools {
		for _, raw := range strings.Split(r.FormValue("minipoolAddresses"), ",") {
			raw = strings.TrimSpace(raw)
			if !common.IsHexAddress(raw) {
				return bad(fmt.Errorf("invalid minipool address %q", raw))
			}
			out.MinipoolAddresses = append(out.MinipoolAddresses, common.HexToAddress(raw))
		}
	} else {
		raw := strings.TrimSpace(r.FormValue("megapoolAddress"))
		if raw != "" {
			if !common.IsHexAddress(raw) {
				return bad(fmt.Errorf("invalid megapool address"))
			}
			out.MegapoolAddress = common.HexToAddress(raw)
		}
		raw = r.FormValue("validatorIds")
		if raw == "" {
			raw = r.FormValue("validatorId")
		}
		for _, part := range strings.Split(raw, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 32)
			if err != nil {
				return bad(fmt.Errorf("invalid validator ID %q", part))
			}
			out.ValidatorIds = append(out.ValidatorIds, uint32(id))
		}
	}
	for _, part := range strings.Split(r.FormValue("participation"), ",") {
		word, ok := new(big.Int).SetString(strings.TrimSpace(part), 10)
		if !ok || word.Sign() < 0 || word.BitLen() > 256 {
			return bad(fmt.Errorf("participation entries must be uint256 values"))
		}
		out.Participation = append(out.Participation, word)
	}
	if err := out.Validate(); err != nil {
		return bad(err)
	}
	return out, nil
}

func CanChallengeHandler(minipools bool) func(snroute.Context) {
	return func(ctx snroute.Context) {
		request, err := parseRequest(ctx.Request, minipools)
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
		result, _, err := services.PreparePerformanceChallenge(ctx.Command(), request, opts)
		response.WriteResponse(ctx.Writer, result, err)
	}
}

func ChallengeHandler(minipools bool) func(snroute.WriteContext) {
	return func(ctx snroute.WriteContext) {
		request, err := parseRequest(ctx.Request, minipools)
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		maxBond, ok := new(big.Int).SetString(ctx.Request.FormValue("maxBond"), 10)
		if !ok || maxBond.Sign() < 0 || maxBond.BitLen() > 256 {
			response.WriteErrorResponse(ctx.Writer, &response.BadRequestError{Err: fmt.Errorf("maxBond must be the approved maximum bond in RPL wei")})
			return
		}
		tx, err := ctx.Transactor()
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		opts := tx.Opts()
		result, prepared, err := services.PreparePerformanceChallenge(ctx.Command(), request, opts)
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		if !result.CanChallenge {
			response.WriteErrorResponse(ctx.Writer, fmt.Errorf("%s", result.Reason))
			return
		}
		if result.ChallengeBond.Cmp(maxBond) > 0 {
			response.WriteErrorResponse(ctx.Writer, fmt.Errorf("challenge bond exceeds the approved maximum; estimate and confirm again"))
			return
		}
		rp, err := services.GetRocketPool(ctx.Command())
		if err != nil {
			response.WriteErrorResponse(ctx.Writer, err)
			return
		}
		hash, err := services.SubmitPerformanceChallenge(rp, prepared, opts)
		response.WriteResponse(ctx.Writer, &api.ChallengeMegapoolPerformanceResponse{TxHash: hash}, err)
	}
}
