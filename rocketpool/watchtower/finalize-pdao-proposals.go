package watchtower

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Finalize PDAO proposals task
type finalizePdaoProposals struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
}

// Create finalize PDAO proposals task task
func newFinalizePdaoProposals(c *cli.Context, logger log.ColorLogger) (*finalizePdaoProposals, error) {

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
	return &finalizePdaoProposals{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
	}, nil

}

// Dissolve timed out minipools
func (t *finalizePdaoProposals) run(state *state.NetworkState) error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	// Log
	t.log.Println("Checking for vetoable proposals to finalize...")

	// Get timed out minipools
	propIDs := t.getFinalizableProposals(state)
	if len(propIDs) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d proposal(s) have been vetoed and will be finalized...", len(propIDs))

	// Finalize proposals
	for _, propID := range propIDs {
		if err := t.finalizeProposal(propID); err != nil {
			t.log.Println(fmt.Errorf("Could not finalize proposal %d: %w", propID, err))
		}
	}

	// Return
	return nil
}

// Get timed out minipools
func (t *finalizePdaoProposals) getFinalizableProposals(state *state.NetworkState) []uint64 {
	finalizableProps := []uint64{}
	for _, prop := range state.ProtocolDaoProposalDetails {
		if prop.State == types.ProtocolDaoProposalState_Vetoed && !prop.IsFinalized {
			finalizableProps = append(finalizableProps, prop.ID)
		}
	}
	return finalizableProps
}

// Dissolve a minipool
func (t *finalizePdaoProposals) finalizeProposal(propID uint64) error {

	// Log
	t.log.Printlnf("Finalizing proposal %d...", propID)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := protocol.EstimateFinalizeGas(t.rp, propID, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to finalize the proposal: %w", err)
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
	hash, err := protocol.Finalize(t.rp, propID, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully finalized proposal %d.", propID)

	// Return
	return nil

}
