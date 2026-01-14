package node

import (
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/node"
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

// Provision Express Tickets task
type provisionExpress struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              wallet.Wallet
	rp             *rocketpool.RocketPool
	d              *client.Client
	gasThreshold   float64
	disabled       bool
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create provision express tickets task
func newProvisionExpressTickets(c *cli.Context, logger log.ColorLogger) (*provisionExpress, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
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

	// Check if automatic transactions are disabled
	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)
	disabled := false
	if gasThreshold == 0 {
		logger.Println("Automatic tx gas threshold is 0, disabling auto-provision.")
		disabled = true
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
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = eth.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Return task
	return &provisionExpress{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		d:              d,
		gasThreshold:   gasThreshold,
		disabled:       disabled,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Provision Express tickets
func (t *provisionExpress) run(state *state.NetworkState) error {
	if !state.IsSaturnDeployed {
		return nil
	}

	// Check if automatic transactions are disabled
	if t.disabled {
		return nil
	}

	// Log
	t.log.Println("Checking for express tickets to provision...")

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check if express tickets have been provisioned
	provisioned, err := node.GetExpressTicketsProvisioned(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !provisioned {
		err = t.provisionExpress(nodeAccount.Address)
		if err != nil {
			return err
		}
	}

	// Return
	return nil

}

// Provision Express tickets for the node
func (t *provisionExpress) provisionExpress(nodeAddress common.Address) error {

	// Log
	t.log.Printlnf("Provisioning express tickets for %s...", nodeAddress.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := node.EstimateProvisionExpressTicketsGas(t.rp, nodeAddress, opts)
	if err != nil {
		return err
	}
	gas := big.NewInt(int64(gasInfo.SafeGasLimit))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWei(t.cfg)
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

	// Provision Express Tickets
	hash, err := node.ProvisionExpressTickets(t.rp, nodeAddress, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully provisioned express tickets for node address %s.", nodeAddress.Hex())

	// Return
	return nil

}
