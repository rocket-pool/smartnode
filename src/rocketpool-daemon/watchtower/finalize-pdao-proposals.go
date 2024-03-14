package watchtower

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Finalize PDAO proposals task
type FinalizePdaoProposals struct {
	sp  *services.ServiceProvider
	log log.ColorLogger
	cfg *config.SmartNodeConfig
	w   *wallet.Wallet
	ec  eth.IExecutionClient
	rp  *rocketpool.RocketPool
}

// Create finalize PDAO proposals task task
func NewFinalizePdaoProposals(sp *services.ServiceProvider, logger log.ColorLogger) *FinalizePdaoProposals {
	return &FinalizePdaoProposals{
		sp:  sp,
		log: logger,
	}
}

// Dissolve timed out minipools
func (t *FinalizePdaoProposals) Run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for vetoable proposals to finalize...")

	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.ec = t.sp.GetEthClient()

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
func (t *FinalizePdaoProposals) getFinalizableProposals(state *state.NetworkState) []uint64 {
	finalizableProps := []uint64{}
	for _, prop := range state.ProtocolDaoProposalDetails {
		if prop.State.Formatted() == types.ProtocolDaoProposalState_Vetoed && !prop.IsFinalized.Get() {
			finalizableProps = append(finalizableProps, prop.ID)
		}
	}
	return finalizableProps
}

// Dissolve a minipool
func (t *FinalizePdaoProposals) finalizeProposal(propID uint64) error {
	// Log
	t.log.Printlnf("Finalizing proposal %d...", propID)

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Make the proposal
	prop, err := protocol.NewProtocolDaoProposal(t.rp, propID)
	if err != nil {
		return fmt.Errorf("error creating binding for proposal %d: %w", propID, err)
	}

	// Get the tx info
	txInfo, err := prop.Finalize(opts)
	if err != nil {
		return fmt.Errorf("error getting finalize tx for proposal %d: %w", propID, err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating finalize TX failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}
	// Log
	t.log.Printlnf("Successfully finalized proposal %d.", propID)

	// Return
	return nil

}
