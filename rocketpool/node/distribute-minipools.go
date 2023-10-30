package node

import (
	"fmt"
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"

	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/state"
	"github.com/rocket-pool/smartnode/rocketpool/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Distribute minipools task
type distributeMinipools struct {
	sp                  *services.ServiceProvider
	log                 log.ColorLogger
	cfg                 *config.RocketPoolConfig
	w                   *wallet.LocalWallet
	rp                  *rocketpool.RocketPool
	bc                  beacon.Client
	d                   *client.Client
	gasThreshold        float64
	distributeThreshold *big.Int
	disabled            bool
	eight               *big.Int
	maxFee              *big.Int
	maxPriorityFee      *big.Int
	gasLimit            uint64
}

// Create distribute minipools task
func NewDistributeMinipools(sp *services.ServiceProvider, logger log.ColorLogger) (*distributeMinipools, error) {
	// Return task
	return &distributeMinipools{
		sp:       sp,
		log:      logger,
		eight:    eth.EthToWei(8),
		gasLimit: 0,
	}, nil
}

// Distribute minipools
func (t *distributeMinipools) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.bc = t.sp.GetBeaconClient()
	t.d = t.sp.GetDocker()
	t.w = t.sp.GetWallet()
	nodeAddress, hasNodeAddress := t.w.GetAddress()
	if !hasNodeAddress {
		return nil
	}
	t.disabled, t.maxFee, t.maxPriorityFee, t.gasThreshold = getAutoTxInfo(t.cfg, &t.log)

	distributeThreshold := t.cfg.Smartnode.DistributeThreshold.Value.(float64)
	// Safety clamp
	if distributeThreshold >= 8 {
		t.log.Printlnf("WARNING: Auto-distribute threshold is more than 8 ETH (%.6f ETH), reducing to 7.5 ETH for safety", distributeThreshold)
		distributeThreshold = 7.5
	} else if distributeThreshold == 0 {
		t.log.Println("Auto-distribute threshold is 0, disabling auto-distribute.")
		t.disabled = true
	}
	t.distributeThreshold = eth.EthToWei(distributeThreshold)

	// Check if auto-distribute is disabled
	if t.disabled {
		return nil
	}

	// Log
	t.log.Println("Checking for minipools to distribute...")
	mpMgr, err := minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get prelaunch minipools
	minipools, err := t.getDistributableMinipools(nodeAddress, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) can have their balances distributed...", len(minipools))

	// Distribute minipools
	for _, mpd := range minipools {
		_, err := t.distributeMinipool(mpMgr, mpd, opts)
		if err != nil {
			t.log.Println(fmt.Errorf("error distributing balance of minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
	}

	// Return
	return nil
}

// Get distributable minipools
func (t *distributeMinipools) getDistributableMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {
	// Filter minipools by status
	distributableMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		if mpd.Status != rptypes.MinipoolStatus_Staking || mpd.Finalised {
			// Ignore minipools that aren't staking and non-finalized
			continue
		}
		if mpd.Version < 3 {
			// Ignore minipools with legacy delegates
			continue
		}
		if mpd.DistributableBalance.Cmp(t.eight) >= 0 {
			// Ignore minipools with distributable balances >= 8 ETH
			continue
		}
		if mpd.DistributableBalance.Cmp(t.distributeThreshold) >= 0 {
			distributableMinipools = append(distributableMinipools, mpd)
		}
	}

	// Return
	return distributableMinipools, nil
}

// Distribute a minipool
func (t *distributeMinipools) distributeMinipool(mpMgr *minipool.MinipoolManager, mpd *rpstate.NativeMinipoolDetails, callOpts *bind.CallOpts) (bool, error) {
	// Log
	t.log.Printlnf("Distributing minipool %s (total balance of %.6f ETH)...", mpd.MinipoolAddress.Hex(), eth.WeiToEth(mpd.Balance))

	mp, err := mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return false, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return false, fmt.Errorf("minipool %s cannot be converted to v3 (current version: %d)", mpd.MinipoolAddress.Hex(), mpd.Version)
	}
	txInfo, err := mpv3.DistributeBalance(opts, true)
	if err != nil {
		return false, fmt.Errorf("error getting distribute minipool tx for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	if txInfo.SimError != "" {
		return false, fmt.Errorf("simulating distribute minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimError)
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
		return false, nil
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
	t.log.Printlnf("Successfully distributed balance of minipool %s.", mpd.MinipoolAddress.Hex())

	// Return
	return true, nil
}
