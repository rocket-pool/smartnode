package megapool

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
)

var errExitNotFinalized = errors.New("validator exit not yet finalized")

func ensureExitFinalized(bc beacon.Client, pubkey types.ValidatorPubkey) (eth2.BeaconState, error) {
	beaconState, err := services.GetBeaconState(bc)
	if err != nil {
		return nil, err
	}
	validators := beaconState.GetValidators()

	validatorIndexStr, err := bc.GetValidatorIndex(pubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon index: %w", err)
	}
	validatorIndex, err := strconv.ParseUint(validatorIndexStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing beacon index %q: %w", validatorIndexStr, err)
	}
	if validatorIndex >= uint64(len(validators)) {
		return nil, fmt.Errorf("%w: validator (beacon index %d) is not yet included in the finalized beacon state", errExitNotFinalized, validatorIndex)
	}
	if validators[validatorIndex].WithdrawableEpoch >= farFutureEpoch {
		return nil, fmt.Errorf("%w: validator (beacon index %d) withdrawable_epoch is still FAR_FUTURE on finalized state", errExitNotFinalized, validatorIndex)
	}
	return beaconState, nil
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

	// Proofs use finalized state — do not build/submit until the exit is there.
	beaconState, err := ensureExitFinalized(bc, pubkey)
	if err != nil {
		if errors.Is(err, errExitNotFinalized) {
			response.ExitNotFinalized = true
			response.CanExit = false
			return &response, nil
		}
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	proof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, pubkey, beaconState)
	if err != nil {
		return nil, err
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	gasLimits, err := megapool.EstimateNotifyExitGas(rp, megapoolAddress, validatorId, slotTimestamp, proof, slotProof, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasLimits = gasLimits
	response.CanExit = true
	return &response, nil

}

func notifyValidatorExit(c *cli.Command, validatorId uint32, opts *bind.TransactOpts) (*api.NotifyValidatorExitResponse, error) {

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

	beaconState, err := ensureExitFinalized(bc, pubkey)
	if err != nil {
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimetamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, pubkey, beaconState)
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	tx, err := megapool.NotifyExit(rp, megapoolAddress, validatorId, slotTimetamp, validatorProof, slotProof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
