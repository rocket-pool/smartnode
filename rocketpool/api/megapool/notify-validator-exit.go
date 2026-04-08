package megapool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

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

	// Check validator status
	if !validatorInfo.Staked {
		response.InvalidStatus = true
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	proof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		return nil, err
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	gasInfo, err := megapool.EstimateNotifyExitGas(rp, megapoolAddress, validatorId, slotTimestamp, proof, slotProof, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	response.CanExit = !response.InvalidStatus
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

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimetamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), nil)
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
