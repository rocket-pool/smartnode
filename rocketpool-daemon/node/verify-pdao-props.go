package node

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

type challenge struct {
	proposal        *protocol.ProtocolDaoProposal
	challengedIndex uint64
	challengedNode  types.VotingTreeNode
	witness         []types.VotingTreeNode
}

type defeat struct {
	proposal        *protocol.ProtocolDaoProposal
	challengedIndex uint64
}

type VerifyPdaoProps struct {
	ctx                 context.Context
	sp                  *services.ServiceProvider
	logger              *slog.Logger
	cfg                 *config.SmartNodeConfig
	w                   *wallet.Wallet
	rp                  *rocketpool.RocketPool
	bc                  beacon.IBeaconClient
	gasThreshold        float64
	maxFee              *big.Int
	maxPriorityFee      *big.Int
	nodeAddress         common.Address
	propMgr             *proposals.ProposalManager
	pdaoMgr             *protocol.ProtocolDaoManager
	lastScannedBlock    *big.Int
	validPropCache      map[uint64]bool
	rootSubmissionCache map[uint64]map[uint64]*protocol.RootSubmitted

	// Smartnode parameters
	intervalSize *big.Int
}

func NewVerifyPdaoProps(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger) *VerifyPdaoProps {
	cfg := sp.GetConfig()
	log := logger.With(slog.String(keys.TaskKey, "Verify PDAO Proposals"))
	maxFee, maxPriorityFee := getAutoTxInfo(cfg, log)
	return &VerifyPdaoProps{
		ctx:                 ctx,
		sp:                  sp,
		logger:              log,
		cfg:                 cfg,
		w:                   sp.GetWallet(),
		rp:                  sp.GetRocketPool(),
		bc:                  sp.GetBeaconClient(),
		gasThreshold:        cfg.AutoTxGasThreshold.Value,
		maxFee:              maxFee,
		maxPriorityFee:      maxPriorityFee,
		lastScannedBlock:    nil,
		intervalSize:        big.NewInt(int64(config.EventLogInterval)),
		validPropCache:      map[uint64]bool{},
		rootSubmissionCache: map[uint64]map[uint64]*protocol.RootSubmitted{},
	}
}

// Verify pDAO proposals
func (t *VerifyPdaoProps) Run(state *state.NetworkState) error {
	// Bindings
	t.nodeAddress, _ = t.w.GetAddress()
	propMgr, err := proposals.NewProposalManager(t.ctx, t.logger, t.cfg, t.rp, t.bc)
	if err != nil {
		return fmt.Errorf("error creating proposal manager: %w", err)
	}
	t.propMgr = propMgr
	pdaoMgr, err := protocol.NewProtocolDaoManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating Protocol DAO manager: %w", err)
	}
	t.pdaoMgr = pdaoMgr

	// Log
	t.logger.Info("Starting check for Protocol DAO proposals to challenge.")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get any challenges that need to be submitted
	challenges, defeats, err := t.getChallengesandDefeats(state, opts)
	if err != nil {
		return fmt.Errorf("error checking for challenges or defeats: %w", err)
	}

	// Create challenges
	submissions := []*eth.TransactionSubmission{}
	for _, challenge := range challenges {
		submission, err := t.createSubmitChallengeTx(challenge)
		if err != nil {
			return fmt.Errorf("error creating challenge against proposal %d, index %d: %w", challenge.proposal.ID, challenge.challengedIndex, err)
		}
		submissions = append(submissions, submission)
	}

	// Create defeats
	for _, defeat := range defeats {
		submission, err := t.createSubmitDefeatTx(defeat)
		if err != nil {
			return fmt.Errorf("error creating TX to defeat of proposal %d, index %d: %w", defeat.proposal.ID, defeat.challengedIndex, err)
		}
		submissions = append(submissions, submission)
	}

	// Submit transactions
	if len(submissions) == 0 {
		return nil
	}
	err = t.submitTxs(submissions)
	if err != nil {
		return fmt.Errorf("error submitting transactions: %w", err)
	}

	t.lastScannedBlock = big.NewInt(int64(state.ElBlockNumber))
	return nil
}

func (t *VerifyPdaoProps) getChallengesandDefeats(state *state.NetworkState, opts *bind.CallOpts) ([]challenge, []defeat, error) {
	// Get proposals *not* made by this node that are still in the challenge phase (Pending)
	eligibleProps := []*protocol.ProtocolDaoProposal{}
	for _, prop := range state.ProtocolDaoProposalDetails {
		if prop.State.Formatted() == types.ProtocolDaoProposalState_Pending &&
			prop.ProposerAddress.Get() != t.nodeAddress {
			eligibleProps = append(eligibleProps, prop)
		} else {
			// Remove old proposals from the caches once they're out of scope
			delete(t.validPropCache, prop.ID)
			delete(t.rootSubmissionCache, prop.ID)
		}
	}
	if len(eligibleProps) == 0 {
		return nil, nil, nil
	}

	// Check which ones have a root hash mismatch and need to be processed further
	mismatchingProps := []*protocol.ProtocolDaoProposal{}
	for _, prop := range eligibleProps {
		if t.validPropCache[prop.ID] {
			// Ignore proposals that have already been cleared
			continue
		}

		// Get the proposal's network tree root
		prop, err := protocol.NewProtocolDaoProposal(t.rp, prop.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating binding for proposal %d: %w", prop.ID, err)
		}
		var propRoot func() types.VotingTreeNode
		err = t.rp.Query(func(mc *batch.MultiCaller) error {
			prop.GetTreeNode(mc, 1)
			return nil
		}, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting root node for proposal %d: %w", prop.ID, err)
		}

		// Get the local tree
		networkTree, err := t.propMgr.GetNetworkTree(prop.TargetBlock.Formatted(), nil)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting network tree for proposal %d: %w", prop.ID, err)
		}
		localRoot := networkTree.Nodes[0]

		// Compare
		if propRoot().Sum.Cmp(localRoot.Sum) == 0 && propRoot().Hash == localRoot.Hash {
			t.logger.Info("Proposal matches the local tree artifacts, so it does not need to be challenged.", slog.Uint64(keys.ProposalKey, prop.ID))
			t.validPropCache[prop.ID] = true
			continue
		}

		// This proposal has a mismatch and must be challenged
		t.logger.Info("Proposal does not match the local tree artifacts and must be challenged.", slog.Uint64(keys.ProposalKey, prop.ID))
		mismatchingProps = append(mismatchingProps, prop)
	}
	if len(mismatchingProps) == 0 {
		return nil, nil, nil
	}

	// Get the window of blocks to scan from
	var startBlock *big.Int
	endBlock := big.NewInt(int64(state.ElBlockNumber))
	if t.lastScannedBlock == nil {
		// Get the slot number the first proposal was created on
		startTime := mismatchingProps[0].CreatedTime.Formatted()
		genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
		secondsPerSlot := time.Second * time.Duration(state.BeaconConfig.SecondsPerSlot)
		startSlot := uint64(startTime.Sub(genesisTime) / secondsPerSlot)

		// Get the Beacon block for the slot
		block, exists, err := t.bc.GetBeaconBlock(t.ctx, fmt.Sprint(startSlot))
		if err != nil {
			return nil, nil, fmt.Errorf("error getting Beacon block at slot %d: %w", startSlot, err)
		}
		if !exists {
			return nil, nil, fmt.Errorf("Beacon block at slot %d was missing", startSlot)
		}

		// Get the EL block for this slot
		startBlock = big.NewInt(int64(block.ExecutionBlockNumber))
	} else {
		startBlock = big.NewInt(0).Add(t.lastScannedBlock, common.Big1)
	}

	// Make containers for mismatching IDs
	ids := make([]uint64, len(mismatchingProps))
	propMap := map[uint64]*protocol.ProtocolDaoProposal{}
	for i, prop := range mismatchingProps {
		ids[i] = prop.ID
		propMap[prop.ID] = mismatchingProps[i]
	}

	// Get and cache all root submissions for the proposals
	rs := t.cfg.GetRocketPoolResources()
	rootSubmissionEvents, err := t.pdaoMgr.GetRootSubmittedEvents(ids, t.intervalSize, startBlock, endBlock, rs.PreviousProtocolDaoVerifierAddresses, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error scanning for RootSubmitted events: %w", err)
	}
	for _, event := range rootSubmissionEvents {
		// Add them to the cache
		propID := event.ProposalID.Uint64()
		rootIndex := event.Index.Uint64()
		eventsForProp, exists := t.rootSubmissionCache[propID]
		if !exists {
			eventsForProp = map[uint64]*protocol.RootSubmitted{}
		}
		eventsForProp[rootIndex] = &event
		t.rootSubmissionCache[propID] = eventsForProp
	}

	// For each proposal, crawl down the tree looking at mismatched indices to challenge until arriving at one that hasn't been challenged yet
	challenges := []challenge{}
	defeats := []defeat{}
	for _, prop := range mismatchingProps {
		challenge, defeat, err := t.getChallengeOrDefeatForProposal(prop, opts)
		if err != nil {
			return nil, nil, err
		}
		if challenge != nil {
			challenges = append(challenges, *challenge)
		}
		if defeat != nil {
			defeats = append(defeats, *defeat)
		}
	}

	return challenges, defeats, nil
}

// Get the challenge against a proposal if one can be found
func (t *VerifyPdaoProps) getChallengeOrDefeatForProposal(prop *protocol.ProtocolDaoProposal, opts *bind.CallOpts) (*challenge, *defeat, error) {
	challengedIndex := uint64(1) // Root

	for {
		// Get the index of the node to challenge
		rootSubmissionEvent, exists := t.rootSubmissionCache[prop.ID][challengedIndex]
		if !exists {
			return nil, nil, fmt.Errorf("challenge against prop %d, index %d has been responded to but the RootSubmitted event was missing", prop.ID, challengedIndex)
		}
		newChallengedIndex, challengedNode, proof, err := t.propMgr.CheckForChallengeableArtifacts(*rootSubmissionEvent)
		if err != nil {
			return nil, nil, fmt.Errorf("error checking for challengeable artifacts on prop %d, index %s: %w", prop.ID, rootSubmissionEvent.Index.String(), err)
		}
		if newChallengedIndex == 0 {
			// Do nothing if the prop can't be challenged
			t.logger.Info("Check showed no challengeable artifacts.", slog.Uint64(keys.ProposalKey, prop.ID), slog.Uint64(keys.IndexKey, challengedIndex))
			return nil, nil, nil
		}
		if newChallengedIndex == challengedIndex {
			// This shouldn't ever happen but it does then error out for safety
			return nil, nil, fmt.Errorf("cycle error: proposal %d had index %d challenged, and the new challengeable artifacts had the same index", prop.ID, challengedIndex)
		}

		// Check if the index has been challenged yet
		var getState func() types.ChallengeState
		err = t.rp.Query(func(mc *batch.MultiCaller) error {
			getState = prop.GetChallengeState(mc, newChallengedIndex)
			return nil
		}, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("error checking challenge state for proposal %d, index %d: %w", prop.ID, challengedIndex, err)
		}
		state := getState()
		switch state {
		case types.ChallengeState_Unchallenged:
			// If it's unchallenged, this is the index to challenge
			return &challenge{
				proposal:        prop,
				challengedIndex: newChallengedIndex,
				challengedNode:  challengedNode,
				witness:         proof,
			}, nil, nil
		case types.ChallengeState_Challenged:
			// Check if the proposal can be defeated
			if time.Since(prop.CreatedTime.Formatted().Add(prop.ChallengeWindow.Formatted())) > 0 {
				return nil, &defeat{
					proposal:        prop,
					challengedIndex: newChallengedIndex,
				}, nil
			}
			// Nothing to do but wait for the proposer to respond
			t.logger.Info("Proposal has already been challenged; waiting for proposer to respond.", slog.Uint64(keys.ProposalKey, prop.ID), slog.Uint64(keys.IndexKey, newChallengedIndex))
			return nil, nil, nil
		case types.ChallengeState_Responded:
			// Delve deeper into the tree looking for the next index to challenge
			challengedIndex = newChallengedIndex
		default:
			return nil, nil, fmt.Errorf("unexpected state '%d' for challenge against proposal %d, index %d", state, prop.ID, challengedIndex)
		}
	}
}

// Submit a challenge against a proposal
func (t *VerifyPdaoProps) createSubmitChallengeTx(challenge challenge) (*eth.TransactionSubmission, error) {
	prop := challenge.proposal
	challengedIndex := challenge.challengedIndex
	t.logger.Info("Creating challenge...", slog.Uint64(keys.ProposalKey, prop.ID), slog.Uint64(keys.IndexKey, challengedIndex))

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	txInfo, err := prop.CreateChallenge(challengedIndex, challenge.challengedNode, challenge.witness, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating the gas required to submit challenge against proposal %d, index %d: %w", prop.ID, challengedIndex, err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return nil, fmt.Errorf("simulating challenge against proposal %d, index %d failed: %s", prop.ID, challengedIndex, txInfo.SimulationResult.SimulationError)
	}

	submission, err := eth.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating submission to challenge against proposal %d, index %d: %w", prop.ID, challengedIndex, err)
	}
	return submission, nil
}

// Defeat a proposal
func (t *VerifyPdaoProps) createSubmitDefeatTx(defeat defeat) (*eth.TransactionSubmission, error) {
	prop := defeat.proposal
	challengedIndex := defeat.challengedIndex
	t.logger.Info("Proposal has been defeated, creating defeat TX...", slog.Uint64(keys.ProposalKey, prop.ID), slog.Uint64(keys.IndexKey, challengedIndex))

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	txInfo, err := prop.Defeat(challengedIndex, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating TX to defeat proposal %d with index %d: %w", prop.ID, challengedIndex, err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return nil, fmt.Errorf("simulating defeat of proposal %d with index %d failed: %s", prop.ID, challengedIndex, txInfo.SimulationResult.SimulationError)
	}

	submission, err := eth.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating submission to defeat proposal %d with index %d: %w", prop.ID, challengedIndex, err)
	}
	return submission, nil
}

// Submit all transactions
func (t *VerifyPdaoProps) submitTxs(submissions []*eth.TransactionSubmission) error {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.logger)
		if err != nil {
			return err
		}
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	if !gas.PrintAndCheckGasInfoForBatch(submissions, true, t.gasThreshold, t.logger, maxFee) {
		return nil
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransactionBatch(t.cfg, t.rp, t.logger, submissions, nil, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted all transactions.")
	return nil
}
