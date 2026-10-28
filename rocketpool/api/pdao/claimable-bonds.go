package pdao

import (
	"fmt"
	"log"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type challengeInfo struct {
	Challenger common.Address
	State      types.ChallengeState
}

type proposalInfo struct {
	*protocol.ProtocolDaoProposalDetails
	IsProposer bool
	Challenges map[uint64]*challengeInfo
}

func getClaimableBonds(c *cli.Command) (*api.PDAOGetClaimableBondsResponse, error) {
	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
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

	// Response
	response := api.PDAOGetClaimableBondsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Set up multicall
	mcAddress := common.HexToAddress(cfg.Smartnode.GetMulticallAddress())
	bbAddress := common.HexToAddress(cfg.Smartnode.GetBalanceBatcherAddress())
	contracts, err := state.NewNetworkContracts(rp, mcAddress, bbAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating network contracts: %w", err)
	}

	// Get all of the proposals
	props, err := state.GetAllProtocolDaoProposalDetails(rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting pDAO proposal details: %w", err)
	}
	if len(props) == 0 {
		response.ClaimableBonds = []api.BondClaimResult{}
		return &response, nil
	}

	// Get some common vars
	beaconCfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon config: %w", err)
	}
	intervalSize := big.NewInt(int64(cfg.Geth.EventLogInterval))
	verifierAddresses := cfg.Smartnode.GetPreviousRocketDAOProtocolVerifierAddresses()

	// Newest proposals are scanned first. An unreadable window is skipped so
	// the remaining proposals are still checked.
	propInfos, proposalIDsForStates, challengedIndicesForStates, scanErr := collectProposalChallenges(
		props,
		nodeAccount.Address,
		beaconCfg.SecondsPerSlot,
		func(prop *protocol.ProtocolDaoProposalDetails, startBlock, endBlock *big.Int) ([]protocol.ChallengeSubmitted, error) {
			return protocol.GetChallengeSubmittedEvents(rp, []uint64{prop.ID}, intervalSize, startBlock, endBlock, verifierAddresses, nil)
		},
	)
	if len(proposalIDsForStates) == 0 {
		if scanErr != nil {
			return nil, scanErr
		}
		response.ClaimableBonds = []api.BondClaimResult{}
		return &response, nil
	}

	// Get the states of each prop / challenged index
	states, err := protocol.GetMultiChallengeStatesFast(rp, mcAddress, proposalIDsForStates, challengedIndicesForStates, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting challenge states: %w", err)
	}

	// Map out the challenge results
	for i := 0; i < len(proposalIDsForStates); i++ {
		propID := proposalIDsForStates[i]
		propInfo := propInfos[propID]

		challengedIndex := challengedIndicesForStates[i]
		state := states[i]
		propInfo.Challenges[challengedIndex].State = state
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
			if propInfo.State < types.ProtocolDaoProposalState_QuorumNotMet {
				// Proposer gets nothing if the challenge isn't done yet
				continue
			}
			for challengedIndex, challengeInfo := range propInfo.Challenges {
				if challengeInfo.State == types.ChallengeState_Responded {
					if challengedIndex == 1 {
						// The proposal bond can be unlocked
						claimResult.UnlockableIndices = append(claimResult.UnlockableIndices, 1)
						claimResult.UnlockAmount.Add(claimResult.UnlockAmount, propInfo.ProposalBond)
					} else {
						// This is a challenged index that can be claimed
						claimResult.RewardableIndices = append(claimResult.RewardableIndices, challengedIndex)
						claimResult.RewardAmount.Add(claimResult.RewardAmount, propInfo.ChallengeBond)
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
				if challengeInfo.State == types.ChallengeState_Paid {
					// Ignore already paid challenges
					continue
				}
				if challengeInfo.State == types.ChallengeState_Responded && propInfo.State != types.ProtocolDaoProposalState_Destroyed {
					// Only refund responded challenges if the proposal was destroyed
					continue
				}
				if challengeInfo.State == types.ChallengeState_Challenged && !unrespondedChallengeRefundable(propInfo.State) {
					// Unresponded challenges may be claimed once the proposal is finished.
					// Destroyed and Vetoed sort before QuorumNotMet, so a numeric comparison
					// hides bonds on proposals defeated during the challenge phase.
					continue
				}

				// Increment how many refundable challenges we made
				isOwnChallenge := (challengeInfo.Challenger == nodeAccount.Address)
				if isOwnChallenge {
					unlockableChallengeCount++
					claimResult.UnlockableIndices = append(claimResult.UnlockableIndices, challengedIndex)
				}

				// Check if this challenge contributed to the proposal's defeat
				if isRewardedIndex(propInfo.DefeatIndex, challengedIndex) {
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
				totalUnlock.Mul(totalUnlock, propInfo.ChallengeBond)
				claimResult.UnlockAmount.Add(claimResult.UnlockAmount, totalUnlock)
			}

			// How much RPL will be rewarded
			if rewardCount > 0 {
				totalReward := big.NewInt(rewardCount)
				totalReward.Mul(totalReward, propInfo.ProposalBond)
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
	response.ClaimableBonds = claimableBonds
	return &response, nil
}

func collectProposalChallenges(
	props []protocol.ProtocolDaoProposalDetails,
	nodeAccount common.Address,
	secondsPerSlot uint64,
	scan func(prop *protocol.ProtocolDaoProposalDetails, startBlock, endBlock *big.Int) ([]protocol.ChallengeSubmitted, error),
) (map[uint64]*proposalInfo, []uint64, []uint64, error) {
	propInfos := map[uint64]*proposalInfo{}
	proposalIDsForStates := []uint64{}
	challengedIndicesForStates := []uint64{}
	var scanErr error

	for i := len(props) - 1; i >= 0; i-- {
		prop := &props[i]
		isProposer := prop.ProposerAddress == nodeAccount
		if !proposalHasClaimableChallenges(prop, isProposer) {
			continue
		}

		startBlock := big.NewInt(int64(prop.TargetBlock))
		blockSpan := uint64(prop.ChallengeWindow.Seconds()) / secondsPerSlot
		endBlock := big.NewInt(int64(uint64(prop.TargetBlock) + blockSpan))

		challengeEvents, err := scan(prop, startBlock, endBlock)
		if err != nil {
			log.Printf("PDAO claimable bonds: skipping proposal %d because its challenge events could not be read: %v", prop.ID, err)
			if scanErr == nil {
				scanErr = fmt.Errorf("error scanning for proposal %d's ChallengeSubmitted events: %w", prop.ID, err)
			}
			// The root bond does not depend on challenge events.
			if isProposer {
				propInfo := newProposalInfo(prop, true)
				propInfos[prop.ID] = propInfo
				appendRootChallenge(propInfo, &proposalIDsForStates, &challengedIndicesForStates)
			}
			continue
		}

		propInfo := newProposalInfo(prop, isProposer)
		propInfos[prop.ID] = propInfo
		appendRootChallenge(propInfo, &proposalIDsForStates, &challengedIndicesForStates)
		for _, event := range challengeEvents {
			challengedIndex := event.Index.Uint64()
			proposalIDsForStates = append(proposalIDsForStates, prop.ID)
			challengedIndicesForStates = append(challengedIndicesForStates, challengedIndex)
			propInfo.Challenges[challengedIndex] = &challengeInfo{
				Challenger: event.Challenger,
			}
		}
	}

	return propInfos, proposalIDsForStates, challengedIndicesForStates, scanErr
}

// unrespondedChallengeRefundable reports whether a Challenged (not yet
// responded) bond can be claimed.
func unrespondedChallengeRefundable(state types.ProtocolDaoProposalState) bool {
	return state == types.ProtocolDaoProposalState_Destroyed ||
		state == types.ProtocolDaoProposalState_Vetoed ||
		state >= types.ProtocolDaoProposalState_QuorumNotMet
}

func proposalHasClaimableChallenges(prop *protocol.ProtocolDaoProposalDetails, isProposer bool) bool {
	if isProposer {
		return prop.State >= types.ProtocolDaoProposalState_QuorumNotMet
	}
	return prop.State != types.ProtocolDaoProposalState_Pending
}

func newProposalInfo(prop *protocol.ProtocolDaoProposalDetails, isProposer bool) *proposalInfo {
	return &proposalInfo{
		ProtocolDaoProposalDetails: prop,
		IsProposer:                 isProposer,
		Challenges:                 map[uint64]*challengeInfo{},
	}
}

func appendRootChallenge(propInfo *proposalInfo, proposalIDs *[]uint64, challengedIndices *[]uint64) {
	*proposalIDs = append(*proposalIDs, propInfo.ID)
	*challengedIndices = append(*challengedIndices, 1)
	propInfo.Challenges[1] = &challengeInfo{}
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

func getClaimableBondsHandler(ctx snroute.Context) {
	resp, err := getClaimableBonds(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
