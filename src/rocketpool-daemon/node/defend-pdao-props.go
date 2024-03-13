package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet"
	"github.com/rocket-pool/smartnode/shared/config"
)

type defendableProposal struct {
	challengeEvent *protocol.ChallengeSubmitted
	proposal       *protocol.ProtocolDaoProposal
}

type DefendPdaoProps struct {
	sp               *services.ServiceProvider
	log              *log.ColorLogger
	cfg              *config.SmartNodeConfig
	w                *wallet.LocalWallet
	rp               *rocketpool.RocketPool
	bc               beacon.IBeaconClient
	gasThreshold     float64
	maxFee           *big.Int
	maxPriorityFee   *big.Int
	gasLimit         uint64
	nodeAddress      common.Address
	propMgr          *proposals.ProposalManager
	pdaoMgr          *protocol.ProtocolDaoManager
	lastScannedBlock *big.Int

	// Smartnode parameters
	intervalSize *big.Int
}

func NewDefendPdaoProps(sp *services.ServiceProvider, logger log.ColorLogger) *DefendPdaoProps {
	return &DefendPdaoProps{
		sp:               sp,
		log:              &logger,
		lastScannedBlock: nil,
	}
}

// Defend pDAO proposals
func (t *DefendPdaoProps) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.w = t.sp.GetWallet()
	t.nodeAddress, _ = t.w.GetAddress()
	t.maxFee, t.maxPriorityFee = getAutoTxInfo(t.cfg, t.log)
	t.gasThreshold = t.cfg.Smartnode.AutoTxGasThreshold.Value.(float64)
	t.intervalSize = big.NewInt(int64(t.cfg.Geth.EventLogInterval))

	// Bindings
	propMgr, err := proposals.NewProposalManager(t.log, t.cfg, t.rp, t.bc)
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
func (t *DefendPdaoProps) getDefendableProposals(state *state.NetworkState, opts *bind.CallOpts) ([]defendableProposal, error) {
	// Get proposals made by this node that are still in the challenge phase (Pending)
	eligibleProps := []*protocol.ProtocolDaoProposal{}
	for _, prop := range state.ProtocolDaoProposalDetails {
		if prop.ProposerAddress.Get() == t.nodeAddress &&
			time.Until(prop.CreatedTime.Formatted().Add(prop.ChallengeWindow.Formatted())) > 0 {
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
		startTime := eligibleProps[0].CreatedTime.Formatted()
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
	propMap := map[uint64]*protocol.ProtocolDaoProposal{}
	for i, prop := range eligibleProps {
		ids[i] = prop.ID
		propMap[prop.ID] = eligibleProps[i]
	}

	// Get any challenges issued for the proposals
	challengeEvents, err := t.pdaoMgr.GetChallengeSubmittedEvents(ids, t.intervalSize, startBlock, endBlock, opts)
	if err != nil {
		return nil, fmt.Errorf("error scanning for ChallengeSubmitted events: %w", err)
	}

	// Parse them out
	defendableProposals := []defendableProposal{}
	for _, event := range challengeEvents {
		// Check if the challenge has been handled yet
		propID := event.ProposalID.Uint64()
		index := event.Index.Uint64()
		prop := propMap[propID]
		var state func() types.ChallengeState
		err = t.rp.Query(func(mc *batch.MultiCaller) error {
			state = prop.GetChallengeState(mc, index)
			return nil
		}, opts)
		if err != nil {
			return nil, fmt.Errorf("error checking challenge state for proposal %d, index %d: %w", prop.ID, index, err)
		}
		if state() == types.ChallengeState_Challenged {
			t.log.Printlnf("Proposal %d, index %d has been challenged by %s.", propID, index, event.Challenger.Hex())
			defendableProposals = append(defendableProposals, defendableProposal{
				challengeEvent: &event,
				proposal:       prop,
			})
		}
	}

	return defendableProposals, nil
}

// Submit a response to a challenge against one of this node's proposals
func (t *DefendPdaoProps) defendProposal(prop defendableProposal) error {
	propID := prop.proposal.ID
	challengedIndex := prop.challengeEvent.Index.Uint64()
	t.log.Printlnf("Responding to challenge against proposal %d, index %d...", propID, challengedIndex)

	// Create the response pollard
	_, pollard, err := t.propMgr.GetArtifactsForChallengeResponse(prop.proposal.TargetBlock.Formatted(), challengedIndex)
	if err != nil {
		return fmt.Errorf("error getting pollard for response to challenge against proposal %d, index %d: %w", propID, challengedIndex, err)
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the tx info
	txInfo, err := prop.proposal.SubmitRoot(challengedIndex, pollard, opts)
	if err != nil {
		return fmt.Errorf("error estimating the gas required to respond to challenge against proposal %d, index %d: %w", propID, challengedIndex, err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating response to challenge against proposal %d, index %d failed: %s", propID, challengedIndex, txInfo.SimError)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.log)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		t.log.Println("NOTICE: Challenge responses bypass the automatic TX gas threshold, responding for safety.")
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Successfully responded to challenge.")
	return nil
}
