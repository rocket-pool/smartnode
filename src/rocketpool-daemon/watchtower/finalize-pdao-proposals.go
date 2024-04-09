package watchtower

import (
	"fmt"
	"log/slog"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Finalize PDAO proposals task
type FinalizePdaoProposals struct {
	sp     *services.ServiceProvider
	logger *slog.Logger
	cfg    *config.SmartNodeConfig
	w      *wallet.Wallet
	ec     eth.IExecutionClient
	rp     *rocketpool.RocketPool
}

// Create finalize PDAO proposals task task
func NewFinalizePdaoProposals(sp *services.ServiceProvider, logger *log.Logger) *FinalizePdaoProposals {
	return &FinalizePdaoProposals{
		sp:     sp,
		cfg:    sp.GetConfig(),
		w:      sp.GetWallet(),
		ec:     sp.GetEthClient(),
		rp:     sp.GetRocketPool(),
		logger: logger.With(slog.String(keys.RoutineKey, "Finalize PDAO Proposals")),
	}
}

// Dissolve timed out minipools
func (t *FinalizePdaoProposals) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Checking for vetoable proposals to finalize...")

	// Get timed out minipools
	propIDs := t.getFinalizableProposals(state)
	if len(propIDs) == 0 {
		return nil
	}

	// Log
	t.logger.Info("Detected finalizable proposals.", slog.Int(keys.CountKey, len(propIDs)))

	// Finalize proposals
	for _, propID := range propIDs {
		if err := t.finalizeProposal(propID); err != nil {
			t.logger.Error("Error finalizing proposal", slog.Uint64(keys.ProposalKey, propID), log.Err(err))
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
	propLogger := t.logger.With(slog.Uint64(keys.ProposalKey, propID))
	propLogger.Info("Finalizing proposal...")

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
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, propLogger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, propLogger, txInfo, opts)
	if err != nil {
		return err
	}
	// Log
	propLogger.Info("Successfully finalized proposal.")

	// Return
	return nil

}
