package megapool

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/storage"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func checkSaturn2Deployed(rp *rocketpool.RocketPool, opts *bind.CallOpts) error {
	saturn2Deployed, err := state.IsSaturn2Deployed(rp, opts)
	if err != nil {
		return fmt.Errorf("error checking if Saturn 2 is deployed: %w", err)
	}
	if !saturn2Deployed {
		return fmt.Errorf("deficit exits are not available until Saturn 2 is deployed")
	}
	return nil
}

// prepareDeficitExit is shared by estimation and submission so a direct POST
// cannot use the contract's owner exemption.
func prepareDeficitExit(rp *rocketpool.RocketPool, address common.Address, validatorIds []uint32, opts *bind.TransactOpts) error {
	if address == (common.Address{}) || len(validatorIds) == 0 {
		return fmt.Errorf("a megapool address and at least one validator ID are required")
	}
	mp, err := megapool.NewMegapool(rp, address, nil)
	if err != nil {
		return err
	}
	if mp.GetVersion() < 2 {
		return fmt.Errorf("megapool %s must upgrade to delegate version 2 or later to support forced exits", address.Hex())
	}
	nodeAddress, err := mp.GetNodeAddress(nil)
	if err != nil {
		return err
	}
	withdrawalAddress, err := storage.GetNodeWithdrawalAddress(rp, nodeAddress, nil)
	if err != nil {
		return err
	}
	if opts.From == nodeAddress || opts.From == withdrawalAddress {
		return fmt.Errorf("deficit exits must be submitted by a caller other than the target node or its withdrawal address; the contract bypasses the deficit check for these addresses")
	}
	fee, err := network.GetExitFee(rp, nil)
	if err != nil {
		return err
	}
	opts.Value = new(big.Int).Mul(fee, new(big.Int).SetUint64(uint64(len(validatorIds))))
	return nil
}

func canExitDeficit(c *cli.Command, address common.Address, validatorIds []uint32) (*api.CanExitMegapoolDeficitResponse, error) {
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	return estimateDeficitExit(rp, address, validatorIds, opts)
}

func estimateDeficitExit(rp *rocketpool.RocketPool, address common.Address, validatorIds []uint32, opts *bind.TransactOpts) (*api.CanExitMegapoolDeficitResponse, error) {
	result := &api.CanExitMegapoolDeficitResponse{ValidatorIds: validatorIds}
	if err := checkSaturn2Deployed(rp, nil); err != nil {
		return nil, err
	}
	if err := prepareDeficitExit(rp, address, validatorIds, opts); err != nil {
		return nil, err
	}
	result.ExitFee = opts.Value
	// The contract checks registration, validator eligibility, and the projected
	// deficit including existing exits and the requested list size.
	var err error
	result.GasLimits, err = network.EstimateExitMegapoolValidatorsGas(rp, address, validatorIds, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot exit the selected validators: %w", err)
	}
	result.CanExit = true
	return result, nil
}

func exitDeficit(c *cli.Command, address common.Address, validatorIds []uint32, maxExitFee *big.Int, t *snroute.TransactOpts) (*api.ExitMegapoolDeficitResponse, error) {
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	return submitDeficitExit(rp, address, validatorIds, maxExitFee, t.Opts())
}

func submitDeficitExit(rp *rocketpool.RocketPool, address common.Address, validatorIds []uint32, maxExitFee *big.Int, opts *bind.TransactOpts) (*api.ExitMegapoolDeficitResponse, error) {
	if err := checkSaturn2Deployed(rp, nil); err != nil {
		return nil, err
	}
	if err := prepareDeficitExit(rp, address, validatorIds, opts); err != nil {
		return nil, err
	}
	if maxExitFee == nil || maxExitFee.Sign() < 0 || opts.Value.Cmp(maxExitFee) > 0 {
		return nil, fmt.Errorf("current total exit fee exceeds the approved maximum; estimate and confirm the exit again")
	}
	hash, err := network.ExitMegapoolValidators(rp, address, validatorIds, opts)
	if err != nil {
		return nil, err
	}
	return &api.ExitMegapoolDeficitResponse{TxHash: hash}, nil
}

func parseDeficitMegapoolAddress(r *http.Request) (common.Address, error) {
	rawAddress := strings.TrimSpace(r.FormValue("megapoolAddress"))
	if !common.IsHexAddress(rawAddress) || common.HexToAddress(rawAddress) == (common.Address{}) {
		return common.Address{}, &response.BadRequestError{Err: fmt.Errorf("a valid nonzero megapoolAddress is required")}
	}
	return common.HexToAddress(rawAddress), nil
}

func parseDeficitExitParams(r *http.Request) (common.Address, []uint32, error) {
	address, err := parseDeficitMegapoolAddress(r)
	if err != nil {
		return common.Address{}, nil, err
	}
	parts := strings.Split(r.FormValue("validatorIds"), ",")
	ids := make([]uint32, 0, len(parts))
	seen := make(map[uint32]bool, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 32)
		if err != nil {
			return common.Address{}, nil, &response.BadRequestError{Err: fmt.Errorf("invalid validator ID %q: expected comma-separated uint32 IDs", part)}
		}
		if seen[uint32(id)] {
			return common.Address{}, nil, &response.BadRequestError{Err: fmt.Errorf("duplicate validator ID %d", id)}
		}
		seen[uint32(id)] = true
		ids = append(ids, uint32(id))
	}
	return address, ids, nil
}

func canExitDeficitHandler(ctx snroute.Context) {
	address, ids, err := parseDeficitExitParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	result, err := canExitDeficit(ctx.Command(), address, ids)
	response.WriteResponse(ctx.Writer, result, err)
}

func exitDeficitHandler(ctx snroute.WriteContext) {
	address, ids, err := parseDeficitExitParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	maxExitFee, ok := new(big.Int).SetString(ctx.Request.FormValue("maxExitFee"), 10)
	if !ok || maxExitFee.Sign() < 0 {
		response.WriteErrorResponse(ctx.Writer, &response.BadRequestError{Err: fmt.Errorf("maxExitFee must be a non-negative total exit fee in wei")})
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	result, err := exitDeficit(ctx.Command(), address, ids, maxExitFee, opts)
	response.WriteResponse(ctx.Writer, result, err)
}
