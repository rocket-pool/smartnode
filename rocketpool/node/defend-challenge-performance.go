package node

import (
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

// Defend this node's validators and optionally enforce community challenges.
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
	bondDiscovery  challengeBondDiscovery
}

type megapoolPerformanceChallenge struct {
	challengeId           *big.Int
	megapoolAddress       common.Address
	nodeAddress           common.Address
	minipoolAddresses     []common.Address
	validatorIds          []uint32
	startEpoch            uint64
	participationCallData []*big.Int
	responseDeadline      *big.Int
	proposer              common.Address
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

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}
	// Settlement applies to challenges of either pool type, including those
	// against other nodes. It must run even when this node has no megapool.
	if err := t.settleChallengeBonds(nodeAccount.Address, state.ElBlockNumber); err != nil {
		t.log.Printlnf("error checking performance challenge bonds: %v", err)
	}

	ownMegapool := common.Address{}
	if nodeDetails, ok := state.NodeDetailsByAddress[nodeAccount.Address]; ok && nodeDetails.MegapoolDeployed {
		ownMegapool = nodeDetails.MegapoolAddress
	}
	challenges, timestamp, err := t.discovery.discover(onchainPerformanceChallenges{rp: t.rp, interval: t.intervalSize}, common.Address{}, state.ElBlockNumber)
	if err != nil {
		return err
	}
	enforcer := t.cfg.Smartnode.EnableEnforcerTasks.Value.(bool)
	for _, challenge := range challenges {
		own := challenge.nodeAddress == nodeAccount.Address || (ownMegapool != (common.Address{}) && challenge.megapoolAddress == ownMegapool)
		if !shouldHandlePerformanceChallenge(own, challenge.proposer == nodeAccount.Address, enforcer, challenge.expired(timestamp)) {
			continue
		}
		if len(challenge.minipoolAddresses) > 0 {
			t.log.Printlnf("Challenge %s: node %s, minipools %v.", challenge.challengeId, challenge.nodeAddress, challenge.minipoolAddresses)
		} else {
			t.log.Printlnf("Challenge %s: megapool %s, validator IDs %v.", challenge.challengeId, challenge.megapoolAddress, challenge.validatorIds)
		}
		t.log.Printlnf("Challenge %s: proposer %s, start epoch %d, %d claimed missed epochs, response deadline %s.", challenge.challengeId, challenge.proposer, challenge.startEpoch, len(challenge.getChallengedEpochs()), time.Unix(challenge.responseDeadline.Int64(), 0).UTC().Format(time.RFC3339))
		if challenge.expired(timestamp) {
			t.log.Printlnf("Challenge %s: response deadline has passed; preparing finalisation.", challenge.challengeId)
			if err := t.finaliseChallenge(challenge); err != nil {
				t.log.Printlnf("error finalising challenge %s: %v", challenge.challengeId, err)
			}
			continue
		}
		if challenge.proposer == nodeAccount.Address {
			continue
		}
		// Recheck live status before potentially expensive proofs; final submission
		// is simulated again to catch responses mined before submission.
		status, err := megapool.GetPerformanceChallengeStatus(t.rp, challenge.challengeId, nil)
		if err != nil {
			t.log.Printlnf("error reading challenge %s: %v", challenge.challengeId, err)
			continue
		}
		if !status.CanDefend(nodeAccount.Address, timestamp) {
			t.log.Printlnf("Challenge %s: no longer eligible for defense; skipping.", challenge.challengeId)
			continue
		}
		t.log.Printlnf("Challenge %s: building a defense proof.", challenge.challengeId)
		defense, err := services.BuildPerformanceChallengeDefense(t.c, challenge.binding())
		if err != nil {
			t.log.Printlnf("challenge %s: %v", challenge.challengeId, err)
			continue
		}
		if err := t.submitDefense(challenge, defense); err != nil {
			t.log.Printlnf("error defending challenge %s: %v", challenge.challengeId, err)
		}
	}
	return nil
}

func shouldHandlePerformanceChallenge(own, proposed, enforcer, expired bool) bool {
	if expired {
		return own || proposed || enforcer
	}
	return !proposed && (own || enforcer)
}

func (c megapoolPerformanceChallenge) binding() megapool.PerformanceChallenge {
	return megapool.PerformanceChallenge{
		ChallengeId:       c.challengeId,
		MegapoolAddress:   c.megapoolAddress,
		NodeAddress:       c.nodeAddress,
		MinipoolAddresses: c.minipoolAddresses,
		ValidatorIds:      c.validatorIds,
		StartEpoch:        c.startEpoch,
		Participation:     c.participationCallData,
	}
}

// finaliseChallenge requests exits after the challenge's stored response deadline.
func (t *defendChallengePerformance) finaliseChallenge(challenge megapoolPerformanceChallenge) error {
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}
	return t.submitChallengeTransaction(challenge.challengeId, "finalisation", opts, func() (gaslimit.Limits, error) {
		return megapool.EstimateFinaliseChallengeGas(t.rp, challenge.challengeId, opts)
	}, func() (common.Hash, error) {
		return megapool.FinaliseChallenge(t.rp, challenge.challengeId, opts)
	})
}

func (t *defendChallengePerformance) submitDefense(challenge megapoolPerformanceChallenge, defense megapool.PerformanceChallengeDefense) error {
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}
	if defense.MinipoolAddress != (common.Address{}) {
		t.log.Printlnf("Challenge %s: using minipool %s, beacon validator index %s.", defense.ChallengeId, defense.MinipoolAddress, defense.Validator.ValidatorIndex)
	} else {
		t.log.Printlnf("Challenge %s: using megapool validator %d, beacon validator index %s.", defense.ChallengeId, defense.ValidatorId, defense.Validator.ValidatorIndex)
	}
	if defense.Participation != nil {
		t.log.Printlnf("Challenge %s: prepared a timely target participation proof for epoch %d, anchored at beacon slot %d, to defeat the entire challenge.", defense.ChallengeId, challenge.startEpoch+defense.Offset, defense.Slot.Slot)
	} else {
		t.log.Printlnf("Challenge %s: prepared an activation proof with activation epoch %d after start epoch %d, anchored at beacon slot %d, to defeat the entire challenge.", defense.ChallengeId, defense.Validator.Validator.ActivationEpoch, challenge.startEpoch, defense.Slot.Slot)
	}
	return t.submitChallengeTransaction(defense.ChallengeId, "defense", opts, func() (gaslimit.Limits, error) {
		return defense.EstimateGas(t.rp, opts)
	}, func() (common.Hash, error) {
		return defense.Submit(t.rp, opts)
	})
}

func (t *defendChallengePerformance) submitChallengeTransaction(id *big.Int, action string, opts *bind.TransactOpts, estimate func() (gaslimit.Limits, error), submit func() (common.Hash, error)) error {
	opts.Value = nil
	limits, err := estimate()
	if err != nil {
		return err
	}
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Sign() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}
	if !limits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		t.log.Printlnf("Challenge %s: deferring %s because the gas price exceeds the configured threshold.", id, action)
		return nil
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = limits.Safe
	if t.gasLimit != 0 {
		opts.GasLimit = t.gasLimit
	}
	t.log.Printlnf("Challenge %s: submitting %s.", id, action)
	hash, err := submit()
	if err != nil {
		return err
	}
	if err := transactions.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log); err != nil {
		return err
	}
	t.log.Printlnf("Challenge %s: successfully completed %s.", id, action)
	return nil
}
