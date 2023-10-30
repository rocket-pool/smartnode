package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"

	"github.com/rocket-pool/smartnode/rocketpool/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool/common/state"
	"github.com/rocket-pool/smartnode/rocketpool/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Promote minipools task
type PromoteMinipools struct {
	sp             *services.ServiceProvider
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.LocalWallet
	rp             *rocketpool.RocketPool
	mpMgr          *minipool.MinipoolManager
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create promote minipools task
func NewPromoteMinipools(sp *services.ServiceProvider, logger log.ColorLogger) (*PromoteMinipools, error) {
	return &PromoteMinipools{
		sp:       sp,
		log:      logger,
		gasLimit: 0,
	}, nil
}

// Stake prelaunch minipools
func (t *PromoteMinipools) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.w = t.sp.GetWallet()
	nodeAddress, _ := t.w.GetAddress()
	t.maxFee, t.maxPriorityFee = getAutoTxInfo(t.cfg, &t.log)

	// Log
	t.log.Println("Checking for minipools to promote...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get prelaunch minipools
	minipools, err := t.getVacantMinipools(nodeAddress, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for promotion...", len(minipools))
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Promote minipools
	timeoutBig := state.NetworkDetails.MinipoolLaunchTimeout
	timeout := time.Duration(timeoutBig.Uint64()) * time.Second
	for _, mpd := range minipools {
		_, err := t.promoteMinipool(mpd, timeout, opts)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not promote minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
	}

	// Return
	return nil
}

// Get vacant minipools
func (t *PromoteMinipools) getVacantMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {
	vacantMinipools := []*rpstate.NativeMinipoolDetails{}

	// Get the scrub period
	scrubPeriod := state.NetworkDetails.PromotionScrubPeriod

	// Get the time of the target block
	block, err := t.rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("Can't get the latest block time: %w", err)
	}
	blockTime := time.Unix(int64(block.Time), 0)

	// Filter by vacancy
	mpds := state.MinipoolDetailsByNode[nodeAddress]
	for _, mpd := range mpds {
		if mpd.IsVacant && mpd.Status == types.MinipoolStatus_Prelaunch {
			creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
			remainingTime := creationTime.Add(scrubPeriod).Sub(blockTime)
			if remainingTime < 0 {
				vacantMinipools = append(vacantMinipools, mpd)
			} else {
				t.log.Printlnf("Minipool %s has %s left until it can be promoted.", mpd.MinipoolAddress.Hex(), remainingTime)
			}
		}
	}

	// Return
	return vacantMinipools, nil
}

// Promote a minipool
func (t *PromoteMinipools) promoteMinipool(mpd *rpstate.NativeMinipoolDetails, minipoolLaunchTimeout time.Duration, callOpts *bind.CallOpts) (bool, error) {
	// Log
	t.log.Printlnf("Promoting minipool %s...", mpd.MinipoolAddress.Hex())

	// Get the updated minipool interface
	mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return false, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return false, fmt.Errorf("cannot promote minipool %s because its delegate version is too low (v%d); please update the delegate to promote it", mpd.MinipoolAddress.Hex(), mpd.Version)
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	txInfo, err := mpv3.Promote(opts)
	if err != nil {
		return false, fmt.Errorf("error getting promote minipool tx for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	if txInfo.SimError != "" {
		return false, fmt.Errorf("simulating promote minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimError)
	}
	var gasLimit *big.Int
	if t.gasLimit != 0 {
		gasLimit = new(big.Int).SetUint64(t.gasLimit)
	} else {
		gasLimit = new(big.Int).SetUint64(txInfo.GasInfo.SafeGasLimit)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(&t.log)
		if err != nil {
			return false, err
		}
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		// Check for the timeout buffer
		creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
		isDue, timeUntilDue, err := tx.IsTransactionDue(t.rp, creationTime, minipoolLaunchTimeout)
		if err != nil {
			t.log.Printlnf("Error checking if minipool is due: %s\nPromoting now for safety...", err.Error())
		}
		if !isDue {
			t.log.Printlnf("Time until promoting will be forced for safety: %s", timeUntilDue)
			return false, nil
		}

		t.log.Println("NOTICE: The minipool has exceeded half of the timeout period, so it will be force-promoted at the current gas price.")
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gasLimit.Uint64()

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, &t.log, txInfo, opts)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully promoted minipool %s.", mpd.MinipoolAddress.Hex())

	// Return
	return true, nil
}
