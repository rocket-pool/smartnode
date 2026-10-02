package megapool

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
)

var errExitNotReady = errors.New("validator exit not yet visible in the proof state")

func ensureExitReady(bc beacon.Client, ec rocketpool.ExecutionClient, pubkey types.ValidatorPubkey) (eth2.BeaconState, uint64, error) {
	beaconState, slotTimestamp, err := services.GetHeadBeaconState(bc, ec)
	if err != nil {
		return nil, 0, err
	}
	if err := checkExitProofState(bc, pubkey, beaconState); err != nil {
		return nil, 0, err
	}
	return beaconState, slotTimestamp, nil
}

func checkExitProofState(bc beacon.Client, pubkey types.ValidatorPubkey, beaconState eth2.BeaconState) error {
	validators := beaconState.GetValidators()

	validatorIndexStr, err := bc.GetValidatorIndex(pubkey)
	if err != nil {
		return fmt.Errorf("error getting beacon index: %w", err)
	}
	validatorIndex, err := strconv.ParseUint(validatorIndexStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing beacon index %q: %w", validatorIndexStr, err)
	}
	if validatorIndex >= uint64(len(validators)) {
		return fmt.Errorf("%w: validator (beacon index %d) is not yet included in the selected beacon state", errExitNotReady, validatorIndex)
	}
	if validators[validatorIndex].WithdrawableEpoch >= farFutureEpoch {
		return fmt.Errorf("%w: validator (beacon index %d) withdrawable_epoch is still FAR_FUTURE in the selected beacon state", errExitNotReady, validatorIndex)
	}
	return nil
}

func canNotifyValidatorExit(c *cli.Command, validatorId uint32) (*api.CanNotifyValidatorExitResponse, error) {

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
	response := api.CanNotifyValidatorExitResponse{}

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

	if !validatorInfo.Staked {
		response.InvalidStatus = true
		response.CanExit = false
		return &response, nil
	}
	if validatorInfo.Exited {
		response.AlreadyExited = true
		response.CanExit = false
		return &response, nil
	}
	if validatorInfo.Exiting {
		response.AlreadyExiting = true
		response.CanExit = false
		return &response, nil
	}

	pubkey := types.ValidatorPubkey(validatorInfo.Pubkey)

	// Check the exit is visible in the state selected for the proof.
	beaconState, slotTimestamp, err := ensureExitReady(bc, rp.Client, pubkey)
	if err != nil {
		if errors.Is(err, errExitNotReady) {
			response.ExitNotReady = true
			response.CanExit = false
			return &response, nil
		}
		return nil, err
	}

	proof, slotProof, err := services.GetValidatorProofFromState(bc, pubkey, beaconState)
	if err != nil {
		return nil, err
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	gasLimits, err := services.EstimateMegapoolNotifyExitGas(rp, megapoolAddress, validatorId, slotTimestamp, proof, slotProof, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasLimits = gasLimits
	response.CanExit = true
	return &response, nil

}

func notifyValidatorExit(c *cli.Command, validatorId uint32, t *snroute.TransactOpts) (*api.NotifyValidatorExitResponse, error) {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
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

	pubkey := types.ValidatorPubkey(validatorInfo.Pubkey)

	beaconState, slotTimestamp, err := ensureExitReady(bc, rp.Client, pubkey)
	if err != nil {
		return nil, err
	}

	validatorProof, slotProof, err := services.GetValidatorProofFromState(bc, pubkey, beaconState)
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	tx, err := services.NotifyMegapoolExit(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}

func canNotifyValidatorExitHandler(ctx snroute.Context) {
	validatorId, err := parseUint32(ctx.Request, "validatorId")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canNotifyValidatorExit(ctx.Command(), validatorId)
	response.WriteResponse(ctx.Writer, resp, err)
}

func notifyValidatorExitHandler(ctx snroute.WriteContext) {
	validatorId, err := parseUint32(ctx.Request, "validatorId")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := notifyValidatorExit(ctx.Command(), validatorId, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
