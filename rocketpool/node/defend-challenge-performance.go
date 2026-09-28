package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/bindings/types"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

// Defend performance challenges against this node's megapool validators
type defendChallengePerformance struct {
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
	discovery      performanceChallengeDiscovery
}

type megapoolPerformanceChallenge struct {
	challengeId           *big.Int
	megapoolAddress       common.Address
	validatorIds          []uint32
	startEpoch            uint64
	participationCallData []*big.Int
	responseDeadline      *big.Int
	proposer              common.Address
}

// challengedValidator holds a challenged megapool validator's on-chain id
// alongside the beacon-chain identifiers needed to verify its target-vote
// participation.
type challengedValidator struct {
	validatorId uint32
	pubkey      types.ValidatorPubkey
	index       uint64
}

// participationCallData word (a Solidity uint256).
const bitsPerParticipationWord = 256

func (c *megapoolPerformanceChallenge) getChallengedEpochs() []uint64 {
	// The challenged epochs are represented as 1s in the bitmaps in the
	// participationCallData. The words are concatenated into a single bit
	// stream starting at startEpoch
	challengedEpochs := []uint64{}
	for wordIndex, participationCallData := range c.participationCallData {
		wordOffset := uint64(wordIndex) * bitsPerParticipationWord
		for i := 0; i < participationCallData.BitLen(); i++ {
			if participationCallData.Bit(i) == 1 {
				challengedEpochs = append(challengedEpochs, c.startEpoch+wordOffset+uint64(i))
			}
		}
	}
	return challengedEpochs
}

// Create the performance challenge defense task
func newDefendChallengePerformance(c *cli.Command, logger log.ColorLogger) (*defendChallengePerformance, error) {

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
		maxFee = math.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = math.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = math.GweiToWei(priorityFeeGwei)
	}

	eventLogInterval, err := cfg.GetEventLogInterval()
	if err != nil {
		return nil, err
	}

	// Return task
	return &defendChallengePerformance{
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

// Check for performance challenges
func (t *defendChallengePerformance) run(state *state.NetworkStateIndex) error {
	// Check if Saturn 2 is deployed
	if !state.Saturn2Deployed {
		return nil
	}

	// Log
	t.log.Println("Checking for performance challenges ...")

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
	mp, err := megapool.NewMegapool(t.rp, megapoolAddress, opts)
	if err != nil {
		return err
	}

	challenges, blockTimestamp, err := t.discovery.discover(onchainPerformanceChallenges{rp: t.rp, interval: t.intervalSize}, megapoolAddress, state.ElBlockNumber)
	if err != nil {
		return err
	}
	for _, challenge := range challenges {
		if challenge.expired(blockTimestamp) {
			t.log.Printlnf("Challenge %s has passed its response deadline; finalising it.", challenge.challengeId)
			if err := t.finaliseChallenge(challenge); err != nil {
				t.log.Printlnf("error finalising performance challenge %s: %v", challenge.challengeId, err)
			}
			continue
		}
		// The contract prohibits a proposer from defeating its own challenge.
		if challenge.proposer == nodeAccount.Address {
			continue
		}
		if err := t.checkChallenge(challenge, mp, opts, state); err != nil {
			t.log.Printlnf("error checking performance challenge %s: %v", challenge.challengeId, err)
		}
	}

	// Return
	return nil

}

// checkChallenge attempts one defense for the entire list, using any listed validator.
func (t *defendChallengePerformance) checkChallenge(challenge megapoolPerformanceChallenge, mp megapool.Megapool, opts *bind.CallOpts, state *state.NetworkStateIndex) error {
	challengedEpochs := challenge.getChallengedEpochs()
	for _, validatorId := range challenge.validatorIds {
		pubkey, err := mp.GetValidatorPubkey(validatorId, opts)
		if err != nil {
			t.log.Printlnf("error getting pubkey for megapool validator %d: %v", validatorId, err)
			continue
		}
		beaconStatus, err := t.bc.GetValidatorStatus(pubkey, nil)
		if err != nil {
			t.log.Printlnf("error getting beacon status for megapool validator %d: %v", validatorId, err)
			continue
		}
		if !beaconStatus.Exists || beaconStatus.Index == "" {
			continue
		}
		index, err := strconv.ParseUint(beaconStatus.Index, 10, 64)
		if err != nil {
			t.log.Printlnf("error parsing beacon index for megapool validator %d: %v", validatorId, err)
			continue
		}
		defender := challengedValidator{validatorId: validatorId, pubkey: pubkey, index: index}
		// Only activation after the challenge start is accepted by the contract's
		// validator-proof response; an early withdrawal alone is not a defense.
		if beaconStatus.ActivationEpoch > challenge.startEpoch {
			return t.respondWithValidator(challenge, defender, state)
		}
		_, epoch, found, err := performance.FindFirstTimelyTargetVote(t.bc, state.BeaconConfig, []uint64{index}, challengedEpochs)
		if err != nil {
			t.log.Printlnf("error verifying participation for megapool validator %d: %v", validatorId, err)
			continue
		}
		if found {
			return t.defendChallenge(t.rp, challenge, defender, epoch)
		}
	}
	t.log.Printlnf("No defense found for performance challenge %s.", challenge.challengeId)
	return nil
}

// finaliseChallenge requests exits after the challenge's stored response deadline.
func (t *defendChallengePerformance) finaliseChallenge(challenge megapoolPerformanceChallenge) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := megapool.EstimateFinaliseChallengeGas(t.rp, challenge.challengeId, opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to finalise challenge %s: %w", challenge.challengeId, err)
	}
	gas := big.NewInt(int64(gasInfo.Safe))

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gasInfo.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Finalise the challenge
	txHash, err := megapool.FinaliseChallenge(t.rp, challenge.challengeId, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = transactions.PrintAndWaitForTransaction(t.cfg, txHash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully finalised performance challenge %s.", challenge.challengeId)

	// Return
	return nil
}

// respondWithValidator proves the defender activated after the challenge start.
func (t *defendChallengePerformance) respondWithValidator(challenge megapoolPerformanceChallenge, defender challengedValidator, state *state.NetworkStateIndex) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Creating a validator proof for megapool validator %d.", defender.validatorId)

	// Build a fresh validator proof against the head state to satisfy the
	// contract's slot recency requirement
	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, 0, t.w, state.BeaconConfig, defender.pubkey, nil)
	if err != nil {
		return fmt.Errorf("error creating the validator proof: %w", err)
	}

	gasInfo, err := megapool.EstimateRespondWithValidatorGas(t.rp, challenge.challengeId, defender.validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	gas := big.NewInt(int64(gasInfo.Safe))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gasInfo.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	t.log.Printlnf("Responding to challenge %s with a validator proof for validator %d.", challenge.challengeId, defender.validatorId)
	txHash, err := megapool.RespondWithValidator(t.rp, challenge.challengeId, defender.validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = transactions.PrintAndWaitForTransaction(t.cfg, txHash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully responded to the performance challenge for validator %d.", defender.validatorId)

	// Return
	return nil
}

// defendChallenge responds to a performance challenge with a participation
// proof of the defender's timely target vote in challengeEpoch.
func (t *defendChallengePerformance) defendChallenge(rp *rocketpool.RocketPool, challenge megapoolPerformanceChallenge, defender challengedValidator, challengeEpoch uint64) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Creating a participation proof that validator index %d made a timely target vote in epoch %d.", defender.index, challengeEpoch)

	proofs, err := services.GetParticipationProof(t.c, defender.index, defender.pubkey, challengeEpoch, challenge.startEpoch, challenge.participationCallData)
	if err != nil {
		return fmt.Errorf("error creating the participation proof: %w", err)
	}

	gasInfo, err := megapool.EstimateRespondWithParticipationGas(rp, challenge.challengeId, defender.validatorId, proofs.Offset, proofs.ChallengeLeaf, proofs.ChallengeWitness, proofs.SlotTimestamp, proofs.Validator, proofs.Participation, proofs.Slot, opts)
	if err != nil {
		return err
	}

	gas := big.NewInt(int64(gasInfo.Safe))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gasInfo.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	t.log.Printlnf("Responding to challenge %s with the timely target vote of validator %d in epoch %d.", challenge.challengeId, defender.validatorId, challengeEpoch)
	txHash, err := megapool.RespondWithParticipation(rp, challenge.challengeId, defender.validatorId, proofs.Offset, proofs.ChallengeLeaf, proofs.ChallengeWitness, proofs.SlotTimestamp, proofs.Validator, proofs.Participation, proofs.Slot, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = transactions.PrintAndWaitForTransaction(t.cfg, txHash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully responded to the performance challenge for validator %d.", defender.validatorId)

	// Return
	return nil
}
