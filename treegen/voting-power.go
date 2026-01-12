package main

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

var (
	oneEth   = big.NewInt(1e18)
	_1_5_Eth = big.NewInt(15e17)
)

type VotingPowerFile struct {
	Network        string                                   `json:"network"`
	Time           time.Time                                `json:"time"`
	ConsensusSlot  uint64                                   `json:"consensusSlot"`
	ExecutionBlock uint64                                   `json:"executionBlock"`
	TotalPower     *rewards.QuotedBigInt                    `json:"totalPower"`
	NodePower      map[common.Address]*rewards.QuotedBigInt `json:"nodePower"`
}

func getNodeVotingPower(s *state.NetworkState, nodeIdx int) *big.Int {
	node := s.NodeDetails[nodeIdx]

	activeMinipoolCount := int64(0)
	for _, mpd := range s.MinipoolDetailsByNode[node.NodeAddress] {
		// Ignore finalised
		if mpd.Finalised {
			continue
		}

		activeMinipoolCount += 1
	}

	// Get provided ETH (32 * minipoolCount - matched)
	ethProvided := big.NewInt(activeMinipoolCount * 32)
	ethProvided.Mul(ethProvided, oneEth)
	ethProvided.Sub(ethProvided, node.EthBorrowed)

	// Add megapool provided ETH
	if s.IsSaturnDeployed && node.MegapoolDeployed {
		megapoolProvidedEth := s.MegapoolDetails[node.MegapoolAddress].NodeBond
		ethProvided.Add(ethProvided, megapoolProvidedEth)
	}

	// Get total RPL staked
	nodeStake := big.NewInt(0)
	nodeStake.Add(nodeStake, node.LegacyStakedRPL)
	if s.IsSaturnDeployed {
		nodeStake.Add(nodeStake, node.MegapoolStakedRPL)
	}

	rplPrice := s.NetworkDetails.RplPrice

	// No RPL staked means no voting power
	if nodeStake.Sign() == 0 {
		return big.NewInt(0)
	}

	// First calculate the maximum rpl that can be used as input
	// maxVotingRpl := (eligibleBondedEth * 1.5 Eth / RplPrice)
	maxVotingRpl := big.NewInt(0)
	maxVotingRpl.Mul(ethProvided, _1_5_Eth)
	maxVotingRpl.Quo(maxVotingRpl, rplPrice)

	// Determine the voting RPL
	// votingRpl := min(maxVotingRpl, nodeStake)
	var votingRpl *big.Int
	if maxVotingRpl.Cmp(nodeStake) <= 0 {
		votingRpl = maxVotingRpl
	} else {
		votingRpl = nodeStake
	}

	// Now take the square root
	// Because the units are in wei, we need to multiply votingRpl by 1 Eth before square rooting.
	votingPower := big.NewInt(0)
	votingPower.Mul(votingRpl, oneEth)
	votingPower.Sqrt(votingPower)
	return votingPower

}

func (g *treeGenerator) GenerateVotingPower(s *state.NetworkState) *VotingPowerFile {
	out := new(VotingPowerFile)

	out.Network = string(g.cfg.Smartnode.Network.Value.(cfgtypes.Network))
	out.ConsensusSlot = s.BeaconSlotNumber
	out.ExecutionBlock = s.ElBlockNumber
	out.TotalPower = rewards.NewQuotedBigInt(0)
	out.NodePower = make(map[common.Address]*rewards.QuotedBigInt, len(s.NodeDetails))
	for idx, node := range s.NodeDetails {

		// Calculate the Voting Power
		nodeVotingPower := rewards.NewQuotedBigInt(0)
		nodeVotingPower.Set(getNodeVotingPower(s, idx))
		out.TotalPower.Add(&out.TotalPower.Int, &nodeVotingPower.Int)
		out.NodePower[node.NodeAddress] = nodeVotingPower
	}

	return out
}
