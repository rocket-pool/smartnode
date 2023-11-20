package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

type verifyPdaoProps struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
	nodeAddress    common.Address

	// Smartnode parameters
	intervalSize *big.Int

	// Beacon parameters
	genesisTime    time.Time
	secondsPerSlot time.Duration

	// Local parameters
	//proposalEventStartBlockMap map[uint64]*big.Int
}

func newVerifyPdaoProps(c *cli.Context, logger log.ColorLogger) (*verifyPdaoProps, error) {
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
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Get the event interval size
	intervalSize := big.NewInt(int64(cfg.Geth.EventLogInterval))

	// Get the genesis time from the BC
	beaconCfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Return task
	return &verifyPdaoProps{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
		nodeAddress:    account.Address,

		intervalSize:   intervalSize,
		genesisTime:    time.Unix(int64(beaconCfg.GenesisTime), 0),
		secondsPerSlot: time.Duration(beaconCfg.SecondsPerSlot) * time.Second,
		//proposalEventStartBlockMap: map[uint64]*big.Int{},
	}, nil
}

/*
// Verify pDAO proposals
func (t *verifyPdaoProps) run(state *state.NetworkState) error {

	// Check for Houston
	isHoustonDeployed, err := state.IsHoustonDeployed(rp, nil)
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !isHoustonDeployed {
		return nil
	}

	// Log
	t.log.Println("Checking for Protocol DAO proposals to verify...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	return nil
}

func (t *verifyPdaoProps) getChallengeableProposals(state *state.NetworkState) ([]dao.ProposalDetails, error) {
	// Get the proposals that are currently in the challenge window
	candidates, err := t.getProposalsInChallengeWindow(state)
	if err != nil {
		return nil, fmt.Errorf("error checking for candidate proposals: %w", err)
	}

	// Get the block and pollard for the proposal

}
*/

/*
func (t *verifyPdaoProps) checkDutiesForProposal(proposalDetails dao.ProposalDetails, headBlock uint64) error {
	// Get the block to start scanning for new events
	startBlock, exists := t.proposalEventStartBlockMap[proposalDetails.ID]
	if !exists {
	}

	// Determine the start block for the even scan window based on the time the proposal was created
	createTime := time.Unix(int64(proposalDetails.CreatedTime), 0)
	timeSinceGenesis := createTime.Sub(t.genesisTime)
	slot := uint64(timeSinceGenesis / t.secondsPerSlot)
	block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(slot))
	if err != nil {
		return fmt.Errorf("error getting creation block for proposal %d: error getting beacon block %d: %w", proposalDetails.ID, slot, err)
	}
	if !exists {
		return fmt.Errorf("error getting creation block for proposal %d: beacon block %d does not exist", proposalDetails.ID, slot)
	}

	startBlock := big.NewInt(int64(block.ExecutionBlockNumber))
	endBlock := big.NewInt(int64(headBlock))

	// Get the events for this proposal
	rootSubmittedEvents, err := voting.GetRootSubmittedEvents(t.rp, proposalDetails.ID, t.intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return fmt.Errorf("error getting root submitted events: %w", err)
	}
	challengeEvents, err := voting.GetChallengeSubmittedEvents(t.rp, proposalDetails.ID, t.intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return fmt.Errorf("error getting challenge submitted events: %w", err)
	}

	// Create lookups of events by index
	rseLookup := map[*big.Int]voting.RootSubmitted{}
	ceLookup := map[*big.Int]voting.ChallengeSubmitted{}
	for _, rse := range rootSubmittedEvents {
		rseLookup[rse.Index] = rse
	}
	for _, ce := range challengeEvents {
		ceLookup[ce.Index] = ce
	}

	// Check if this is the node's proposal
	if proposalDetails.ProposerAddress == t.nodeAddress {
		err = t.handleOwnProposal(proposalDetails, rseLookup, ceLookup)
		if err != nil {
			return fmt.Errorf("error handling own proposal %d: %w", proposalDetails.ID, err)
		}
	} else {
		// TODO: look and see if challenges need to happen
	}

	// Build the tree for the proposal

	// Update the start block cache for this proposal
	//t.proposalEventStartBlockMap[proposalDetails.ID] = big.NewInt(0).Add(endBlock, common.Big1)
	return nil
}

func (t *verifyPdaoProps) handleOwnProposal(details dao.ProposalDetails, rseLookup map[*big.Int]voting.RootSubmitted, ceLookup map[*big.Int]voting.ChallengeSubmitted) error {
	return nil
}

func (t *verifyPdaoProps) getProposalsInChallengeWindow(state *state.NetworkState) ([]dao.ProposalDetails, error) {
	// TODO
	return nil, nil
}
*/
