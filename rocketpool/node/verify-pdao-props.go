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
	"github.com/urfave/cli/v3"
)

type challenge struct {
	proposalID      uint64
	challengedIndex uint64
	challengedNode  types.VotingTreeNode
	witness         []types.VotingTreeNode
}

type defeat struct {
	proposalID      uint64
	challengedIndex uint64
}

// proposalNodeGetter retrieves the on-chain voting tree node submitted with a proposal.
type proposalNodeGetter interface {
	GetNode(proposalID uint64, index uint64) (types.VotingTreeNode, error)
}

// networkTreeProvider builds or loads the local voting tree for a target block.
type networkTreeProvider interface {
	GetNetworkTree(blockNumber uint32) (*proposals.NetworkVotingTree, error)
}

// challengeStateGetter checks the on-chain challenge state for a proposal index.
type challengeStateGetter interface {
	GetChallengeState(proposalID uint64, index uint64) (types.ChallengeState, error)
}

// rootSubmissionEventProvider fetches RootSubmitted events from the chain.
type rootSubmissionEventProvider interface {
	GetRootSubmittedEvents(proposalIDs []uint64, startBlock *big.Int, endBlock *big.Int) ([]protocol.RootSubmitted, error)
}

// challengeArtifactChecker checks a RootSubmitted event for challengeable artifacts.
type challengeArtifactChecker interface {
	CheckForChallengeableArtifacts(event protocol.RootSubmitted) (uint64, types.VotingTreeNode, []types.VotingTreeNode, error)
}

type verifyPdaoProps struct {
	c                   *cli.Command
	log                 *log.ColorLogger
	cfg                 *config.RocketPoolConfig
	w                   wallet.Wallet
	rp                  *rocketpool.RocketPool
	bc                  beacon.Client
	gasThreshold        float64
	maxFee              *big.Int
	maxPriorityFee      *big.Int
	gasLimit            uint64
	nodeAddress         common.Address
	propMgr             *proposals.ProposalManager
	lastScannedBlock    *big.Int
	validPropCache      map[uint64]bool
	rootSubmissionCache map[uint64]map[uint64]*protocol.RootSubmitted

	//Smart Node parameters
	intervalSize *big.Int
}

func newVerifyPdaoProps(c *cli.Command, logger log.ColorLogger) (*verifyPdaoProps, error) {
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
	return &verifyPdaoProps{
		c:                   c,
		log:                 &logger,
		cfg:                 cfg,
		w:                   w,
		rp:                  rp,
		bc:                  bc,
		gasThreshold:        gasThreshold,
		maxFee:              maxFee,
		maxPriorityFee:      priorityFee,
		gasLimit:            0,
		nodeAddress:         account.Address,
		propMgr:             propMgr,
		lastScannedBlock:    nil,
		validPropCache:      map[uint64]bool{},
		rootSubmissionCache: map[uint64]map[uint64]*protocol.RootSubmitted{},

		intervalSize: intervalSize,
	}, nil
}

// liveProposalNodeGetter uses on-chain calls to fetch proposal tree nodes.
type liveProposalNodeGetter struct {
	rp   *rocketpool.RocketPool
	opts *bind.CallOpts
}

func (g *liveProposalNodeGetter) GetNode(proposalID uint64, index uint64) (types.VotingTreeNode, error) {
	return protocol.GetNode(g.rp, proposalID, index, g.opts)
}

// liveNetworkTreeProvider delegates to the ProposalManager.
type liveNetworkTreeProvider struct {
	propMgr *proposals.ProposalManager
}

func (p *liveNetworkTreeProvider) GetNetworkTree(blockNumber uint32) (*proposals.NetworkVotingTree, error) {
	return p.propMgr.GetNetworkTree(blockNumber, nil)
}

// liveChallengeStateGetter reads challenge state from the chain.
type liveChallengeStateGetter struct {
	rp   *rocketpool.RocketPool
	opts *bind.CallOpts
}

func (g *liveChallengeStateGetter) GetChallengeState(proposalID uint64, index uint64) (types.ChallengeState, error) {
	return protocol.GetChallengeState(g.rp, proposalID, index, g.opts)
}

// liveRootSubmissionEventProvider fetches RootSubmitted events from the chain.
type liveRootSubmissionEventProvider struct {
	rp                *rocketpool.RocketPool
	intervalSize      *big.Int
	verifierAddresses []common.Address
	opts              *bind.CallOpts
}

func (p *liveRootSubmissionEventProvider) GetRootSubmittedEvents(proposalIDs []uint64, startBlock *big.Int, endBlock *big.Int) ([]protocol.RootSubmitted, error) {
	return protocol.GetRootSubmittedEvents(p.rp, proposalIDs, p.intervalSize, startBlock, endBlock, p.verifierAddresses, p.opts)
}

// liveChallengeArtifactChecker delegates to the ProposalManager.
type liveChallengeArtifactChecker struct {
	propMgr *proposals.ProposalManager
}

func (c *liveChallengeArtifactChecker) CheckForChallengeableArtifacts(event protocol.RootSubmitted) (uint64, types.VotingTreeNode, []types.VotingTreeNode, error) {
	return c.propMgr.CheckForChallengeableArtifacts(event)
}

// Verify pDAO proposals
func (t *verifyPdaoProps) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for Protocol DAO proposals to challenge...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get any challenges that need to be submitted
	challenges, defeats, err := t.getChallengesandDefeats(state, opts)
	if err != nil {
		return fmt.Errorf("error checking for challenges or defeats: %w", err)
	}

	// Submit challenges
	for _, challenge := range challenges {
		err := t.submitChallenge(challenge)
		if err != nil {
			return fmt.Errorf("error submitting challenge against proposal %d, index %d: %w", challenge.proposalID, challenge.challengedIndex, err)
		}
	}

	// Submit defeats
	for _, defeat := range defeats {
		err := t.submitDefeat(defeat)
		if err != nil {
			return fmt.Errorf("error submitting defeat of proposal %d, index %d: %w", defeat.proposalID, defeat.challengedIndex, err)
		}
	}

	t.lastScannedBlock = big.NewInt(int64(state.ElBlockNumber))
	return nil
}

func (t *verifyPdaoProps) getChallengesandDefeats(ns *state.NetworkState, opts *bind.CallOpts) ([]challenge, []defeat, error) {
	nodeGetter := &liveProposalNodeGetter{rp: t.rp, opts: opts}
	treeProvider := &liveNetworkTreeProvider{propMgr: t.propMgr}
	stateGetter := &liveChallengeStateGetter{rp: t.rp, opts: opts}
	verifierAddresses := t.cfg.Smartnode.GetPreviousRocketDAOProtocolVerifierAddresses()
	eventProvider := &liveRootSubmissionEventProvider{rp: t.rp, intervalSize: t.intervalSize, verifierAddresses: verifierAddresses, opts: opts}
	artifactChecker := &liveChallengeArtifactChecker{propMgr: t.propMgr}

	return getChallengesFromState(
		ns, t.nodeAddress, t.log, t.bc,
		nodeGetter, treeProvider, stateGetter, eventProvider, artifactChecker,
		t.validPropCache, t.rootSubmissionCache, t.lastScannedBlock,
	)
}

// getChallengesFromState computes which proposals need to be challenged or defeated.
// All chain dependencies are injected via interfaces so this can be tested with
// static network state and stub implementations.
func getChallengesFromState(
	ns *state.NetworkState,
	nodeAddress common.Address,
	log *log.ColorLogger,
	bc beacon.Client,
	nodeGetter proposalNodeGetter,
	treeProvider networkTreeProvider,
	stateGetter challengeStateGetter,
	eventProvider rootSubmissionEventProvider,
	artifactChecker challengeArtifactChecker,
	validPropCache map[uint64]bool,
	rootSubmissionCache map[uint64]map[uint64]*protocol.RootSubmitted,
	lastScannedBlock *big.Int,
) ([]challenge, []defeat, error) {
	eligibleProps := []protocol.ProtocolDaoProposalDetails{}
	for _, prop := range ns.ProtocolDaoProposalDetails {
		if prop.State == types.ProtocolDaoProposalState_Pending &&
			prop.ProposerAddress != nodeAddress {
			eligibleProps = append(eligibleProps, prop)
		} else {
			// Remove old proposals from the caches once they're out of scope
			delete(validPropCache, prop.ID)
			delete(rootSubmissionCache, prop.ID)
		}
	}
	if len(eligibleProps) == 0 {
		return nil, nil, nil
	}

	// Check which ones have a root hash mismatch and need to be processed further
	mismatchingProps := []protocol.ProtocolDaoProposalDetails{}
	for _, prop := range eligibleProps {
		// Ignore proposals that have already been cleared
		if validPropCache[prop.ID] {
			continue
		}

		// Get the proposal's network tree root
		propRoot, err := nodeGetter.GetNode(prop.ID, 1)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting root node for proposal %d: %w", prop.ID, err)
		}

		// Get the local tree
		networkTree, err := treeProvider.GetNetworkTree(prop.TargetBlock)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting network tree for proposal %d: %w", prop.ID, err)
		}
		localRoot := networkTree.Nodes[0]
		// Compare
		if propRoot.Sum.Cmp(localRoot.Sum) == 0 && propRoot.Hash == localRoot.Hash {
			log.Printlnf("Proposal %d matches the local tree artifacts, so it does not need to be challenged.", prop.ID)
			validPropCache[prop.ID] = true
			continue
		}

		// This proposal has a mismatch and must be challenged

		log.Printlnf("Proposal %d does not match the local tree artifacts and must be challenged.", prop.ID)
		mismatchingProps = append(mismatchingProps, prop)
	}
	if len(mismatchingProps) == 0 {
		return nil, nil, nil
	}

	// Get the window of blocks to scan from
	var startBlock *big.Int
	endBlock := big.NewInt(int64(ns.ElBlockNumber))
	if lastScannedBlock == nil {
		startTime := mismatchingProps[0].CreatedTime
		genesisTime := time.Unix(int64(ns.BeaconConfig.GenesisTime), 0)
		secondsPerSlot := time.Second * time.Duration(ns.BeaconConfig.SecondsPerSlot)
		startSlot := uint64(startTime.Sub(genesisTime) / secondsPerSlot)

		// Get the Beacon block for the slot
		block, exists, err := bc.GetBeaconBlock(fmt.Sprint(startSlot))
		if err != nil {
			return nil, nil, fmt.Errorf("error getting Beacon block at slot %d: %w", startSlot, err)
		}
		if !exists {
			return nil, nil, fmt.Errorf("Beacon block at slot %d was missing", startSlot)
		}

		// Get the EL block for this slot
		startBlock = big.NewInt(int64(block.ExecutionBlockNumber))
	} else {
		startBlock = big.NewInt(0).Add(lastScannedBlock, common.Big1)
	}

	// Make containers for mismatching IDs
	ids := make([]uint64, len(mismatchingProps))
	for i, prop := range mismatchingProps {
		ids[i] = prop.ID
	}

	rootSubmissionEvents, err := eventProvider.GetRootSubmittedEvents(ids, startBlock, endBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("error scanning for RootSubmitted events: %w", err)
	}
	for _, event := range rootSubmissionEvents {
		// Add them to the cache
		propID := event.ProposalID.Uint64()
		rootIndex := event.Index.Uint64()
		eventsForProp, exists := rootSubmissionCache[propID]
		if !exists {
			eventsForProp = map[uint64]*protocol.RootSubmitted{}
		}
		eventsForProp[rootIndex] = &event
		rootSubmissionCache[propID] = eventsForProp
	}

	// Derive the slot time from the network state's beacon config
	slotTime := time.Unix(int64(ns.BeaconConfig.GenesisTime+ns.BeaconSlotNumber*ns.BeaconConfig.SecondsPerSlot), 0)

	// For each proposal, crawl down the tree looking for unchallenged mismatches
	challenges := []challenge{}
	defeats := []defeat{}
	for _, prop := range mismatchingProps {
		challenge, defeat, err := getChallengeOrDefeatForProposal(prop, log, slotTime, stateGetter, artifactChecker, rootSubmissionCache)
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
func getChallengeOrDefeatForProposal(
	prop protocol.ProtocolDaoProposalDetails,
	log *log.ColorLogger,
	slotTime time.Time,
	stateGetter challengeStateGetter,
	artifactChecker challengeArtifactChecker,
	rootSubmissionCache map[uint64]map[uint64]*protocol.RootSubmitted,
) (*challenge, *defeat, error) {
	challengedIndex := uint64(1) // Root

	for {
		// Get the index of the node to challenge
		rootSubmissionEvent, exists := rootSubmissionCache[prop.ID][challengedIndex]
		if !exists {
			return nil, nil, fmt.Errorf("challenge against prop %d, index %d has been responded to but the RootSubmitted event was missing", prop.ID, challengedIndex)
		}
		newChallengedIndex, challengedNode, proof, err := artifactChecker.CheckForChallengeableArtifacts(*rootSubmissionEvent)
		if err != nil {
			return nil, nil, fmt.Errorf("error checking for challengeable artifacts on prop %d, index %s: %w", prop.ID, rootSubmissionEvent.Index.String(), err)
		}
		if newChallengedIndex == 0 {
			// Do nothing if the prop can't be challenged
			log.Printlnf("Check against proposal %d, index %d showed no challengeable artifacts.", prop.ID, challengedIndex)
			return nil, nil, nil
		}
		if newChallengedIndex == challengedIndex {
			// This shouldn't ever happen but it does then error out for safety
			return nil, nil, fmt.Errorf("cycle error: proposal %d had index %d challenged, and the new challengeable artifacts had the same index", prop.ID, challengedIndex)
		}

		// Check if the index has been challenged yet
		state, err := stateGetter.GetChallengeState(prop.ID, newChallengedIndex)
		if err != nil {
			return nil, nil, fmt.Errorf("error checking challenge state for proposal %d, index %d: %w", prop.ID, challengedIndex, err)
		}
		switch state {
		case types.ChallengeState_Unchallenged:
			// If it's unchallenged, this is the index to challenge
			return &challenge{
				proposalID:      prop.ID,
				challengedIndex: newChallengedIndex,
				challengedNode:  challengedNode,
				witness:         proof,
			}, nil, nil
		case types.ChallengeState_Challenged:
			// Check if the proposal can be defeated
			if slotTime.After(prop.CreatedTime.Add(prop.ChallengeWindow)) {
				return nil, &defeat{
					proposalID:      prop.ID,
					challengedIndex: newChallengedIndex,
				}, nil
			}
			// Nothing to do but wait for the proposer to respond
			log.Printlnf("Proposal %d, index %d has already been challenged; waiting for proposer to respond.", prop.ID, newChallengedIndex)
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
func (t *verifyPdaoProps) submitChallenge(challenge challenge) error {
	propID := challenge.proposalID
	challengedIndex := challenge.challengedIndex
	t.log.Printlnf("Submitting challenge against proposal %d, index %d...", propID, challengedIndex)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := protocol.EstimateCreateChallengeGas(t.rp, propID, challengedIndex, challenge.challengedNode, challenge.witness, opts)
	if err != nil {
		return fmt.Errorf("error estimating the gas required to submit challenge against proposal %d, index %d: %w", propID, challengedIndex, err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Respond to the challenge
	hash, err := protocol.CreateChallenge(t.rp, propID, challengedIndex, challenge.challengedNode, challenge.witness, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully submitted challenge.")

	// Return
	return nil
}

// Defeat a proposal
func (t *verifyPdaoProps) submitDefeat(defeat defeat) error {
	propID := defeat.proposalID
	challengedIndex := defeat.challengedIndex
	t.log.Printlnf("Proposal %d has been defeated with node index %d, submitting defeat...", propID, challengedIndex)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := protocol.EstimateDefeatProposalGas(t.rp, propID, challengedIndex, opts)
	if err != nil {
		return fmt.Errorf("error estimating the gas required to defeat proposal %d with index %d: %w", propID, challengedIndex, err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Respond to the challenge
	hash, err := protocol.DefeatProposal(t.rp, propID, challengedIndex, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully defeated proposal.")

	// Return
	return nil
}
