package node

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	rpvalidator "github.com/rocket-pool/smartnode/shared/utils/validator"
)

// TODO: flip to true once the did-not-exit contract method exists
const didNotExitTxEnabled = false

// A minipool validator that did not exit within the cooperative exit phase
type didNotExitValidator struct {
	validatorIndex  uint64
	pubkey          types.ValidatorPubkey
	minipoolAddress common.Address
}

// Check minipool exit requests task
type checkMinipoolExitRequests struct {
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
}

// Create check minipool exit requests task
func newCheckMinipoolExitRequests(c *cli.Command, logger log.ColorLogger) (*checkMinipoolExitRequests, error) {

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

	// Return task
	return &checkMinipoolExitRequests{
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

// Check for minipool validators that did not respond to an exit request
func (t *checkMinipoolExitRequests) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for minipool validators that did not respond to an exit request...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the pending exit requests
	exitRequests, err := minipool.GetMinipoolExitRequests(t.rp, opts)
	if err != nil {
		return err
	}
	if len(exitRequests) == 0 {
		return nil
	}

	// Get the cooperative exit phase duration
	cooperativeExitPhase, err := protocol.GetCooperativeExitPhase(t.rp, opts)
	if err != nil {
		return err
	}

	// Index the minipool details by validator pubkey
	minipoolDetailsByPubkey := make(map[types.ValidatorPubkey]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	for i := range state.MinipoolDetails {
		minipoolDetailsByPubkey[state.MinipoolDetails[i].Pubkey] = &state.MinipoolDetails[i]
	}

	validatorsToProve := []didNotExitValidator{}
	for _, request := range exitRequests {

		status, err := t.bc.GetValidatorStatusByIndex(strconv.FormatUint(request.ValidatorIndex, 10), nil)
		if err != nil {
			t.log.Printlnf("Error getting the status of validator %d: %s", request.ValidatorIndex, err.Error())
			continue
		}
		if !status.Exists {
			t.log.Printlnf("Validator %d not found on the beacon chain", request.ValidatorIndex)
			continue
		}

		// An initiated exit satisfies the request
		if status.ExitEpoch != FarFutureEpoch {
			t.log.Printlnf("Validator %d has already exited", request.ValidatorIndex)
			continue
		}

		minipoolDetails, exists := minipoolDetailsByPubkey[status.Pubkey]
		if !exists {
			t.log.Printlnf("Validator %d does not belong to a known minipool", request.ValidatorIndex)
			continue
		}

		// If the minipool belongs to this node, cooperate: sign and submit the voluntary exit
		if minipoolDetails.NodeAddress == nodeAccount.Address {
			t.log.Printlnf("Minipool %s (validator %d) belongs to this node; submitting a voluntary exit", minipoolDetails.MinipoolAddress.Hex(), request.ValidatorIndex)
			err := t.exitOwnMinipool(minipoolDetails, status)
			if err != nil {
				t.log.Printlnf("Error exiting minipool %s: %s", minipoolDetails.MinipoolAddress.Hex(), err.Error())
			}
			continue
		}

		// Delegates from version 4 support a forced exit request, so the
		// did-not-exit path only applies to older delegates
		if minipoolDetails.Version >= 4 {
			t.log.Printlnf("Minipool %s (validator %d) uses delegate version %d; submitting ForceExit", minipoolDetails.MinipoolAddress.Hex(), request.ValidatorIndex, minipoolDetails.Version)
			err := t.forceExitMinipool(minipoolDetails, opts)
			if err != nil {
				t.log.Printlnf("Error force-exiting minipool %s: %s", minipoolDetails.MinipoolAddress.Hex(), err.Error())
			}
			continue
		}
		// Skip requests still within the cooperative exit phase
		deadline := time.Unix(int64(request.RequestTimestamp), 0).Add(cooperativeExitPhase)
		if time.Now().Before(deadline) {
			continue
		}

		validatorsToProve = append(validatorsToProve, didNotExitValidator{
			validatorIndex:  request.ValidatorIndex,
			pubkey:          status.Pubkey,
			minipoolAddress: minipoolDetails.MinipoolAddress,
		})
	}

	// Check if there are any validators to prove
	if len(validatorsToProve) == 0 {
		return nil
	}

	beaconState, err := services.GetBeaconState(t.bc)
	if err != nil {
		return err
	}

	for _, validator := range validatorsToProve {

		// Log
		t.log.Printlnf("The validator %d (minipool %s) did not exit within the cooperative exit phase", validator.validatorIndex, validator.minipoolAddress.Hex())

		err := t.proveDidNotExit(beaconState, state, validator)
		// dont return if there was an error, just log it so we can continue with the next validator
		if err != nil {
			t.log.Printlnf("Error proving validator %d did not exit: %s", validator.validatorIndex, err.Error())
		}
	}

	// Return
	return nil

}

// Sign and broadcast the voluntary exit for a minipool validator belonging to this node
func (t *checkMinipoolExitRequests) exitOwnMinipool(mpd *rpstate.NativeMinipoolDetails, status beacon.ValidatorStatus) error {

	// Check the minipool status
	if mpd.Status != types.Staking {
		return fmt.Errorf("minipool %s is not in staking status", mpd.MinipoolAddress.Hex())
	}

	// Get the validator private key
	validatorKey, err := t.w.GetValidatorKeyByPubkey(mpd.Pubkey)
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
	t.log.Printlnf("Successfully submitted a voluntary exit for validator %s (minipool %s).", status.Index, mpd.MinipoolAddress.Hex())

	// Return
	return nil
}

func (t *checkMinipoolExitRequests) forceExitMinipool(mpd *rpstate.NativeMinipoolDetails, callOpts *bind.CallOpts) error {

	mp, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, callOpts)
	if err != nil {
		return fmt.Errorf("cannot create binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	mpv4, success := minipool.GetMinipoolAsV4(mp)
	if !success {
		return fmt.Errorf("minipool %s cannot be converted to v4 (current version: %d)", mpd.MinipoolAddress.Hex(), mp.GetVersion())
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := mpv4.EstimateForceExitGas(opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to force exit minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
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

	// Force exit the minipool
	hash, err := mpv4.ForceExit(opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted ForceExit for minipool %s.", mpd.MinipoolAddress.Hex())

	// Return
	return nil
}

func (t *checkMinipoolExitRequests) proveDidNotExit(beaconState eth2.BeaconState, state *state.NetworkState, validator didNotExitValidator) error {

	t.log.Printlnf("[STARTED] Crafting a did-not-exit proof. This process can take several seconds and is CPU and memory intensive. If you don't see a [FINISHED] log entry your system may not have enough resources to perform this operation.")

	// The megapool proof structs are generic SSZ proofs; the address parameter is unused by GetValidatorProof
	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, 0, t.w, state.BeaconConfig, common.Address{}, validator.pubkey, beaconState)
	if err != nil {
		t.log.Printlnf("[ERROR] There was an error during the proof creation process: %s", err.Error())
		return err
	}

	t.log.Printlnf("[FINISHED] The did-not-exit proof for validator %d has been successfully created (exit epoch %d at slot %d).", validator.validatorIndex, validatorProof.Validator.ExitEpoch, slotProof.Slot)

	if !didNotExitTxEnabled {
		// TODO: remove this check once the did-not-exit contract method exists
		t.log.Printlnf("[TODO] The contract method to report that validator %d did not exit is not yet available; skipping the transaction.", validator.validatorIndex)
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := minipool.EstimateNotifyMinipoolDidNotExitGas(t.rp, validator.validatorIndex, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to report that validator %d did not exit: %s", validator.validatorIndex, err.Error())
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

	// Report that the validator did not exit
	tx, err := minipool.NotifyMinipoolDidNotExit(t.rp, validator.validatorIndex, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully reported that validator %d did not exit.", validator.validatorIndex)

	// Return
	return nil
}
