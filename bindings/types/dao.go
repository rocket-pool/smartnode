package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/utils/json"
)

// DAO proposal states
type ProposalState uint8

const (
	Pending ProposalState = iota
	Active
	Cancelled
	Defeated
	Succeeded
	Expired
	Executed
)

var ProposalStates = []string{"Pending", "Active", "Cancelled", "Defeated", "Succeeded", "Expired", "Executed"}

// pDAO proposal states
type ProtocolDaoProposalState uint8

const (
	ProtocolDaoProposalState_Pending ProtocolDaoProposalState = iota
	ProtocolDaoProposalState_ActivePhase1
	ProtocolDaoProposalState_ActivePhase2
	ProtocolDaoProposalState_Destroyed
	ProtocolDaoProposalState_Vetoed
	ProtocolDaoProposalState_QuorumNotMet
	ProtocolDaoProposalState_Defeated
	ProtocolDaoProposalState_Succeeded
	ProtocolDaoProposalState_Expired
	ProtocolDaoProposalState_Executed
)

var ProtocolDaoProposalStates = []string{"Pending", "Active (Phase 1)", "Active (Phase 2)", "Destroyed", "Vetoed", "Quorum not Met", "Defeated", "Succeeded", "Expired", "Executed"}

// pDAO voting direction
type VoteDirection uint8

const (
	VoteDirection_NoVote VoteDirection = iota
	VoteDirection_Abstain
	VoteDirection_For
	VoteDirection_Against
	VoteDirection_AgainstWithVeto
)

var VoteDirections = []string{"Not Voted", "Abstain", "In Favor", "Against", "Against with Veto"}

// DAO setting types
type ProposalSettingType uint8

const (
	ProposalSettingType_Uint256 ProposalSettingType = iota
	ProposalSettingType_Bool
	ProposalSettingType_Address
)

// Challenge states
type ChallengeState uint8

const (
	ChallengeState_Unchallenged ChallengeState = iota
	ChallengeState_Challenged
	ChallengeState_Responded
	ChallengeState_Paid
)

// Info about a node's voting power
type NodeVotingInfo struct {
	NodeAddress common.Address `json:"nodeAddress"`
	VotingPower *big.Int       `json:"votingPower"`
	Delegate    common.Address `json:"delegate"`
}

// A node of the voting Merkle Tree (not a Rocket Pool node)
type VotingTreeNode struct {
	Sum  *big.Int    `json:"sum"`
	Hash common.Hash `json:"hash"`
}

// String conversion
func (s ProposalState) String() string {
	if int(s) >= len(ProposalStates) {
		return ""
	}
	return ProposalStates[s]
}
func StringToProposalState(value string) (ProposalState, error) {
	for state, str := range ProposalStates {
		if value == str {
			return ProposalState(state), nil
		}
	}
	return 0, fmt.Errorf("Invalid proposal state '%s'", value)
}

// JSON encoding
func (s ProposalState) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == "" {
		return []byte{}, fmt.Errorf("Invalid proposal state '%d'", s)
	}
	return json.Marshal(str)
}
func (s *ProposalState) UnmarshalJSON(data []byte) error {
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}
	state, err := StringToProposalState(dataStr)
	if err == nil {
		*s = state
	}
	return err
}
