package pdao

import (
	"fmt"
	"math/big"
	"net/url"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	challengeStateBatchSize int = 500
)

// ===============
// === Factory ===
// ===============

type protocolDaoGetClaimableBondsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoGetClaimableBondsContextFactory) Create(args url.Values) (*protocolDaoGetClaimableBondsContext, error) {
	c := &protocolDaoGetClaimableBondsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoGetClaimableBondsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoGetClaimableBondsContext, api.ProtocolDaoGetClaimableBondsData](
		router, "get-claimable-bonds", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoGetClaimableBondsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	pdaoMgr *protocol.ProtocolDaoManager
}

func (c *protocolDaoGetClaimableBondsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Protocol DAO manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoGetClaimableBondsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
	)
}

func (c *protocolDaoGetClaimableBondsContext) PrepareData(data *api.ProtocolDaoGetClaimableBondsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Get the proposals
	props, err := c.pdaoMgr.GetProposals(c.pdaoMgr.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Protocol DAO proposal details: %w", err)
	}

	// Get some common vars
	beaconCfg, err := c.bc.GetEth2Config(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Beacon config: %w", err)
	}
	intervalSize := big.NewInt(int64(config.EventLogInterval))

	// Get the lists of proposals / challenge indices to check the state for
	propInfos := map[uint64]*proposalInfo{}
	proposalIDsForStates := []uint64{}
	challengedIndicesForStates := []uint64{}
	for i, prop := range props {
		shouldProcess := false
		isProposer := false

		state := prop.State.Formatted()
		if prop.ProposerAddress.Get() == c.nodeAddress {
			isProposer = true
			if state >= rptypes.ProtocolDaoProposalState_QuorumNotMet {
				shouldProcess = true
			}
		} else {
			if state != rptypes.ProtocolDaoProposalState_Pending {
				shouldProcess = true
			}
		}

		if shouldProcess {
			// Add it to the map
			propInfo := &proposalInfo{
				ProtocolDaoProposal: props[i],
				IsProposer:          isProposer,
				Challenges:          map[uint64]*challengeInfo{},
			}
			propInfos[prop.ID] = propInfo

			// Get the events for all challenges against this proposal
			startBlock := prop.TargetBlock.Raw()                                                       // Target block is a good start for the event window
			blockSpan := uint64(prop.ChallengeWindow.Formatted().Seconds()) / beaconCfg.SecondsPerSlot // The max possible number of blocks in the challenge window
			endBlock := big.NewInt(0).Add(startBlock, big.NewInt(int64(blockSpan)))

			resources := c.cfg.GetRocketPoolResources()
			challengeEvents, err := c.pdaoMgr.GetChallengeSubmittedEvents([]uint64{prop.ID}, intervalSize, startBlock, endBlock, resources.PreviousProtocolDaoVerifierAddresses, nil)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error scanning for proposal %d's ChallengeSubmitted events: %w", prop.ID, err)
			}

			// Add an explicit challenge event to the root to see if the proposal bond is refundable
			proposalIDsForStates = append(proposalIDsForStates, prop.ID)
			challengedIndicesForStates = append(challengedIndicesForStates, 1)
			propInfo.Challenges[1] = &challengeInfo{
				Challenger: common.Address{},
				State:      0,
			}

			// Add each event to the list to get the states for
			for _, event := range challengeEvents {
				challengedIndex := event.Index.Uint64()
				proposalIDsForStates = append(proposalIDsForStates, prop.ID)
				challengedIndicesForStates = append(challengedIndicesForStates, challengedIndex)
				propInfo.Challenges[challengedIndex] = &challengeInfo{
					Challenger: event.Challenger,
					State:      0,
				}
			}
		}
	}

	// Get the states of all challenges
	states := make([]func() rptypes.ChallengeState, len(proposalIDsForStates))
	err = c.rp.BatchQuery(len(proposalIDsForStates), challengeStateBatchSize, func(mc *batch.MultiCaller, i int) error {
		propID := proposalIDsForStates[i]
		index := challengedIndicesForStates[i]
		prop := propInfos[propID]
		states[i] = prop.GetChallengeState(mc, index)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting challenge states: %w", err)
	}

	// Map out the challenge results
	for i := 0; i < len(proposalIDsForStates); i++ {
		propID := proposalIDsForStates[i]
		propInfo := propInfos[propID]

		challengedIndex := challengedIndicesForStates[i]
		state := states[i]
		propInfo.Challenges[challengedIndex].State = state()
	}

	// Go through each challenge to see if it's refundable
	claimResults := map[uint64]*api.BondClaimResult{}
	for _, propInfo := range propInfos {
		// Make a claim result for this proposal
		claimResult := &api.BondClaimResult{
			ProposalID:        propInfo.ID,
			IsProposer:        propInfo.IsProposer,
			UnlockableIndices: []uint64{},
			RewardableIndices: []uint64{},
			UnlockAmount:      big.NewInt(0),
			RewardAmount:      big.NewInt(0),
		}
		claimResults[propInfo.ID] = claimResult

		// Handle proposals we were the proposer of
		if claimResult.IsProposer {
			state := propInfo.State.Formatted()
			if state < rptypes.ProtocolDaoProposalState_QuorumNotMet {
				// Proposer gets nothing if the challenge was defeated or isn't done yet
				continue
			}
			for challengedIndex, challengeInfo := range propInfo.Challenges {
				if challengeInfo.State == rptypes.ChallengeState_Responded {
					if challengedIndex == 1 {
						// The proposal bond can be unlocked
						claimResult.UnlockableIndices = append(claimResult.UnlockableIndices, 1)
						claimResult.UnlockAmount.Add(claimResult.UnlockAmount, propInfo.ProposalBond.Get())
					} else {
						// This is a challenged index that can be claimed
						claimResult.RewardableIndices = append(claimResult.RewardableIndices, challengedIndex)
						claimResult.RewardAmount.Add(claimResult.RewardAmount, propInfo.ChallengeBond.Get())
					}
				}
			}
		} else {
			// Check if this node has any unpaid challenges
			totalContributingChallenges := int64(0)
			rewardCount := int64(0)
			unlockableChallengeCount := int64(0)
			for challengedIndex, challengeInfo := range propInfo.Challenges {
				// Ignore the root
				if challengedIndex == 1 {
					continue
				}

				// Make sure the prop and challenge are in the right states
				if challengeInfo.State == rptypes.ChallengeState_Paid {
					// Ignore already paid challenges
					continue
				}
				if challengeInfo.State == rptypes.ChallengeState_Responded && propInfo.State.Formatted() != rptypes.ProtocolDaoProposalState_Destroyed {
					// Only refund responded challenges if the proposal was destroyed
					continue
				}
				if challengeInfo.State == rptypes.ChallengeState_Challenged && propInfo.State.Formatted() < rptypes.ProtocolDaoProposalState_QuorumNotMet {
					// Unresponded challenges may be claimed after the proposal is finished
					continue
				}

				// Increment how many refundable challenges we made
				isOwnChallenge := (challengeInfo.Challenger == c.nodeAddress)
				if isOwnChallenge {
					unlockableChallengeCount++
					claimResult.UnlockableIndices = append(claimResult.UnlockableIndices, challengedIndex)
				}

				// Check if this challenge contributed to the proposal's defeat
				if isRewardedIndex(propInfo.DefeatIndex.Formatted(), challengedIndex) {
					totalContributingChallenges++
					if isOwnChallenge {
						// Reward valid challenges from this node
						rewardCount++
						claimResult.RewardableIndices = append(claimResult.RewardableIndices, challengedIndex)
					}
				}
			}

			// Mark how much RPL can be unlocked
			if unlockableChallengeCount > 0 {
				totalUnlock := big.NewInt(unlockableChallengeCount)
				totalUnlock.Mul(totalUnlock, propInfo.ChallengeBond.Get())
				claimResult.UnlockAmount.Add(claimResult.UnlockAmount, totalUnlock)
			}

			// How much RPL will be rewarded
			if rewardCount > 0 {
				totalReward := big.NewInt(rewardCount)
				totalReward.Mul(totalReward, propInfo.ProposalBond.Get())
				totalReward.Div(totalReward, big.NewInt(totalContributingChallenges))
				claimResult.RewardAmount.Add(claimResult.RewardAmount, totalReward)
			}
		}
	}

	// Make a sorted list of claimable bonds
	claimableBonds := make([]api.BondClaimResult, 0, len(claimResults))
	for _, result := range claimResults {
		if result.RewardAmount.Cmp(big.NewInt(0)) > 0 || result.UnlockAmount.Cmp(big.NewInt(0)) > 0 {
			claimableBonds = append(claimableBonds, *result)
		}
	}
	sort.SliceStable(claimableBonds, func(i, j int) bool {
		first := claimableBonds[i]
		second := claimableBonds[j]
		return first.ProposalID < second.ProposalID
	})

	// Update the response and return
	data.ClaimableBonds = claimableBonds
	return types.ResponseStatus_Success, nil
}

type challengeInfo struct {
	Challenger common.Address
	State      rptypes.ChallengeState
}

type proposalInfo struct {
	*protocol.ProtocolDaoProposal
	IsProposer bool
	Challenges map[uint64]*challengeInfo
}

// Check if a node was part of a proposal's defeat path
func isRewardedIndex(defeatIndex uint64, nodeIndex uint64) bool {
	for i := defeatIndex; i > 1; i /= 2 {
		if i == nodeIndex {
			return true
		}
	}
	return false
}
