package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/moby/moby/client"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/rocketpool/validator"
	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
)

type megapoolStakeAction int

const (
	stakeSkip megapoolStakeAction = iota
	stakeWithProof
	stakeWithoutProof
)

// megapoolStakeChoice is the pure input to the proof-versus-proofless decision.
type megapoolStakeChoice struct {
	inPrestake        bool
	inBeaconState     bool
	prooflessEnabled  bool
	signatureRetained bool
	deadline          *big.Int
	headTimestamp     uint64
}

type prestakeStakeCandidate struct {
	pubkey      types.ValidatorPubkey
	beaconIndex string
}

// chooseMegapoolStake prefers stakeWithoutProof once the challenge deadline has
// passed. A proof is only required while that deadline is still running and the
// validator is already in the beacon state.
func chooseMegapoolStake(c megapoolStakeChoice) megapoolStakeAction {
	if !c.inPrestake {
		return stakeSkip
	}
	if c.prooflessEnabled && c.signatureRetained && challengeDeadlineElapsed(c.deadline, c.headTimestamp) {
		return stakeWithoutProof
	}
	if c.inBeaconState {
		return stakeWithProof
	}
	return stakeSkip
}

// prooflessStakeSupported reports whether the megapool delegate implements the
// proofless stake getters. Delegate version 1 reverts on those calls.
func prooflessStakeSupported(delegateVersion uint8) bool {
	return delegateVersion >= 2
}

// validatorNeedsStakeProof reports whether staking this validator requires the
// beacon state. Membership in that state is unknown here: the deadline is
// checked first, and a validator that can stake without a proof does not need it.
func validatorNeedsStakeProof(c megapoolStakeChoice, hasBeaconIndex bool) bool {
	if !hasBeaconIndex {
		return false
	}
	withoutState := c
	withoutState.inBeaconState = false
	return chooseMegapoolStake(withoutState) != stakeWithoutProof
}

// challengeDeadlineElapsed reports whether the execution block timestamp has reached the validator's stored deadline.
func challengeDeadlineElapsed(deadline *big.Int, headTimestamp uint64) bool {
	if deadline == nil || deadline.Sign() <= 0 {
		return false
	}
	return deadline.Cmp(new(big.Int).SetUint64(headTimestamp)) <= 0
}

// validatorInSelectedBeaconState reports whether validatorIndex is inside the beacon state selected for proofs.
// An empty index means the beacon chain does not know the validator yet.
func validatorInSelectedBeaconState(validatorIndex string, stateValidatorCount int) (bool, error) {
	if validatorIndex == "" {
		return false, nil
	}
	index, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing validator index %s: %w", validatorIndex, err)
	}
	return index < uint64(stateValidatorCount), nil
}

// Stake megapool validator task
type stakeMegapoolValidator struct {
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

// Create stake megapool validator task
func newStakeMegapoolValidator(c *cli.Command, logger log.ColorLogger) (*stakeMegapoolValidator, error) {

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

	gas := loadAutoTxGas(cfg, &logger)

	// Return task
	return &stakeMegapoolValidator{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		d:              d,
		gasThreshold:   gas.thresholdGwei,
		maxFee:         gas.maxFee,
		maxPriorityFee: gas.maxPriorityFee,
		gasLimit:       0,
	}, nil

}

// Prestake megapool validator
func (t *stakeMegapoolValidator) run(state *state.NetworkStateIndex) error {
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
	validatorInfo, err := services.GetMegapoolValidatorDetails(t.rp, t.bc, megapoolAddress, opts, false)
	if err != nil {
		return err
	}

	// Every prestaked validator is a candidate. An empty beacon index still needs the proofless path.
	validatorsToStake := make(map[uint32]prestakeStakeCandidate)
	for i := range validatorInfo {
		if !validatorInfo[i].InPrestake {
			continue
		}
		validatorsToStake[validatorInfo[i].ValidatorId] = prestakeStakeCandidate{
			pubkey:      types.ValidatorPubkey(validatorInfo[i].PubKey),
			beaconIndex: validatorInfo[i].BeaconStatus.Index,
		}
	}
	if len(validatorsToStake) == 0 {
		return nil
	}

	// getStakeWithoutProofEnabled and stakeWithoutProof are Saturn 2 methods.
	// A failure here must not send a transaction that the delegate would revert.
	prooflessEnabled := false
	if state.Saturn2Deployed {
		prooflessEnabled, err = protocol.GetStakeWithoutProofEnabled(t.rp, opts)
		if err != nil {
			return err
		}
	}
	if prooflessEnabled {
		delegateVersion, err := rocketpool.GetContractVersion(t.rp, megapoolAddress, opts)
		if err != nil {
			return fmt.Errorf("error getting megapool %s delegate version: %w", megapoolAddress.Hex(), err)
		}
		if !prooflessStakeSupported(delegateVersion) {
			t.log.Printlnf("Megapool %s delegate version %d stakes with a beacon proof.", megapoolAddress.Hex(), delegateVersion)
			prooflessEnabled = false
		}
	}
	if !prooflessEnabled && !candidatesHaveBeaconIndex(validatorsToStake) {
		return nil
	}

	// The deadline uses the execution head timestamp. The beacon state is downloaded
	// only for validators that still need a proof, after proofless stakes are sent.
	var headTimestamp uint64
	if prooflessEnabled {
		headTimestamp, err = t.executionHeadTimestamp()
		if err != nil {
			return err
		}
	}

	type proofStake struct {
		id                uint32
		candidate         prestakeStakeCandidate
		deadline          *big.Int
		signatureRetained bool
	}

	stakedValidators := 0
	validatorsProcessed := 0
	const batchSize = 6
	var stakeErr error
	var proofQueue []proofStake
	for validatorId, candidate := range validatorsToStake {
		if validatorsProcessed >= batchSize {
			break
		}
		if !prooflessEnabled {
			if candidate.beaconIndex != "" {
				proofQueue = append(proofQueue, proofStake{id: validatorId, candidate: candidate})
			}
			continue
		}

		deadline, signatureRetained, err := t.prestakeStakeInputs(mp, opts, validatorId)
		if err != nil {
			stakeErr = err
			break
		}
		choice := megapoolStakeChoice{
			inPrestake:        true,
			prooflessEnabled:  true,
			signatureRetained: signatureRetained,
			deadline:          deadline,
			headTimestamp:     headTimestamp,
		}
		if chooseMegapoolStake(choice) == stakeWithoutProof {
			if err = t.submitStake(mp, nil, 0, validatorId, candidate, stakeWithoutProof); err != nil {
				stakeErr = err
				break
			}
			stakedValidators++
			validatorsProcessed++
			continue
		}
		if validatorNeedsStakeProof(choice, candidate.beaconIndex != "") {
			proofQueue = append(proofQueue, proofStake{
				id:                validatorId,
				candidate:         candidate,
				deadline:          deadline,
				signatureRetained: signatureRetained,
			})
			continue
		}
		t.logProoflessSkip(validatorId, deadline, signatureRetained, headTimestamp)
	}

	if stakeErr == nil && validatorsProcessed < batchSize && len(proofQueue) > 0 {
		beaconState, proofTimestamp, err := services.GetHeadBeaconState(t.bc, t.rp.Client)
		if err != nil {
			stakeErr = err
		} else {
			stateValidatorCount := len(beaconState.GetValidators())
			for _, item := range proofQueue {
				if validatorsProcessed >= batchSize {
					break
				}
				inState, beaconIndex, err := t.validatorInBeaconState(item.candidate, stateValidatorCount)
				if err != nil {
					stakeErr = err
					break
				}
				action := chooseMegapoolStake(megapoolStakeChoice{
					inPrestake:        true,
					inBeaconState:     inState,
					prooflessEnabled:  prooflessEnabled,
					signatureRetained: item.signatureRetained,
					deadline:          item.deadline,
					headTimestamp:     headTimestamp,
				})
				if action == stakeSkip {
					if prooflessEnabled {
						t.logProoflessSkip(item.id, item.deadline, item.signatureRetained, headTimestamp)
					} else if beaconIndex != "" {
						t.log.Printlnf("Validator id %d (beacon index %s) is not yet included in the selected beacon state. Will retry on next cycle.", item.id, beaconIndex)
					}
					continue
				}
				if err = t.submitStake(mp, beaconState, proofTimestamp, item.id, item.candidate, action); err != nil {
					stakeErr = err
					break
				}
				stakedValidators++
				validatorsProcessed++
			}
		}
	}

	return t.finishStake(stakedValidators, stakeErr)
}

func (t *stakeMegapoolValidator) executionHeadTimestamp() (uint64, error) {
	header, err := t.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("error getting execution head: %w", err)
	}
	if header == nil {
		return 0, fmt.Errorf("execution head is not available")
	}
	return header.Time, nil
}

func (t *stakeMegapoolValidator) submitStake(mp megapool.Megapool, beaconState eth2.BeaconState, slotTimestamp uint64, validatorId uint32, candidate prestakeStakeCandidate, action megapoolStakeAction) error {
	t.log.Printlnf("The validator id %d needs to be staked", validatorId)
	var err error
	if action == stakeWithoutProof {
		err = t.stakeValidatorWithoutProof(mp, validatorId)
	} else {
		err = t.stakeValidator(t.rp, beaconState, slotTimestamp, mp, validatorId, candidate.pubkey)
	}
	if err != nil {
		t.log.Printlnf("Error staking validator %d: %v", validatorId, err)
	}
	return err
}

func (t *stakeMegapoolValidator) finishStake(stakedValidators int, stakeErr error) error {
	if stakedValidators > 0 {
		if err := validator.RestartValidator(t.cfg, t.bc, &t.log, t.d); err != nil {
			return err
		}
	}
	return stakeErr
}

func (t *stakeMegapoolValidator) stakeValidator(rp *rocketpool.RocketPool, beaconState eth2.BeaconState, slotTimestamp uint64, mp megapool.Megapool, validatorId uint32, validatorPubkey types.ValidatorPubkey) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Crafting a proof that the correct credentials were used on the first beacon chain deposit. This process can take several seconds and is CPU and memory intensive.")

	validatorProof, slotProof, err := services.GetValidatorProofFromState(t.bc, validatorPubkey, beaconState)
	if err != nil {
		t.log.Printlnf("There was an error during the proof creation process: %v", err)
		return err
	}

	t.log.Printlnf("The beacon state proof has been successfully created.")

	// Get the gas limit
	gasLimits, err := services.EstimateMegapoolStakeGas(rp, mp.GetAddress(), validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to stake megapool validator %d: %v", validatorId, err)
		return err
	}
	gas := big.NewInt(int64(gasLimits.Safe))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gasLimits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Call stake
	tx, err := services.StakeMegapool(rp, mp.GetAddress(), validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = transactions.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully staked validator %d.", validatorId)

	// Return
	return nil
}

func (t *stakeMegapoolValidator) validatorInBeaconState(candidate prestakeStakeCandidate, stateValidatorCount int) (bool, string, error) {
	index := candidate.beaconIndex
	if index != "" {
		// Refresh the index. The status used to build the candidate can be one slot ahead of the proof state.
		refreshed, err := t.bc.GetValidatorIndex(candidate.pubkey)
		if err != nil {
			return false, index, err
		}
		index = refreshed
	}
	included, err := validatorInSelectedBeaconState(index, stateValidatorCount)
	return included, index, err
}

func (t *stakeMegapoolValidator) prestakeStakeInputs(mp megapool.Megapool, opts *bind.CallOpts, validatorId uint32) (*big.Int, bool, error) {
	deadline, err := mp.GetValidatorChallengeDeadline(validatorId, opts)
	if err != nil {
		return nil, false, fmt.Errorf("error getting validator %d challenge deadline: %w", validatorId, err)
	}
	signature, err := mp.GetValidatorPrestakeSignature(validatorId, opts)
	if err != nil {
		return nil, false, fmt.Errorf("error getting validator %d prestake signature: %w", validatorId, err)
	}
	return deadline, len(signature) != 0, nil
}

func (t *stakeMegapoolValidator) logProoflessSkip(validatorId uint32, deadline *big.Int, signatureRetained bool, headTimestamp uint64) {
	if signatureRetained && deadline != nil && deadline.Sign() > 0 && !challengeDeadlineElapsed(deadline, headTimestamp) {
		t.log.Printlnf("Validator id %d is not yet included in the selected beacon state and the challenge period has not elapsed. Will retry on next cycle.", validatorId)
		return
	}
	t.log.Printlnf("Validator id %d is not yet included in the selected beacon state and is not eligible for proofless staking. Will retry on next cycle.", validatorId)
}

func (t *stakeMegapoolValidator) stakeValidatorWithoutProof(mp megapool.Megapool, validatorId uint32) error {
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Staking validator %d without a beacon state proof. The challenge period has elapsed.", validatorId)

	gasLimits, err := mp.EstimateStakeWithoutProofGas(validatorId, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to stake megapool validator %d without a proof: %v", validatorId, err)
		return err
	}
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}
	if !gasLimits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gasLimits.Safe

	txHash, err := mp.StakeWithoutProof(validatorId, opts)
	if err != nil {
		return err
	}
	if err := transactions.PrintAndWaitForTransaction(t.cfg, txHash, t.rp.Client, &t.log); err != nil {
		return err
	}
	t.log.Printlnf("Successfully staked validator %d.", validatorId)
	return nil
}

func candidatesHaveBeaconIndex(candidates map[uint32]prestakeStakeCandidate) bool {
	for _, candidate := range candidates {
		if candidate.beaconIndex != "" {
			return true
		}
	}
	return false
}
