package node

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpsvc "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

// Manage fee recipient task
type manageFeeRecipient struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	d   *client.Client
	bc  beacon.Client
}

// Create manage fee recipient task
func newManageFeeRecipient(c *cli.Context, logger log.ColorLogger) (*manageFeeRecipient, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &manageFeeRecipient{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		rp:  rp,
		d:   d,
		bc:  bc,
	}, nil

}

// Manage fee recipient
func (m *manageFeeRecipient) run(state *state.NetworkState) error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(m.c, true); err != nil {
		return err
	}

	// Log
	m.log.Println("Checking for correct fee recipient...")

	// Get node account
	nodeAccount, err := m.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the fee recipient info for the node
	feeRecipientInfo, err := rputils.GetFeeRecipientInfo(m.rp, m.bc, nodeAccount.Address, state)
	if err != nil {
		return fmt.Errorf("error getting fee recipient info: %w", err)
	}

	// Get the correct fee recipient address
	var correctFeeRecipient common.Address
	if feeRecipientInfo.IsInSmoothingPool || feeRecipientInfo.IsInOptOutCooldown {
		correctFeeRecipient = feeRecipientInfo.SmoothingPoolAddress
	} else {
		correctFeeRecipient = feeRecipientInfo.FeeDistributorAddress
	}

	// Check if the VC is using the correct fee recipient
	fileExists, correctAddress, err := rpsvc.CheckFeeRecipientFile(correctFeeRecipient, m.cfg)
	if err != nil {
		return fmt.Errorf("error validating fee recipient files: %w", err)
	}

	if !fileExists {
		m.log.Println("Fee recipient files don't all exist, regenerating...")
	} else if !correctAddress {
		m.log.Printlnf("WARNING: Fee recipient files did not contain the correct fee recipient of %s, regenerating...", correctFeeRecipient.Hex())
	} else {
		// Files are all correct, return.
		return nil
	}

	// Regenerate the fee recipient files
	err = rpsvc.UpdateFeeRecipientFile(correctFeeRecipient, m.cfg)
	if err != nil {
		m.log.Println("***ERROR***")
		m.log.Printlnf("Error updating fee recipient files: %s", err.Error())
		m.log.Println("Shutting down the validator client for safety to prevent you from being penalized...")
		alerting.AlertFeeRecipientChanged(m.cfg, correctFeeRecipient, false)

		err = validator.StopValidator(m.cfg, m.bc, &m.log, m.d)
		if err != nil {
			return fmt.Errorf("error stopping validator client: %w", err)
		}
		return nil
	} else {
		alerting.AlertFeeRecipientChanged(m.cfg, correctFeeRecipient, true)
	}

	// Restart the VC
	m.log.Println("Fee recipient files updated successfully! Restarting validator client...")
	err = validator.RestartValidator(m.cfg, m.bc, &m.log, m.d)
	if err != nil {
		return fmt.Errorf("error restarting validator client: %w", err)
	}

	// Log & return
	m.log.Println("Successfully restarted, you are now validating safely.")
	return nil

}
