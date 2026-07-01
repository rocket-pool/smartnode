package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Stake megapool validator task
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
}

type megapoolPerformanceChallenge struct {
	megapoolAddress       common.Address
	validatorIds          []uint32
	startEpoch            uint64
	participationCallData []*big.Int
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

// Create stake megapool validator task
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
	}, nil

}

// Check for performance challenges
func (t *defendChallengePerformance) run(state *state.NetworkState) error {
	// Check if Saturn 2 is deployed
	if !state.Saturn2Deployed {
		t.log.Println("Saturn 2 is not deployed, skipping performance challenges check.")
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
	mp, err := megapool.NewMegaPoolV1(t.rp, megapoolAddress, nil)
	if err != nil {
		return err
	}

	participationCallData := []*big.Int{new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 200),
		big.NewInt(1)),
	}
	// Use a megapool challenge stub for now
	challenge := megapoolPerformanceChallenge{
		megapoolAddress:       megapoolAddress,
		validatorIds:          []uint32{0, 1, 2},
		participationCallData: participationCallData,
		startEpoch:            105000,
	}

	challengedEpochs := challenge.getChallengedEpochs()
	t.log.Printlnf("Challenged epochs: %v", challengedEpochs)

	// Resolve every challenged validator's pubkey and beacon-chain index up
	// front so the per-epoch beacon data can be fetched once and shared across
	// all of them when verifying target-vote participation.
	validatorsByIndex := make(map[uint64]challengedValidator, len(challenge.validatorIds))
	validatorIndices := make([]uint64, 0, len(challenge.validatorIds))
	for _, validatorId := range challenge.validatorIds {
		pubkey, err := mp.GetValidatorPubkey(validatorId, opts)
		if err != nil {
			t.log.Printlnf("error getting pubkey for megapool validator %d: %v", validatorId, err)
			continue
		}
		beaconStatus, err := t.bc.GetValidatorStatus(pubkey, nil)
		if err != nil {
			t.log.Printlnf("error getting beacon status for megapool validator %d (%s): %v", validatorId, pubkey.Hex(), err)
			continue
		}
		if !beaconStatus.Exists || beaconStatus.Index == "" {
			t.log.Printlnf("Megapool validator %d (%s) is not on the beacon chain yet, skipping.", validatorId, pubkey.Hex())
			continue
		}
		validatorIndex, err := strconv.ParseUint(beaconStatus.Index, 10, 64)
		if err != nil {
			t.log.Printlnf("error parsing beacon index %q for megapool validator %d: %v", beaconStatus.Index, validatorId, err)
			continue
		}
		validatorsByIndex[validatorIndex] = challengedValidator{
			validatorId: validatorId,
			pubkey:      pubkey,
			index:       validatorIndex,
		}
		validatorIndices = append(validatorIndices, validatorIndex)
	}

	// Find the first challenged validator that made a successful target vote
	// within the challenged range. A single (validator, epoch) proof is enough
	// to defend the challenge
	validatorIndex, epoch, found, err := performance.FindFirstTimelyTargetVote(t.bc, state.BeaconConfig, validatorIndices, challengedEpochs)
	if err != nil {
		return fmt.Errorf("error verifying target-vote participation for challenged megapool validators: %w", err)
	}
	if !found {
		t.log.Println("No challenged validator made a successful target vote in the challenged epochs.")
		return nil
	}

	// Defend the challenge using that validator id and epoch.
	defender := validatorsByIndex[validatorIndex]
	t.log.Printlnf("Megapool validator %d made a timely target vote in epoch %d; defending the performance challenge.", defender.validatorId, epoch)
	if err := t.defendChallenge(t.rp, mp, defender.validatorId, state, defender.pubkey, epoch, opts); err != nil {
		t.log.Printlnf("error defending performance challenge for megapool validator %d: %v", defender.validatorId, err)
	}

	// Return
	return nil

}

func (t *defendChallengePerformance) defendChallenge(rp *rocketpool.RocketPool, mp megapool.Megapool, validatorId uint32, state *state.NetworkState, validatorPubkey types.ValidatorPubkey, challengeEpoch uint64, callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Creating a validator performance proof that validator id %d participated in the epoch %v.", validatorId, challengeEpoch)

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, 0, t.w, state.BeaconConfig, mp.GetAddress(), validatorPubkey, nil)
	if err != nil {
		t.log.Printlnf("[ERROR] There was an error during the proof creation process: %w", err)
		return err
	}

	t.log.Printlnf("The validator performance proof has been successfully created.")
	var gasInfo rocketpool.GasInfo

	gasInfo, err = megapool.EstimateNotifyExitGas(rp, mp.GetAddress(), validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
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

	var tx *coretypes.Transaction

	t.log.Printlnf("Notifying that validator %d is exiting.", validatorId)
	tx, err = megapool.NotifyExit(rp, mp.GetAddress(), validatorId, slotTimestamp, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully responded the performance challenge for validator %d.", validatorId)

	// Return
	return nil
}
