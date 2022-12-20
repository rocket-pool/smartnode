package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
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
	bc             beacon.Client
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Check if auto-staking is disabled
	gasThreshold := cfg.Smartnode.MinipoolStakeGasThreshold.Value.(float64)

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
		bc:             bc,
		d:              d,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Stake prelaunch minipools
func (t *promoteMinipools) run() error {

	// Reload the wallet (in case a call to `node deposit` changed it)
	if err := t.w.Reload(); err != nil {
		return err
	}

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}

	// Log
	t.log.Println("Checking for minipools to promote...")

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get prelaunch minipools
	minipools, err := t.getVacantMinipools(nodeAccount.Address)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Get eth2 config
	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for promotion...", len(minipools))

	// Promote minipools
	for _, mp := range minipools {
		_, err := t.promoteMinipool(mp, eth2Config)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not promote minipool %s: %w", mp.GetAddress().Hex(), err))
			return err
		}
	}

	// Return
	return nil

}

// Get vacant minipools
func (t *promoteMinipools) getVacantMinipools(nodeAddress common.Address) ([]minipool.Minipool, error) {

	// Get node minipool addresses
	addresses, err := minipool.GetNodeMinipoolAddresses(t.rp, nodeAddress, nil)
	if err != nil {
		return []minipool.Minipool{}, err
	}

	// Create minipool contracts
	minipools := make([]minipool.Minipool, len(addresses))
	for mi, address := range addresses {
		mp, err := minipool.NewMinipool(t.rp, address, nil)
		if err != nil {
			return []minipool.Minipool{}, err
		}
		minipools[mi] = mp
	}

	// Data
	var wg errgroup.Group
	statuses := make([]minipool.StatusDetails, len(minipools))

	// Load minipool statuses
	for mi, mp := range minipools {
		mi, mp := mi, mp
		wg.Go(func() error {
			status, err := mp.GetStatusDetails(nil)
			if err == nil {
				statuses[mi] = status
			}
			return err
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return []minipool.Minipool{}, err
	}

	// Get the scrub period
	scrubPeriodSeconds, err := trustednode.GetScrubPeriod(t.rp, nil)
	if err != nil {
		return []minipool.Minipool{}, err
	}
	scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

	// Get the time of the latest block
	latestEth1Block, err := t.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return []minipool.Minipool{}, fmt.Errorf("Can't get the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Filter minipools by vacancy
	vacantMinipools := []minipool.Minipool{}
	for mi, mp := range minipools {
		if statuses[mi].IsVacant && statuses[mi].Status == types.Prelaunch {
			creationTime := statuses[mi].StatusTime
			remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
			if remainingTime < 0 {
				vacantMinipools = append(vacantMinipools, mp)
			} else {
				t.log.Printlnf("Minipool %s has %s left until it can be promoted.", mp.GetAddress().Hex(), remainingTime)
			}
		}
	}

	// Return
	return vacantMinipools, nil

}

// Promote a minipool
func (t *promoteMinipools) promoteMinipool(mp minipool.Minipool, eth2Config beacon.Eth2Config) (bool, error) {

	// Log
	t.log.Printlnf("Promoting minipool %s...", mp.GetAddress().Hex())

	// Get the updated minipool interface
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
		return false, fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		// Check for the timeout buffer
		creationTime, err := mp.GetStatusTime(nil)
		if err != nil {
			t.log.Printlnf("Error checking minipool launch time: %s\nPromoting now for safety...", err.Error())
		}
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
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully promoted minipool %s.", mp.GetAddress().Hex())

	// Return
	return true, nil

}
