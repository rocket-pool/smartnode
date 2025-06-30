package watchtower

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
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
	if !state.IsSaturnDeployed {
		return nil
	}

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

	_, err := t.bc.GetAllValidators()
	if err != nil {
		return fmt.Errorf("error fetching validators from bc: %w", err)
	}

	// Calculate the current epoch based on state.BeaconSlotNumber

	// state.BeaconSlotNumber

	// batchSize := 20                  // TODO: Fetch from contract
	// notifyThresholdInSeconds := 1000 // TODO: Fetch from contract

	// for _, validator := range state.MegapoolValidatorGlobalIndex {
	// 	if validator.ValidatorInfo.Staked && !validator.ValidatorInfo.Exited && !validator.ValidatorInfo.Exiting && !validator.ValidatorInfo.Locked {
	// 		validatorFromState := bcValidators[validator.ValidatorInfo.ValidatorIndex]

	// 		if validatorFromState.WithdrawableEpoch {
	// 			t.log.Printlnf("Validator %d has an invalid credential %s while the expected is %s. Dissolving...", validator.ValidatorInfo.ValidatorIndex, validatorFromState.WithdrawalCredentials, expectedWithdrawalAddress.Bytes())
	// 			t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
	// 		}

	// 	}
	// }

	return nil
}
