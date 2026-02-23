package watchtower

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

type challengeValidatorsExiting struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
	bc  *services.BeaconClientManager
}

func newChallengeValidatorsExiting(c *cli.Context, logger log.ColorLogger) (*challengeValidatorsExiting, error) {
	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	// Get the beacon client
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &challengeValidatorsExiting{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
		bc:  bc,
	}, nil
}

// Flag validators exiting that didn't notify the exit
func (t *challengeValidatorsExiting) run(state *state.NetworkState) error {
	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	// Log
	t.log.Println("Challenging validators exiting without a notification...")

	// Dissolve validators
	err := t.challengeValidatorsExiting(state)
	if err != nil {
		return err
	}

	return nil
}

// Get megapool validators that can be dissolved due to using invalid credentials
func (t *challengeValidatorsExiting) challengeValidatorsExiting(state *state.NetworkState) error {

	// Calculate the current epoch based on state.BeaconSlotNumber
	currentSlot := state.BeaconSlotNumber
	currentEpoch := currentSlot / state.BeaconConfig.SlotsPerEpoch

	batchSize := 20 // TODO: Fetch from contract
	notifyThresholdInEpochs, err := protocol.GetNotifyThreshold(t.rp, nil)
	if err != nil {
		return fmt.Errorf("error getting notify threshold: %w", err)
	}

	challengeMegapoolAddressToIds := make(map[common.Address][]uint32)
	batched := 0
	for _, validator := range state.MegapoolValidatorGlobalIndex {
		if batched >= batchSize {
			t.log.Printlnf("Batched %d validators, exiting...", batched)
			break
		}
		if validator.ValidatorInfo.Staked && !validator.ValidatorInfo.Exited && !validator.ValidatorInfo.Exiting && !validator.ValidatorInfo.Locked {
			validatorFromState := state.MegapoolValidatorDetails[types.ValidatorPubkey(validator.Pubkey)]

			if validatorFromState.WithdrawableEpoch-notifyThresholdInEpochs <= currentEpoch {
				t.log.Printlnf("Validator %s has an withdrawable epoch %d which is past the notify threshold... Challenging", validatorFromState.Index, validatorFromState.WithdrawableEpoch)
				batched++
				challengeMegapoolAddressToIds[validator.MegapoolAddress] = append(challengeMegapoolAddressToIds[validator.MegapoolAddress], validator.ValidatorId)
			}

		}
	}
	if batched > 0 {
		t.log.Printlnf("Challenging %d validators exiting without a notification...", batched)

		// Get the transactor
		opts, err := t.w.GetNodeAccountTransactor()
		if err != nil {
			return fmt.Errorf("error getting transactor: %w", err)
		}

		exitChallenges := []megapool.ExitChallenge{}

		// Iterate over the megapools creating the challenge exit objects
		for megapoolAddress := range challengeMegapoolAddressToIds {

			// Create the challenge exit
			validatorIds := challengeMegapoolAddressToIds[megapoolAddress]
			exitChallenge := megapool.ExitChallenge{
				Megapool:     megapoolAddress,
				ValidatorIds: validatorIds,
			}
			exitChallenges = append(exitChallenges, exitChallenge)
		}

		// Challenge the validators
		gasInfo, err := megapool.EstimateChallengeExitGas(t.rp, exitChallenges, opts)
		if err != nil {
			return fmt.Errorf("error calling estimate challenge exit: %w", err)
		}

		// Print the gas info
		maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasInfo.SafeGasLimit

		// Challenge
		tx, err := megapool.ChallengeExit(t.rp, exitChallenges, opts)
		if err != nil {
			return err
		}

		// Print TX info and wait for it to be included in a block
		err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
		if err != nil {
			return err
		}

		t.log.Printlnf("Challenged %d validators exiting", batched)

	}

	return nil
}
