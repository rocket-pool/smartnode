package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

type autoInitVotingPower struct {
	c              *cli.Context
	log            *log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
	nodeAddress    common.Address
}

func newAutoInitVotingPower(c *cli.Context, logger log.ColorLogger, gasThreshold float64) (*autoInitVotingPower, error) {
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

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested priority fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Return task
	return &autoInitVotingPower{
		c:              c,
		log:            &logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
		nodeAddress:    account.Address,
	}, nil

}

// Auto Initialize Vote Power
func (t *autoInitVotingPower) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for node initialized voting power...")

	// Check if voting is initialized then initialize if it isn't
	votingInitialized, err := network.GetVotingInitialized(t.rp, t.nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("Error getting voting initialized status %w", err)
	}
	if !votingInitialized {
		err := t.submitInitializeVotingPower()
		if err != nil {
			return fmt.Errorf("Error submitting initialize voting power %w", err)
		}
	}

	// Return
	return nil
}

func (t *autoInitVotingPower) submitInitializeVotingPower() error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateInitializeVotingGas(t.rp, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to initialize voting: %w", err)
	}
	gas := big.NewInt(int64(gasInfo.SafeGasLimit))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		return nil
	}

	// Lower the priority fee when the suggested maxfee is lower than the user requested priority fee
	if maxFee.Cmp(t.maxPriorityFee) < 0 {
		t.maxPriorityFee = new(big.Int).Div(maxFee, big.NewInt(2))
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	// Initialize the Voting Power
	hash, err := network.InitializeVoting(t.rp, opts)
	if err != nil {
		return fmt.Errorf("Error initializing voting: %w", err)
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully Initialized Vote Power.")

	// Return
	return nil
}
