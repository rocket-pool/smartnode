package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canNotifyFinalBalance(c *cli.Context, validatorId uint32, slot uint64) (*api.CanNotifyFinalBalanceResponse, error) {

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

	validatorInfo, err := mp.GetValidatorInfo(validatorId, nil)
	if err != nil {
		return nil, err
	}

	// Check validator status
	if !validatorInfo.Exiting {
		response.InvalidStatus = true
	}

	proof, err := services.GetWithdrawalProofForSlot(c, slot, validatorInfo.ValidatorIndex)
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err)
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	withdrawal := megapool.Withdrawal{
		Index:                 proof.WithdrawalIndex,
		ValidatorIndex:        validatorInfo.ValidatorIndex,
		WithdrawalCredentials: proof.WithdrawalAddress,
		AmountInGwei:          proof.Amount.Uint64(),
	}

	// Notify the validator exit
	gasInfo, err := mp.EstimateNotifyFinalBalance(validatorId, proof.WithdrawalSlot, big.NewInt(int64(proof.WithdrawalIndex)), withdrawal, proof.Slot, proof.Witnesses, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	response.CanExit = !response.InvalidStatus
	return &response, nil

}

func notifyFinalBalance(c *cli.Context, validatorId uint32, slot uint64) (*api.NotifyValidatorExitResponse, error) {

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
	validatorInfo, err := mp.GetValidatorInfo(validatorId, nil)
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	proof, err := services.GetWithdrawalProofForSlot(c, slot, validatorInfo.ValidatorIndex)
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err)
	}

	withdrawal := megapool.Withdrawal{
		Index:                 proof.WithdrawalIndex,
		ValidatorIndex:        validatorInfo.ValidatorIndex,
		WithdrawalCredentials: proof.WithdrawalAddress,
		AmountInGwei:          proof.Amount.Uint64(),
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Notify the validator exit
	hash, err := mp.NotifyFinalBalance(validatorId, proof.WithdrawalSlot, big.NewInt(int64(proof.WithdrawalIndex)), withdrawal, slot, proof.Witnesses, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
