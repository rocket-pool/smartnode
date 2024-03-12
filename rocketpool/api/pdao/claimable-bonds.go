package pdao

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
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

func getClaimableBonds(c *cli.Context) (*api.PDAOGetClaimableBondsResponse, error) {
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
	contracts, err := state.NewNetworkContracts(rp, mcAddress, bbAddress, true, nil)
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

	// Get the lists of proposals / challenge indices to check the state for
	propInfos := map[uint64]*proposalInfo{}
	proposalIDsForStates := []uint64{}
	challengedIndicesForStates := []uint64{}
	for i, prop := range props {
		shouldProcess := false
		isProposer := false

		if prop.ProposerAddress == nodeAccount.Address {
			isProposer = true
			if prop.State != types.ProtocolDaoProposalState_Defeated &&
				prop.State >= types.ProtocolDaoProposalState_QuorumNotMet {
				shouldProcess = true
			}
		} else {
			if prop.State != types.ProtocolDaoProposalState_Pending {
				shouldProcess = true
			}
		}

		if shouldProcess {
			// Add it to the map
			propInfo := &proposalInfo{
				ProtocolDaoProposalDetails: &props[i],
				IsProposer:                 isProposer,
				Challenges:                 map[uint64]*challengeInfo{},
			}
			propInfos[prop.ID] = propInfo

			// Get the events for all challenges against this proposal
			startBlock := big.NewInt(int64(prop.TargetBlock))                              // Target block is a good start for the event window
			blockSpan := uint64(prop.ChallengeWindow.Seconds()) / beaconCfg.SecondsPerSlot // The max possible number of blocks in the challenge window
			endBlockUint := uint64(prop.TargetBlock) + blockSpan
			endBlock := big.NewInt(int64(endBlockUint))

			// Get the RocketRewardsPool addresses
			verifierAddresses := cfg.Smartnode.GetPreviousRocketDAOProtocolVerifierAddresses()
			challengeEvents, err := protocol.GetChallengeSubmittedEvents(rp, []uint64{prop.ID}, intervalSize, startBlock, endBlock, verifierAddresses, nil)
			if err != nil {
				return nil, fmt.Errorf("error scanning for proposal %d's ChallengeSubmitted events: %w", prop.ID, err)
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
			if propInfo.State == types.ProtocolDaoProposalState_Defeated ||
				propInfo.State < types.ProtocolDaoProposalState_QuorumNotMet {
				// Proposer gets nothing if the challenge was defeated or isn't done yet
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
				if challengeInfo.State != types.ChallengeState_Challenged {
					if propInfo.State == types.ProtocolDaoProposalState_Defeated {
						if challengeInfo.State != types.ChallengeState_Responded {
							// If the proposal is defeated, a challenge must be in the challenged or responded states
							continue
						}
					} else {
						// Only refund non-responded challenges if the proposal wasn't defeated
						continue
					}
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

// Check if a node was part of a proposal's defeat path
func isRewardedIndex(defeatIndex uint64, nodeIndex uint64) bool {
	for i := defeatIndex; i > 1; i /= 2 {
		if i == nodeIndex {
			return true
		}
	}
	return false
}

func getElBlockForTimestamp(bc beacon.Client, beaconCfg beacon.Eth2Config, creationTime time.Time) (*big.Int, error) {
	// Get the slot number the first proposal was created on
	genesisTime := time.Unix(int64(beaconCfg.GenesisTime), 0)
	secondsPerSlot := time.Second * time.Duration(beaconCfg.SecondsPerSlot)
	startSlot := uint64(creationTime.Sub(genesisTime) / secondsPerSlot)

	// Get the Beacon block for the slot
	block, exists, err := bc.GetBeaconBlock(fmt.Sprint(startSlot))
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon block at slot %d: %w", startSlot, err)
	}
	if !exists {
		return nil, fmt.Errorf("Beacon block at slot %d was missing", startSlot)
	}

	// Get the EL block for this slot
	return big.NewInt(int64(block.ExecutionBlockNumber)), nil
}
