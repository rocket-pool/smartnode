package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canDissolveWithProof(c *cli.Context, validatorId uint32) (*api.CanDissolveWithProofResponse, error) {

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
	response := api.CanDissolveWithProofResponse{}

	// Check if the megapool is deployed
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !megapoolDeployed {
		response.CanDissolve = false
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

	validatorInfo, err := mp.GetValidatorInfoAndPubkey(validatorId, nil)
	if err != nil {
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	proof, slotTimestamp, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey))
	if err != nil {
		return nil, err
	}

	if !validatorInfo.InPrestake {
		response.CanDissolve = false
		response.NotInPrestake = true
		return &response, nil
	}

	// Check if the withdrawal credentials mismatch the expected ones
	expectedCredentials, err := mp.GetWithdrawalCredentials(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting the exptected withdrawal credeentials: %w", err)
	}
	if expectedCredentials.Cmp(common.Hash(proof.Validator.WithdrawalCredentials)) == 0 {
		response.CanDissolve = false
		response.ValidCredentials = true
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := megapool.EstimateDissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, proof, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	response.CanDissolve = true

	return &response, nil

}

func dissolveWithProof(c *cli.Context, validatorId uint32) (*api.DissolveWithProofResponse, error) {

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
	response := api.DissolveWithProofResponse{}

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

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimestamp, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey))
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Dissolve
	tx, err := megapool.DissolveWithProof(rp, megapoolAddress, validatorId, slotTimestamp, validatorProof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
