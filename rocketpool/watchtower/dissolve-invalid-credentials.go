package watchtower

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
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

// Dissolve timed out minipools task
type dissolveInvalidCredentials struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
	bc  *services.BeaconClientManager
}

const FarFutureEpoch uint64 = 0xffffffffffffffff

// Create dissolve timed out megapool validators task
func newDissolveInvalidCredentials(c *cli.Context, logger log.ColorLogger) (*dissolveInvalidCredentials, error) {

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
	return &dissolveInvalidCredentials{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
		bc:  bc,
	}, nil

}

// Dissolve timed out megapool validators
func (t *dissolveInvalidCredentials) run(state *state.NetworkState) error {
	if !state.IsSaturnDeployed {
		return nil
	}

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	// Log
	t.log.Println("Checking for invalid info on megapool validators to dissolve...")

	// Dissolve validators
	err := t.dissolveInvalidCredentialValidators(state)
	if err != nil {
		return err
	}

	return nil
}

// Get megapool validators that can be dissolved due to using invalid credentials
func (t *dissolveInvalidCredentials) dissolveInvalidCredentialValidators(state *state.NetworkState) error {

	for _, validator := range state.MegapoolValidatorGlobalIndex {
		if validator.ValidatorInfo.InPrestake {
			expectedWithdrawalAddress := services.CalculateMegapoolWithdrawalCredentials(validator.MegapoolAddress)
			// Fetch the validator from the beacon state to compare credentials
			validatorFromState, err := t.bc.GetValidatorStatus(types.ValidatorPubkey(validator.Pubkey), nil)
			if err != nil {
				t.log.Printlnf("Error fetching validator %s from beacon state: %s", validatorFromState.Index, err)
				continue
			}
			if validatorFromState.Index != "" && !bytes.Equal(validatorFromState.WithdrawalCredentials.Bytes(), expectedWithdrawalAddress.Bytes()) {
				t.log.Printlnf("Validator %s has an invalid credential %s while the expected is %s. Dissolving...", validatorFromState.Index, validatorFromState.WithdrawalCredentials, expectedWithdrawalAddress.Bytes())
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Withdrawable epoch should be FAR_FUTURE_EPOCH
			if validatorFromState.WithdrawableEpoch != FarFutureEpoch {
				t.log.Printlnf("Validator %s has a withdrawable epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.WithdrawableEpoch, FarFutureEpoch)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Exit epoch should be FAR_FUTURE_EPOCH
			if validatorFromState.ExitEpoch != FarFutureEpoch {
				t.log.Printlnf("Validator %s has an exit epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ExitEpoch, FarFutureEpoch)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Slashed should be false
			if validatorFromState.Slashed {
				t.log.Printlnf("Validator %s is slashed while the expected is false. Dissolving...", validatorFromState.Index)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Effective balance should be less than 32 ETH
			if validatorFromState.EffectiveBalance >= 32000000000 {
				t.log.Printlnf("Validator %s has an effective balance of %d while the expected is less than 32 ETH. Dissolving...", validatorFromState.Index, validatorFromState.EffectiveBalance)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Activation eligibility epoch should be FAR_FUTURE_EPOCH
			if validatorFromState.ActivationEligibilityEpoch != FarFutureEpoch {
				t.log.Printlnf("Validator %s has an activation eligibility epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ActivationEligibilityEpoch, FarFutureEpoch)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
			// Activation epoch should be FAR_FUTURE_EPOCH
			if validatorFromState.ActivationEpoch != FarFutureEpoch {
				t.log.Printlnf("Validator %s has an activation epoch of %d while the expected is %d. Dissolving...", validatorFromState.Index, validatorFromState.ActivationEpoch, FarFutureEpoch)
				t.dissolveMegapoolValidator(validator, expectedWithdrawalAddress)
			}
		}

	}
	return nil
}

func (t *dissolveInvalidCredentials) dissolveMegapoolValidator(validator megapool.ValidatorInfoFromGlobalIndex, expectedWithdrawalCredentials common.Hash) error {
	// Log
	t.log.Printlnf("Dissolving megapool validator ID: %d from megapool %s...", validator.ValidatorId, validator.MegapoolAddress)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		return err
	}

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, 0, t.w, eth2Config, validator.MegapoolAddress, types.ValidatorPubkey(validator.Pubkey), nil)
	if err != nil {
		return fmt.Errorf("error getting validator proof: %w", err)
	}

	// Get the gas limit
	gasInfo, err := megapool.EstimateDissolveWithProof(t.rp, validator.MegapoolAddress, validator.ValidatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to dissolve the validator: %w", err)
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

	// Dissolve
	tx, err := megapool.DissolveWithProof(t.rp, validator.MegapoolAddress, validator.ValidatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully dissolved megapool validator ID: %s from megapool %s. (Invalid credentials)", validator.ValidatorId, validator.MegapoolAddress)

	// Return
	return nil
}
