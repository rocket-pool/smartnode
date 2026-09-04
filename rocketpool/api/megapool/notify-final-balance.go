package megapool

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canNotifyFinalBalance(c *cli.Command, validatorId uint32, withdrawalSlot uint64) (*api.CanNotifyFinalBalanceResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
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

	// Response
	response := api.CanNotifyFinalBalanceResponse{}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}
	validatorInfo, err := mp.GetValidatorInfoAndPubkey(validatorId, nil)
	if err != nil {
		return nil, err
	}

	validatorStatus, err := bc.GetValidatorStatus(types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator status from beacon chain: %w", err)
	}
	validatorIndex, err := strconv.ParseUint(validatorStatus.Index, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the validator index")
	}
	// If the slot was not provided, use the validator's withdrawable epoch supplied by the beacon client
	if withdrawalSlot == 0 {
		withdrawalSlot = validatorStatus.WithdrawableEpoch * 32
	}

	proofVersion, proofData, slotTimestamp, err := services.GetFinalBalanceProofBundle(c, withdrawalSlot, validatorIndex, types.ValidatorPubkey(validatorInfo.Pubkey), megapoolAddress, w)
	if err != nil {
		return nil, err
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	gasLimits, err := megapool.EstimateNotifyFinalBalance(rp, megapoolAddress, validatorId, slotTimestamp, proofVersion, proofData, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasLimits = gasLimits
	response.CanExit = !response.InvalidStatus
	return &response, nil

}

func notifyFinalBalance(c *cli.Command, validatorId uint32, withdrawalSlot uint64, t *snroute.TransactOpts) (*api.NotifyValidatorExitResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
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

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NotifyValidatorExitResponse{}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the validator pubkey
	validatorInfo, err := mp.GetValidatorInfoAndPubkey(validatorId, nil)
	if err != nil {
		return nil, err
	}

	validatorStatus, err := bc.GetValidatorStatus(types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator status from beacon chain: %w", err)
	}
	validatorIndex, err := strconv.ParseUint(validatorStatus.Index, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the validator index")
	}
	// If the slot was not provided, use the validator's withdrawable epoch supplied by the beacon client
	if withdrawalSlot == 0 {
		withdrawalSlot = validatorStatus.WithdrawableEpoch * 32
	}

	proofVersion, proofData, slotTimestamp, err := services.GetFinalBalanceProofBundle(c, withdrawalSlot, validatorIndex, types.ValidatorPubkey(validatorInfo.Pubkey), megapoolAddress, w)
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	tx, err := megapool.NotifyFinalBalance(rp, megapoolAddress, validatorId, slotTimestamp, proofVersion, proofData, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

func canNotifyFinalBalanceHandler(ctx snroute.Context) {
	validatorId, err := parseUint32(ctx.Request, "validatorId")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	slot, err := parseUint64(ctx.Request, "slot")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canNotifyFinalBalance(ctx.Command(), validatorId, slot)
	response.WriteResponse(ctx.Writer, resp, err)
}

func notifyFinalBalanceHandler(ctx snroute.WriteContext) {
	validatorId, err := parseUint32(ctx.Request, "validatorId")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	slot, err := parseUint64(ctx.Request, "slot")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := notifyFinalBalance(ctx.Command(), validatorId, slot, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
