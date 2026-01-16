package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

type defendableProposal struct {
	challengeEvent *protocol.ChallengeSubmitted
	proposal       *protocol.ProtocolDaoProposalDetails
}

type defendPdaoProps struct {
	c                *cli.Context
	log              *log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                wallet.Wallet
	rp               *rocketpool.RocketPool
	bc               beacon.Client
	gasThreshold     float64
	maxFee           *big.Int
	maxPriorityFee   *big.Int
	gasLimit         uint64
	nodeAddress      common.Address
	propMgr          *proposals.ProposalManager
	lastScannedBlock *big.Int

	//Smart Node parameters
	intervalSize *big.Int
}

func newDefendPdaoProps(c *cli.Context, logger log.ColorLogger) (*defendPdaoProps, error) {
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

	// Get the event interval size
	intervalSize := big.NewInt(int64(cfg.Geth.EventLogInterval))

	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Make a proposal manager
	propMgr, err := proposals.NewProposalManager(&logger, cfg, rp, bc)

	// Return task
	return &defendPdaoProps{
		c:                c,
		log:              &logger,
		cfg:              cfg,
		w:                w,
		rp:               rp,
		bc:               bc,
		gasThreshold:     gasThreshold,
		maxFee:           maxFee,
		maxPriorityFee:   priorityFee,
		gasLimit:         0,
		nodeAddress:      account.Address,
		propMgr:          propMgr,
		lastScannedBlock: nil,

		intervalSize: intervalSize,
	}, nil
}

// Defend pDAO proposals
func (t *defendPdaoProps) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for Protocol DAO proposal challenges to defend...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get any proposals that need to be defended
	defendableProps, err := t.getDefendableProposals(state, opts)
	if err != nil {
		return fmt.Errorf("error checking for defendable proposals: %w", err)
	}
	if len(defendableProps) == 0 {
		return nil
	}

	// Defend props
	for _, prop := range defendableProps {
		err := t.defendProposal(prop)
		if err != nil {
			return fmt.Errorf("error submitting response for proposal %d, challenged index %d: %w", prop.proposal.ID, prop.challengeEvent.Index.Uint64(), err)
		}
	}
	t.lastScannedBlock = big.NewInt(int64(state.ElBlockNumber))

	return nil
}

// Get a list of this node's proposals with open challenges against them
func (t *defendPdaoProps) getDefendableProposals(state *state.NetworkState, opts *bind.CallOpts) ([]defendableProposal, error) {
	// Get proposals made by this node that are still in the challenge phase (Pending)
	eligibleProps := []protocol.ProtocolDaoProposalDetails{}
	for _, prop := range state.ProtocolDaoProposalDetails {
		if prop.ProposerAddress == t.nodeAddress &&
			prop.State == types.ProtocolDaoProposalState_Pending {
			eligibleProps = append(eligibleProps, prop)
		}
	}
	if len(eligibleProps) == 0 {
		return nil, nil
	}

	// Get the window of blocks to scan from
	var startBlock *big.Int
	endBlock := big.NewInt(int64(state.ElBlockNumber))
	if t.lastScannedBlock == nil {
		// Get the slot number the first proposal was created on
		startTime := eligibleProps[0].CreatedTime
		genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
		secondsPerSlot := time.Second * time.Duration(state.BeaconConfig.SecondsPerSlot)
		startSlot := uint64(startTime.Sub(genesisTime) / secondsPerSlot)

		// Get the Beacon block for the slot
		block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(startSlot))
		if err != nil {
			return nil, fmt.Errorf("error getting Beacon block at slot %d: %w", startSlot, err)
		}
		if !exists {
			return nil, fmt.Errorf("beacon block at slot %d was missing", startSlot)
		}

		// Get the EL block for this slot
		startBlock = big.NewInt(int64(block.ExecutionBlockNumber))
	} else {
		startBlock = big.NewInt(0).Add(t.lastScannedBlock, common.Big1)
	}

	// Make containers for eligible IDs
	ids := make([]uint64, len(eligibleProps))
	propMap := map[uint64]*protocol.ProtocolDaoProposalDetails{}
	for i, prop := range eligibleProps {
		ids[i] = prop.ID
		propMap[prop.ID] = &eligibleProps[i]
	}

	// Get the RocketRewardsPool addresses
	verifierAddresses := t.cfg.Smartnode.GetPreviousRocketDAOProtocolVerifierAddresses()

	// Get any challenges issued for the proposals
	challengeEvents, err := protocol.GetChallengeSubmittedEvents(t.rp, ids, t.intervalSize, startBlock, endBlock, verifierAddresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error scanning for ChallengeSubmitted events: %w", err)
	}

	// Parse them out
	defendableProposals := []defendableProposal{}
	for _, event := range challengeEvents {
		// Check if the challenge has been handled yet
		propID := event.ProposalID.Uint64()
		index := event.Index.Uint64()
		state, err := protocol.GetChallengeState(t.rp, propID, index, opts)
		if err != nil {
			return nil, fmt.Errorf("error checking state of challenge on proposal %d, index %d: %w", propID, index, err)
		}
		if state == types.ChallengeState_Challenged {
			t.log.Printlnf("Proposal %d, index %d has been challenged by %s.", propID, index, event.Challenger.Hex())
			defendableProposals = append(defendableProposals, defendableProposal{
				challengeEvent: &event,
				proposal:       propMap[propID],
			})
		}
	}

	return defendableProposals, nil
}

// Submit a response to a challenge against one of this node's proposals
func (t *defendPdaoProps) defendProposal(prop defendableProposal) error {
	propID := prop.proposal.ID
	challengedIndex := prop.challengeEvent.Index.Uint64()
	t.log.Printlnf("Responding to challenge against proposal %d, index %d...", propID, challengedIndex)

	// Create the response pollard
	_, pollard, err := t.propMgr.GetArtifactsForChallengeResponse(prop.proposal.TargetBlock, challengedIndex)
	if err != nil {
		return fmt.Errorf("error getting pollard for response to challenge against proposal %d, index %d: %w", propID, challengedIndex, err)
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := protocol.EstimateSubmitRootGas(t.rp, propID, challengedIndex, pollard, opts)
	if err != nil {
		return fmt.Errorf("error estimating the gas required to respond to challenge against proposal %d, index %d: %w", propID, challengedIndex, err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		t.log.Println("NOTICE: Challenge responses bypass the automatic TX gas threshold, responding for safety.")
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Respond to the challenge
	hash, err := protocol.SubmitRoot(t.rp, propID, challengedIndex, pollard, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully responded to challenge.")

	// Return
	return nil
}
