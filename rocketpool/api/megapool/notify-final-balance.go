
package megapool

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Get the network
	network := cfg.Smartnode.Network.Value.(cfgtypes.Network)

	validatorStatus, err := bc.GetValidatorStatus(types.ValidatorPubkey(validatorInfo.Pubkey), nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator status from beacon chain: %w", err)
	}
	validatorIndex, err := strconv.ParseUint(validatorStatus.Index, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the validator index")
	}
	// If the slot was not provided, use the validator's withdrawable epoch supplied by the beacon client
	if slot == 0 {
		slot = validatorStatus.WithdrawableEpoch * 32
	}

	withdrawalProof, slotUsed, stateUsed, err := services.GetWithdrawalProofForSlot(c, slot, validatorIndex)
	if err != nil {
		fmt.Printf("An error occurred while getting the withdrawal proof: %s\n", err)
		// try to fetch the withdrawal proof from the Rocket Pool API
		withdrawalProof, slotUsed, err = services.GetWithdrawalProofForSlotFromAPI(c, slot, validatorIndex, network)
		if err != nil {
			fmt.Printf("An error occurred while getting the withdrawal proof from the Rocket Pool API: %s\n", err)
			return nil, err
		}
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	withdrawal := megapool.Withdrawal{
		Index:                 withdrawalProof.WithdrawalIndex,
		ValidatorIndex:        validatorIndex,
		WithdrawalCredentials: withdrawalProof.WithdrawalAddress,
		AmountInGwei:          withdrawalProof.Amount.Uint64(),
	}

	finalBalanceProof := megapool.WithdrawalProof{
		WithdrawalSlot: withdrawalProof.WithdrawalSlot,
		WithdrawalNum:  uint16(withdrawalProof.IndexInWithdrawalsArray),
		Withdrawal:     withdrawal,
		Witnesses:      withdrawalProof.Witnesses,
	}

	// Get the eth2Config and validatorProof
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, slotUsed, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), stateUsed)
	if err != nil {
		return nil, err
	}

	// Notify the validator exit
	gasInfo, err := megapool.EstimateNotifyFinalBalance(rp, megapoolAddress, validatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
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

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
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
	if slot == 0 {
		slot = validatorStatus.WithdrawableEpoch * 32
	}

	withdrawalProof, proofSlot, stateUsed, err := services.GetWithdrawalProofForSlot(c, slot, validatorIndex)
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err)
	}

	withdrawal := megapool.Withdrawal{
		Index:                 withdrawalProof.WithdrawalIndex,
		ValidatorIndex:        validatorIndex,
		WithdrawalCredentials: withdrawalProof.WithdrawalAddress,
		AmountInGwei:          withdrawalProof.Amount.Uint64(),
	}

	finalBalanceProof := megapool.WithdrawalProof{
		WithdrawalSlot: withdrawalProof.WithdrawalSlot,
		WithdrawalNum:  uint16(withdrawalProof.IndexInWithdrawalsArray),
		Withdrawal:     withdrawal,
		Witnesses:      withdrawalProof.Witnesses,
	}

	// Get the eth2Config and validatorProof
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}
	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(c, proofSlot, w, eth2Config, megapoolAddress, types.ValidatorPubkey(validatorInfo.Pubkey), stateUsed)
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Notify the validator exit
	tx, err := megapool.NotifyFinalBalance(rp, megapoolAddress, validatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
