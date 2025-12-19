package watchtower

import (
	"fmt"
	"time"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
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

// Dissolve timed out megapool validators task
type dissolveTimedOutMegapoolValidators struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
}

// Create dissolve timed out megapool validators task
func newDissolveTimedOutMegapoolValidators(c *cli.Context, logger log.ColorLogger) (*dissolveTimedOutMegapoolValidators, error) {

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

	// Return task
	return &dissolveTimedOutMegapoolValidators{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
	}, nil

}

// Dissolve timed out megapool validators
func (t *dissolveTimedOutMegapoolValidators) run(state *state.NetworkState) error {
	if !state.IsSaturnDeployed {
		return nil
	}

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	// Log
	t.log.Println("Checking for timed out megapool validators to dissolve...")

	// Dissolve validators
	err := t.dissolveMegapoolValidators(state)
	if err != nil {
		return err
	}

	return nil
}

// Get megapool validators that can be dissolved
func (t *dissolveTimedOutMegapoolValidators) dissolveMegapoolValidators(state *state.NetworkState) error {
	timeBeforeDissolve, err := protocol.GetMegapoolTimeBeforeDissolve(t.rp, nil)
	if err != nil {
		return err
	}

	for _, validator := range state.MegapoolValidatorGlobalIndex {
		if validator.ValidatorInfo.InPrestake {
			assignTime := time.Unix(int64(validator.ValidatorInfo.LastAssignmentTime), 0)
			if time.Since(assignTime) >= time.Duration(timeBeforeDissolve)*time.Second {
				// dissolve
				t.dissolveMegapoolValidator(validator)
			}

		}
	}
	return nil
}

func (t *dissolveTimedOutMegapoolValidators) dissolveMegapoolValidator(validator megapool.ValidatorInfoFromGlobalIndex) error {
	// Log
	t.log.Printlnf("Dissolving megapool validator ID: %d from megapool %s...", validator.ValidatorId, validator.MegapoolAddress)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Load the megapool contract
	mp, err := megapool.NewMegaPoolV1(t.rp, validator.MegapoolAddress, nil)
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := mp.EstimateDissolveValidatorGas(validator.ValidatorId, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to dissolve the megapool validator: %w", err)
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
	hash, err := mp.DissolveValidator(validator.ValidatorId, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully dissolved megapool validator ID: %s from megapool %s.", validator.ValidatorId, validator.MegapoolAddress)

	// Return
	return nil
}
