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

type defendableProposal struct {
	challengeEvent *protocol.ChallengeSubmitted
	proposal       *protocol.ProtocolDaoProposal
}

type DefendPdaoProps struct {
	ctx              context.Context
	sp               *services.ServiceProvider
	logger           *slog.Logger
	cfg              *config.SmartNodeConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	bc               beacon.IBeaconClient
	rs               *config.RocketPoolResources
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

func NewDefendPdaoProps(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger) *DefendPdaoProps {
	cfg := sp.GetConfig()
	log := logger.With(slog.String(keys.RoutineKey, "Defend PDAO Proposals"))
	maxFee, maxPriorityFee := getAutoTxInfo(cfg, log)
	return &DefendPdaoProps{
		ctx:              ctx,
		sp:               sp,
		logger:           log,
		cfg:              cfg,
		w:                sp.GetWallet(),
		rp:               sp.GetRocketPool(),
		bc:               sp.GetBeaconClient(),
		rs:               cfg.GetRocketPoolResources(),
		gasThreshold:     cfg.AutoTxGasThreshold.Value,
		maxFee:           maxFee,
		maxPriorityFee:   maxPriorityFee,
		lastScannedBlock: nil,
		intervalSize:     big.NewInt(int64(config.EventLogInterval)),
	}
}

// Defend pDAO proposals
func (t *DefendPdaoProps) Run(state *state.NetworkState) error {
	t.nodeAddress, _ = t.w.GetAddress()

	// Bindings
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
	t.logger.Info("Started checking for Protocol DAO proposal challenges to defend.")

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
			prop.State.Formatted() == types.ProtocolDaoProposalState_Pending {
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
		block, exists, err := t.bc.GetBeaconBlock(t.ctx, fmt.Sprint(startSlot))
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
	challengeEvents, err := t.pdaoMgr.GetChallengeSubmittedEvents(ids, t.intervalSize, startBlock, endBlock, t.rs.PreviousProtocolDaoVerifierAddresses, opts)
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
			t.logger.Info("Challenge detected", slog.Uint64(keys.ProposalKey, propID), slog.Uint64(keys.IndexKey, index), slog.String(keys.ChallengerKey, event.Challenger.Hex()))
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
	t.logger.Info("Responding to challenge...", slog.Uint64(keys.ProposalKey, propID), slog.Uint64(keys.IndexKey, challengedIndex))

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
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating response to challenge against proposal %d, index %d failed: %s", propID, challengedIndex, txInfo.SimulationResult.SimulationError)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.logger)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, true, t.gasThreshold, t.logger, maxFee, t.gasLimit) {
		t.logger.Warn("NOTICE: Challenge responses bypass the automatic TX gas threshold, responding for safety.")
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully responded to challenge.")
	return nil
}
