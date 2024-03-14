package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Promote minipools task
type promoteMinipools struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	d              *client.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create promote minipools task
func newPromoteMinipools(c *cli.Context, logger log.ColorLogger) (*promoteMinipools, error) {

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

	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Return task
	return &promoteMinipools{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		d:              d,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Stake prelaunch minipools
func (t *promoteMinipools) run(state *state.NetworkState) error {

	// Log
	t.log.Println("Checking for minipools to promote...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get prelaunch minipools
	minipools, err := t.getVacantMinipools(nodeAccount.Address, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for promotion...", len(minipools))

	// Promote minipools
	for _, mpd := range minipools {
		_, err := t.promoteMinipool(mpd, opts)
		alerting.AlertMinipoolPromoted(t.cfg, mpd.MinipoolAddress, err == nil)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not promote minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
	}

	// Return
	return nil

}

// Get vacant minipools
func (t *promoteMinipools) getVacantMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {

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
		if mpd.IsVacant && mpd.Status == types.Prelaunch {
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
func (t *promoteMinipools) promoteMinipool(mpd *rpstate.NativeMinipoolDetails, callOpts *bind.CallOpts) (bool, error) {

	// Log
	t.log.Printlnf("Promoting minipool %s...", mpd.MinipoolAddress.Hex())

	// Get the updated minipool interface
	mp, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, callOpts)
	if err != nil {
		return false, fmt.Errorf("cannot create binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return false, fmt.Errorf("cannot promote minipool %s because its delegate version is too low (v%d); please update the delegate to promote it", mp.GetAddress().Hex(), mp.GetVersion())
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	gasInfo, err := mpv3.EstimatePromoteGas(opts)
	if err != nil {
		return false, fmt.Errorf("Could not estimate the gas required to promote the minipool: %w", err)
	}
	var gas *big.Int
	if t.gasLimit != 0 {
		gas = new(big.Int).SetUint64(t.gasLimit)
	} else {
		gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return false, err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		// Check for the timeout buffer
		creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
		isDue, timeUntilDue, err := api.IsTransactionDue(t.rp, creationTime)
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
	opts.GasLimit = gas.Uint64()

	// Promote minipool
	hash, err := mpv3.Promote(opts)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully promoted minipool %s.", mpd.MinipoolAddress.Hex())

	// Return
	return true, nil

}
