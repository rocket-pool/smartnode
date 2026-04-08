package megapool

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canStake(c *cli.Command, validatorId uint64) (*api.CanStakeResponse, error) {

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

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanStakeResponse{}

	// Check if the megapool is deployed
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !megapoolDeployed {
		response.CanStake = false
		return &response, nil
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

	// Get the validator count
	validatorCount, err := mp.GetValidatorCount(nil)
	if err != nil {
		return nil, err
	}

	if validatorId >= uint64(validatorCount) {
		response.CanStake = false
		return &response, nil
	}

	validatorInfo, err := mp.GetValidatorInfoAndPubkey(uint32(validatorId), nil)
	if err != nil {
		return nil, err
	}

	if !validatorInfo.InPrestake {
		response.CanStake = false
		return &response, nil
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		if strings.Contains(err.Error(), "index not found") {
			response.CanStake = false
			response.IndexNotFound = true
			return &response, nil
		}
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := megapool.EstimateStakeGas(rp, megapoolAddress, uint32(validatorId), slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	response.CanStake = true

	return &response, nil

}

func stake(c *cli.Command, validatorId uint64, opts *bind.TransactOpts) (*api.StakeResponse, error) {

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

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.StakeResponse{}

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

	validatorInfo, err := mp.GetValidatorInfoAndPubkey(uint32(validatorId), nil)
	if err != nil {
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		return nil, err
	}

	// Stake
	tx, err := megapool.Stake(rp, megapoolAddress, uint32(validatorId), slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
