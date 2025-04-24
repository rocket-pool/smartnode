package node

import (
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Stake megapool validator task
type stakeMegapoolValidator struct {
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

// Create stake megapool validator task
func newStakeMegapoolValidator(c *cli.Context, logger log.ColorLogger) (*stakeMegapoolValidator, error) {

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
	bc, err := services.GetBeaconClient(c)
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
	return &stakeMegapoolValidator{
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

// Prestake megapool validator
func (t *stakeMegapoolValidator) run(state *state.NetworkState) error {
	if !state.IsSaturnDeployed {
		return nil
	}

	// Log
	t.log.Println("Checking for megapool validators to stake...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check if the megapool is deployed
	deployed, err := megapool.GetMegapoolDeployed(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}
	if !deployed {
		return nil
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(t.rp, megapoolAddress, nil)
	if err != nil {
		return err
	}

	// Iterate over megapool validators checking whether they're ready to stake
	validatorCount, err := mp.GetValidatorCount(nil)
	if err != nil {
		return err
	}
	validatorInfo, err := services.GetMegapoolValidatorDetails(t.rp, t.bc, mp, megapoolAddress, uint32(validatorCount))
	if err != nil {
		return err
	}

	for i := uint32(0); i < uint32(validatorCount); i++ {
		if validatorInfo[i].InPrestake {
			// Log
			t.log.Printlnf("The validator %d needs to be staked", validatorInfo[i].ValidatorId)

			// Call Stake
			t.stakeValidator(mp, validatorInfo[i].ValidatorId, state, types.ValidatorPubkey(validatorInfo[i].PubKey), opts)
		}
	}

	// Return
	return nil

}

func (t *stakeMegapoolValidator) stakeValidator(mp megapool.Megapool, validatorId uint32, state *state.NetworkState, validatorPubkey types.ValidatorPubkey, callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("[STARTED] Crafting a proof that the correct credentials were used on the first beacon chain deposit. This process can take several seconds and is CPU and memory intensive. If you don't see a [FINISHED] log entry your system may not have enough resources to perform this operation.")

	proof, err := services.GetStakeValidatorInfo(t.c, t.w, state.BeaconConfig, mp.GetAddress(), validatorPubkey)
	if err != nil {
		t.log.Printlnf("[ERROR] There was an error during the proof creation process: %w", err)
		return err
	}

	t.log.Printlnf("[FINISHED] The beacon state proof has been successfully created.")

	// Get the gas limit
	gasInfo, err := mp.EstimateStakeGas(validatorId, proof, opts)
	if err != nil {
		return err
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
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Call stake
	hash, err := mp.Stake(validatorId, proof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully staked validator %d.", validatorId)

	// Return
	return nil
}
