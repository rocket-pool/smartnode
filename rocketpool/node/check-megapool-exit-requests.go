package node

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	rpvalidator "github.com/rocket-pool/smartnode/shared/utils/validator"
)

// Check megapool exit requests task
type checkMegapoolExitRequests struct {
	c              *cli.Command
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	d              *client.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
	intervalSize   *big.Int
}

// Create check megapool exit requests task
func newCheckMegapoolExitRequests(c *cli.Command, logger log.ColorLogger) (*checkMegapoolExitRequests, error) {

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

	// Get the user-requested priority fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = eth.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Get the event log interval
	eventLogInterval, err := cfg.GetEventLogInterval()
	if err != nil {
		return nil, err
	}

	// Return task
	return &checkMegapoolExitRequests{
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
		intervalSize:   big.NewInt(int64(eventLogInterval)),
	}, nil

}

// Check for megapool validators that did not respond to an exit request
func (t *checkMegapoolExitRequests) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for megapool validators that did not respond to an exit request...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the cooperative exit phase duration
	cooperativeExitPhase, err := protocol.GetCooperativeExitPhase(t.rp, opts)
	if err != nil {
		return err
	}

	// Search MegapoolExitRequested events over a window covering the cooperative exit phase
	lookbackBlocks := uint64(cooperativeExitPhase / (time.Duration(state.BeaconConfig.SecondsPerSlot) * time.Second))
	if lookbackBlocks == 0 {
		lookbackBlocks = 1
	}
	var fromBlock *big.Int
	if state.ElBlockNumber > lookbackBlocks {
		fromBlock = big.NewInt(int64(state.ElBlockNumber - lookbackBlocks))
	} else {
		fromBlock = big.NewInt(0)
	}
	toBlock := big.NewInt(int64(state.ElBlockNumber))

	exitRequests, err := network.GetMegapoolExitRequests(t.rp, t.intervalSize, fromBlock, toBlock, opts)
	if err != nil {
		return err
	}
	if len(exitRequests) == 0 {
		return nil
	}

	// Get this node's megapool address, if one is deployed
	ownMegapoolDeployed := false
	nodeDetails, exists := state.NodeDetailsByAddress[nodeAccount.Address]
	if exists {
		ownMegapoolDeployed = nodeDetails.MegapoolDeployed
	}

	for _, request := range exitRequests {

		status, err := t.bc.GetValidatorStatus(request.Pubkey, nil)
		if err != nil {
			t.log.Printlnf("Error getting the status of validator %s: %s", request.Pubkey.Hex(), err.Error())
			continue
		}
		if !status.Exists {
			t.log.Printlnf("Validator %s not found on the beacon chain", request.Pubkey.Hex())
			continue
		}

		validatorIndex, err := strconv.ParseUint(status.Index, 10, 64)
		if err != nil {
			t.log.Printlnf("Error parsing validator index %s: %s", status.Index, err.Error())
			continue
		}

		// An initiated exit satisfies the request
		if status.ExitEpoch != FarFutureEpoch {
			t.log.Printlnf("Validator %d has already exited", validatorIndex)
			continue
		}

		// If the megapool belongs to this node, cooperate signing and submitting a voluntary exit
		if ownMegapoolDeployed && request.MegapoolAddress == nodeDetails.MegapoolAddress {
			t.log.Printlnf("Megapool %s (validator %d) belongs to this node; submitting a voluntary exit", request.MegapoolAddress.Hex(), validatorIndex)
			err := t.exitOwnMegapoolValidator(state, request, status)
			if err != nil {
				t.log.Printlnf("Error exiting megapool %s validator %d: %s", request.MegapoolAddress.Hex(), request.ValidatorId, err.Error())
			}
			continue
		}

		mp, err := megapool.NewMegapool(t.rp, request.MegapoolAddress, opts)
		if err != nil {
			t.log.Printlnf("Error creating a binding for megapool %s: %s", request.MegapoolAddress.Hex(), err.Error())
			continue
		}

		// Skip requests still within the cooperative exit phase
		deadline := time.Unix(int64(request.RequestTimestamp), 0).Add(cooperativeExitPhase)
		if time.Now().Before(deadline) {
			continue
		}

		// Megapools from version 2 support a forced exit request, so the
		// did-not-exit penalty only applies to version 1
		if mpv2, ok := mp.(megapool.MegapoolV2); ok {
			t.log.Printlnf("Megapool %s (validator %d) uses version %d; submitting ForceExit", request.MegapoolAddress.Hex(), validatorIndex, mp.GetVersion())
			err := t.forceExitMegapoolValidator(mpv2, request)
			if err != nil {
				t.log.Printlnf("Error force-exiting megapool %s validator %d: %s", request.MegapoolAddress.Hex(), request.ValidatorId, err.Error())
			}
			continue
		}

		// Log
		t.log.Printlnf("The validator %d (megapool %s) did not exit within the cooperative exit phase", validatorIndex, request.MegapoolAddress.Hex())

		err = t.penaliseMegapoolValidator(request)
		// dont return if there was an error, just log it so we can continue with the next validator
		if err != nil {
			t.log.Printlnf("Error penalising megapool %s validator %d: %s", request.MegapoolAddress.Hex(), request.ValidatorId, err.Error())
		}
	}

	// Return
	return nil

}

// Sign and broadcast the voluntary exit for a megapool validator belonging to this node
func (t *checkMegapoolExitRequests) exitOwnMegapoolValidator(state *state.NetworkState, request network.MegapoolExitRequest, status beacon.ValidatorStatus) error {

	// Check the validator status on the megapool
	validatorInfo, exists := state.MegapoolValidatorInfo[request.Pubkey]
	if !exists {
		return fmt.Errorf("validator %s not found in the megapool validator info map", request.Pubkey.Hex())
	}
	if !validatorInfo.ValidatorInfo.Staked {
		return fmt.Errorf("megapool %s validator %d is not staked", request.MegapoolAddress.Hex(), request.ValidatorId)
	}

	// Get the validator private key
	validatorKey, err := t.w.GetValidatorKeyByPubkey(request.Pubkey)
	if err != nil {
		return err
	}

	// Get beacon head
	head, err := t.bc.GetBeaconHead()
	if err != nil {
		return err
	}

	// Get voluntary exit signature domain
	signatureDomain, err := t.bc.GetDomainData(eth2types.DomainVoluntaryExit[:], head.Epoch, false)
	if err != nil {
		return err
	}

	// Get signed voluntary exit message
	signature, err := rpvalidator.GetSignedExitMessage(validatorKey, status.Index, head.Epoch, signatureDomain)
	if err != nil {
		return err
	}

	// Broadcast voluntary exit message
	if err := t.bc.ExitValidator(status.Index, head.Epoch, signature); err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted a voluntary exit for validator %s (megapool %s).", status.Index, request.MegapoolAddress.Hex())

	// Return
	return nil
}

func (t *checkMegapoolExitRequests) forceExitMegapoolValidator(mp megapool.MegapoolV2, request network.MegapoolExitRequest) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateForceMegapoolExitGas(t.rp, request.MegapoolAddress, request.ValidatorId, opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to force exit megapool %s validator %d: %w", request.MegapoolAddress.Hex(), request.ValidatorId, err)
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

	// Force exit the validator via the megapool contract
	hash, err := network.ForceMegapoolExit(t.rp, request.MegapoolAddress, request.ValidatorId, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted ForceExit for megapool %s validator %d.", request.MegapoolAddress.Hex(), request.ValidatorId)

	// Return
	return nil
}

func (t *checkMegapoolExitRequests) penaliseMegapoolValidator(request network.MegapoolExitRequest) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimatePenaliseMegapoolValidatorGas(t.rp, request.MegapoolAddress, request.ValidatorId, opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to penalise megapool %s validator %d: %w", request.MegapoolAddress.Hex(), request.ValidatorId, err)
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

	// Penalise the megapool validator for failing to exit within the cooperative phase
	hash, err := network.PenaliseMegapoolValidator(t.rp, request.MegapoolAddress, request.ValidatorId, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully penalised megapool %s (validator %d) for not exiting.", request.MegapoolAddress.Hex(), request.ValidatorId)

	// Return
	return nil
}
