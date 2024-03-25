package watchtower

import (
	"fmt"
	"time"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Settings
const MinipoolStatusBatchSize = 20

// Dissolve timed out minipools task
type DissolveTimedOutMinipools struct {
	sp    *services.ServiceProvider
	log   log.ColorLogger
	cfg   *config.SmartNodeConfig
	w     *wallet.Wallet
	rp    *rocketpool.RocketPool
	ec    eth.IExecutionClient
	mpMgr *minipool.MinipoolManager
}

// Create dissolve timed out minipools task
func NewDissolveTimedOutMinipools(sp *services.ServiceProvider, logger log.ColorLogger) *DissolveTimedOutMinipools {
	return &DissolveTimedOutMinipools{
		sp:  sp,
		cfg: sp.GetConfig(),
		w:   sp.GetWallet(),
		rp:  sp.GetRocketPool(),
		ec:  sp.GetEthClient(),
		log: logger,
	}
}

// Dissolve timed out minipools
func (t *DissolveTimedOutMinipools) Run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for timed out minipools to dissolve...")

	// Update contract bindings
	var err error
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Get timed out minipools
	minipools, err := t.getTimedOutMinipools(state)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) have timed out and will be dissolved...", len(minipools))

	// Dissolve minipools
	for _, mp := range minipools {
		if err := t.dissolveMinipool(mp); err != nil {
			t.log.Println(fmt.Errorf("error dissolving minipool %s: %w", mp.Common().Address.Hex(), err))
		}
	}

	// Return
	return nil
}

// Get timed out minipools
func (t *DissolveTimedOutMinipools) getTimedOutMinipools(state *state.NetworkState) ([]minipool.IMinipool, error) {
	timedOutMinipools := []minipool.IMinipool{}
	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
	blockTime := genesisTime.Add(secondsSinceGenesis)

	// Filter minipools by status
	launchTimeoutBig := state.NetworkDetails.MinipoolLaunchTimeout
	launchTimeout := time.Duration(launchTimeoutBig.Uint64()) * time.Second
	for _, mpd := range state.MinipoolDetails {
		statusTime := time.Unix(mpd.StatusTime.Int64(), 0)
		if mpd.Status == rptypes.MinipoolStatus_Prelaunch && blockTime.Sub(statusTime) >= launchTimeout {
			mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
			if err != nil {
				return nil, fmt.Errorf("error creating binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
			}
			timedOutMinipools = append(timedOutMinipools, mp)
		}
	}

	// Return
	return timedOutMinipools, nil
}

// Dissolve a minipool
func (t *DissolveTimedOutMinipools) dissolveMinipool(mp minipool.IMinipool) error {
	address := mp.Common().Address
	// Log
	t.log.Printlnf("Dissolving minipool %s...", address.Hex())

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the tx info
	txInfo, err := mp.Common().Dissolve(opts)
	if err != nil {
		return fmt.Errorf("error getting dissolve tx for minipool %s: %w", address.Hex(), err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating dissolve TX failed: %s", txInfo.SimulationResult.SimulationError)
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
	t.log.Printlnf("Successfully dissolved minipool %s.", address.Hex())

	// Return
	return nil
}
