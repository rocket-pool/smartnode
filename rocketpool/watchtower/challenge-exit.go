package watchtower

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/flashbots"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

type challengeValidatorsExiting struct {
	c   *cli.Command
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
	bc  *services.BeaconClientManager
}

func newChallengeValidatorsExiting(c *cli.Command, logger log.ColorLogger) (*challengeValidatorsExiting, error) {
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
func (t *challengeValidatorsExiting) run(state *state.NetworkStateIndex) error {
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

// Get megapool validators that can be challenged for exiting without a notification
func (t *challengeValidatorsExiting) challengeValidatorsExiting(state *state.NetworkStateIndex) error {

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
	for _, validator := range state.MegapoolValidators {
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
		gasLimits, err := megapool.EstimateChallengeExitGas(t.rp, exitChallenges, opts)
		if err != nil {
			return fmt.Errorf("error calling estimate challenge exit: %w", err)
		}

		// Print the gas info
		maxFee := math.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
		if !gasLimits.PrintAndCheck(false, 0, &t.log, maxFee, 0) {
			return nil
		}

		// Set the gas settings
		opts.GasFeeCap = maxFee
		opts.GasTipCap = math.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
		opts.GasLimit = gasLimits.Safe

		// Prefer a private bundle, with public fallback only when no relay is available.
		tx, bundled, err := t.submitExitChallenges(exitChallenges, opts)
		if err != nil {
			return err
		}
		if bundled {
			t.log.Printlnf("Exit challenge bundle included: %s", tx.Hash())
		} else if err := transactions.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.ec, &t.log); err != nil {
			return err
		}

		t.log.Printlnf("Challenged %d validators exiting", batched)

	}

	return nil
}

// submitExitChallenges signs locally and submits a one-transaction bundle.
// The second result is true once the bundle has been included. A missing relay
// or DNS failure during submission permits a local public fallback.
func (t *challengeValidatorsExiting) submitExitChallenges(exitChallenges []megapool.ExitChallenge, opts *bind.TransactOpts) (*gethtypes.Transaction, bool, error) {
	if opts.NoSend {
		return nil, false, fmt.Errorf("cannot submit exit challenges with transaction submission disabled")
	}
	parent := opts.Context
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, flashbots.DefaultSubmissionTimeout)
	defer cancel()

	// Build and sign without broadcasting through the contract client.
	privateOpts := *opts
	privateOpts.NoSend = true
	privateOpts.Context = ctx
	tx, err := megapool.ChallengeExit(t.rp, exitChallenges, &privateOpts)
	if err != nil {
		return nil, false, err
	}
	submitLocally := func(reason string) (*gethtypes.Transaction, bool, error) {
		t.log.Printlnf("%s; submitting the exit challenge through the local execution client", reason)
		if err := t.ec.SendTransaction(parent, tx); err != nil {
			return nil, false, err
		}
		return tx, false, nil
	}

	relayURL := t.cfg.Smartnode.GetFlashbotsRelayUrl()
	if relayURL == "" {
		return submitLocally(fmt.Sprintf("flashbotsRelayUrl is not configured for network %s", t.cfg.GetNetwork()))
	}

	t.log.Printlnf("Submitting exit challenge transaction %s as a private bundle...", tx.Hash())
	included, err := flashbots.SubmitBundleAndWait(ctx, nil, t.ec, relayURL, []*gethtypes.Transaction{tx}, flashbots.DefaultBundleBlockCount)
	if err != nil {
		if errors.Is(err, flashbots.ErrRelayUnavailable) {
			return submitLocally(flashbots.ErrRelayUnavailable.Error())
		}
		return nil, false, fmt.Errorf("error submitting exit challenge bundle: %w", err)
	}
	if !included {
		return nil, false, fmt.Errorf("exit challenge bundle for transaction %s was not included", tx.Hash())
	}

	// Bundle inclusion checks receipt availability; also require successful execution.
	receipt, err := t.ec.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, false, fmt.Errorf("error checking exit challenge receipt: %w", err)
	}
	if receipt == nil || receipt.Status != gethtypes.ReceiptStatusSuccessful {
		return nil, false, fmt.Errorf("exit challenge transaction %s failed", tx.Hash())
	}
	return tx, true, nil
}
