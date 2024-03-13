package node

import (
	"fmt"
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Distribute minipools task
type DistributeMinipools struct {
	sp                  *services.ServiceProvider
	log                 log.ColorLogger
	cfg                 *config.SmartNodeConfig
	w                   *wallet.LocalWallet
	rp                  *rocketpool.RocketPool
	bc                  beacon.IBeaconClient
	d                   *client.Client
	mpMgr               *minipool.MinipoolManager
	gasThreshold        float64
	distributeThreshold *big.Int
	eight               *big.Int
	maxFee              *big.Int
	maxPriorityFee      *big.Int
}

// Create distribute minipools task
func NewDistributeMinipools(sp *services.ServiceProvider, logger log.ColorLogger) *DistributeMinipools {
	return &DistributeMinipools{
		sp:    sp,
		log:   logger,
		eight: eth.EthToWei(8),
	}
}

// Distribute minipools
func (t *DistributeMinipools) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.bc = t.sp.GetBeaconClient()
	t.d = t.sp.GetDocker()
	t.w = t.sp.GetWallet()
	nodeAddress, _ := t.w.GetAddress()
	t.maxFee, t.maxPriorityFee = getAutoTxInfo(t.cfg, &t.log)
	t.gasThreshold = t.cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	// Check if auto-distributing is disabled
	if t.gasThreshold == 0 {
		t.log.Println("Automatic tx gas threshold is 0, disabling auto-distribute.")
		return nil
	}
	distributeThreshold := t.cfg.Smartnode.DistributeThreshold.Value.(float64)
	// Safety clamp
	if distributeThreshold >= 8 {
		t.log.Printlnf("WARNING: Auto-distribute threshold is more than 8 ETH (%.6f ETH), reducing to 7.5 ETH for safety", distributeThreshold)
		distributeThreshold = 7.5
	} else if distributeThreshold == 0 {
		t.log.Println("Auto-distribute threshold is 0, disabling auto-distribute.")
		return nil
	}
	t.distributeThreshold = eth.EthToWei(distributeThreshold)

	// Log
	t.log.Println("Checking for minipools to distribute...")

	// Get prelaunch minipools
	minipools, err := t.getDistributableMinipools(nodeAddress, state)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) can have their balances distributed...", len(minipools))
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Get distribute minipools submissions
	txSubmissions := make([]*eth.TransactionSubmission, len(minipools))
	for i, mpd := range minipools {
		txSubmissions[i], err = t.createDistributeMinipoolTx(mpd)
		if err != nil {
			t.log.Println(fmt.Errorf("error preparing submission to distribute minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
	}

	// Distribute
	err = t.distributeMinipools(txSubmissions)
	if err != nil {
		return fmt.Errorf("error distributing minipools: %w", err)
	}

	// Return
	return nil
}

// Get distributable minipools
func (t *DistributeMinipools) getDistributableMinipools(nodeAddress common.Address, state *state.NetworkState) ([]*rpstate.NativeMinipoolDetails, error) {
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

// Get submission info for distributing a minipool
func (t *DistributeMinipools) createDistributeMinipoolTx(mpd *rpstate.NativeMinipoolDetails) (*eth.TransactionSubmission, error) {
	// Log
	t.log.Printlnf("Preparing to distribute minipool %s (total balance of %.6f ETH)...", mpd.MinipoolAddress.Hex(), eth.WeiToEth(mpd.Balance))

	// Get the updated minipool interface
	mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("minipool %s cannot be converted to v3 (current version: %d)", mpd.MinipoolAddress.Hex(), mpd.Version)
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	txInfo, err := mpv3.DistributeBalance(opts, true)
	if err != nil {
		return nil, fmt.Errorf("error getting distribute minipool tx for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	if txInfo.SimError != "" {
		return nil, fmt.Errorf("simulating distribute minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimError)
	}

	submission, err := core.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating distribute tx submission for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	return submission, nil
}

// Distribute all available minipools
func (t *DistributeMinipools) distributeMinipools(submissions []*eth.TransactionSubmission) error {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(&t.log)
		if err != nil {
			return err
		}
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	if !gas.PrintAndCheckGasInfoForBatch(submissions, true, t.gasThreshold, &t.log, maxFee) {
		return nil
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransactionBatch(t.cfg, t.rp, &t.log, submissions, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully distributed balance of all minipools.")
	return nil
}
