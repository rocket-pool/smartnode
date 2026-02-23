package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Prestake megapool validator task
type prestakeMegapoolValidator struct {
	c                   *cli.Context
	log                 log.ColorLogger
	cfg                 *config.RocketPoolConfig
	w                   wallet.Wallet
	rp                  *rocketpool.RocketPool
	d                   *client.Client
	gasThreshold        float64
	maxFee              *big.Int
	maxPriorityFee      *big.Int
	gasLimit            uint64
	autoAssignmentDelay uint16
}

// Create prestake megapool validator task
func newPrestakeMegapoolValidator(c *cli.Context, logger log.ColorLogger) (*prestakeMegapoolValidator, error) {

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
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = eth.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	autoAssignmentDelay := cfg.Smartnode.AutoAssignmentDelay.Value.(uint16)

	// Return task
	return &prestakeMegapoolValidator{
		c:                   c,
		log:                 logger,
		cfg:                 cfg,
		w:                   w,
		rp:                  rp,
		d:                   d,
		gasThreshold:        gasThreshold,
		maxFee:              maxFee,
		maxPriorityFee:      priorityFee,
		gasLimit:            0,
		autoAssignmentDelay: autoAssignmentDelay,
	}, nil

}

// Prestake megapool validator
func (t *prestakeMegapoolValidator) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for megapool validators to pre-stake...")

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check if the megapool is deployed
	deployed, err := megapool.GetMegapoolDeployed(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !deployed {
		return nil
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}

	// Check the next megapool address to be assigned (at the top of the queue)
	nextAssignment, err := deposit.GetQueueTop(t.rp, nil)
	if err != nil {
		return err
	}

	// Check if the next megapool address is the same as the deployed megapool address
	if nextAssignment.Receiver.Cmp(megapoolAddress) != 0 {
		return nil
	}

	// Log
	t.log.Printlnf("The next validator to be assigned belongs to this node's megapool")

	// Check when the last assignment happened and wait autoAssignmentDelay hours
	// Get the head moved block
	block, err := t.rp.Client.HeaderByNumber(context.Background(), nextAssignment.HeadMovedBlock)
	if err != nil {
		return fmt.Errorf("Can't get the block time when the last assignment happened: %w", err)
	}
	lastAssignment := time.Unix(int64(block.Time), 0)

	remainingTime := lastAssignment.Add(time.Duration(t.autoAssignmentDelay) * time.Hour).Sub(time.Now())
	if remainingTime < 0 {
		t.log.Printlnf("%d hours have passed since the last assignment. Trying to assign", t.autoAssignmentDelay)

		// Check if the assignment is possible
		if nextAssignment.AssignmentPossible {
			// Call assign
			t.assignDeposit(nil)
		}
	} else {
		t.log.Printlnf("Time left until the automatic stake %s", remainingTime)
	}

	// Return
	return nil

}

func (t *prestakeMegapoolValidator) assignDeposit(callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := deposit.EstimateAssignDepositsGas(t.rp, big.NewInt(1), opts)
	if err != nil {
		t.log.Printlnf("error estimating assignment %w", err)
		return err
	}
	gas := big.NewInt(int64(gasInfo.SafeGasLimit))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Call assign
	hash, err := deposit.AssignDeposits(t.rp, big.NewInt(1), opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully assigned ETH to the next megapool validator.")

	// Return
	return nil
}
