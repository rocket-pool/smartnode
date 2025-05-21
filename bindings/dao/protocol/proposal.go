package protocol

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	strutils "github.com/rocket-pool/rocketpool-go/utils/strings"
	"golang.org/x/sync/errgroup"
)

// Settings
const (
	ProposalDAONamesBatchSize = 50
	ProposalDetailsBatchSize  = 10
)

// =====================
// === Proposal Info ===
// =====================

// Proposal details
type ProtocolDaoProposalDetails struct {
	ID                   uint64                         `json:"id"`
	DAO                  string                         `json:"dao"`
	ProposerAddress      common.Address                 `json:"proposerAddress"`
	TargetBlock          uint32                         `json:"targetBlock"`
	Message              string                         `json:"message"`
	CreatedTime          time.Time                      `json:"createdTime"`
	ChallengeWindow      time.Duration                  `json:"challengeWindow"`
	VotingStartTime      time.Time                      `json:"startTime"`
	Phase1EndTime        time.Time                      `json:"phase1EndTime"`
	Phase2EndTime        time.Time                      `json:"phase2EndTime"`
	ExpiryTime           time.Time                      `json:"expiryTime"`
	VotingPowerRequired  *big.Int                       `json:"votingPowerRequired"`
	VotingPowerFor       *big.Int                       `json:"votingPowerFor"`
	VotingPowerAgainst   *big.Int                       `json:"votingPowerAgainst"`
	VotingPowerAbstained *big.Int                       `json:"votingPowerAbstained"`
	VotingPowerToVeto    *big.Int                       `json:"votingPowerVeto"`
	IsDestroyed          bool                           `json:"isDestroyed"`
	IsFinalized          bool                           `json:"isFinalized"`
	IsExecuted           bool                           `json:"isExecuted"`
	IsVetoed             bool                           `json:"isVetoed"`
	VetoQuorum           *big.Int                       `json:"vetoQuorum"`
	Payload              []byte                         `json:"payload"`
	PayloadStr           string                         `json:"payloadStr"`
	State                types.ProtocolDaoProposalState `json:"state"`
	ProposalBond         *big.Int                       `json:"proposalBond"`
	ChallengeBond        *big.Int                       `json:"challengeBond"`
	DefeatIndex          uint64                         `json:"defeatIndex"`
}

// Get all proposal details
func GetProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]ProtocolDaoProposalDetails, error) {
	// Get proposal count
	proposalCount, err := GetTotalProposalCount(rp, opts)
	if err != nil {
		return []ProtocolDaoProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]ProtocolDaoProposalDetails, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += ProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetProposalDetails(rp, pi+1, opts) // Proposals are 1-indexed
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []ProtocolDaoProposalDetails{}, err
		}

	}

	// Return
	return details, nil
}

// Get a proposal's details
func GetProposalDetails(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (ProtocolDaoProposalDetails, error) {
	var wg errgroup.Group
	var prop ProtocolDaoProposalDetails
	prop.ID = proposalId

	// Load data
	wg.Go(func() error {
		var err error
		prop.ProposerAddress, err = GetProposalProposer(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.TargetBlock, err = GetProposalBlock(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Message, err = GetProposalMessage(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingStartTime, err = GetProposalVotingStartTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Phase1EndTime, err = GetProposalPhase1EndTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Phase2EndTime, err = GetProposalPhase2EndTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.ExpiryTime, err = GetProposalExpiryTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.CreatedTime, err = GetProposalCreationTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingPowerRequired, err = GetProposalVotingPowerRequired(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingPowerFor, err = GetProposalVotingPowerFor(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingPowerAgainst, err = GetProposalVotingPowerAgainst(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingPowerAbstained, err = GetProposalVotingPowerAbstained(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VotingPowerToVeto, err = GetProposalVotingPowerVetoed(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.IsDestroyed, err = GetProposalIsDestroyed(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.IsFinalized, err = GetProposalIsFinalized(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.IsExecuted, err = GetProposalIsExecuted(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.IsVetoed, err = GetProposalIsVetoed(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.VetoQuorum, err = GetProposalVetoQuorum(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Payload, err = GetProposalPayload(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.State, err = GetProposalState(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.DefeatIndex, err = GetDefeatIndex(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.ProposalBond, err = GetProposalBond(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.ChallengeBond, err = GetChallengeBond(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.ChallengeWindow, err = GetChallengePeriod(rp, proposalId, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return ProtocolDaoProposalDetails{}, err
	}

	// Get proposal payload string
	payloadStr, err := GetProposalPayloadString(rp, prop.Payload, opts)
	if err != nil {
		payloadStr = "(unknown)"
	}
	prop.PayloadStr = payloadStr
	return prop, nil
}

// Get the block that was used for voting power calculation in a proposal
func GetProposalBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint32, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getProposalBlock", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return 0, fmt.Errorf("error getting proposal block for proposal %d: %w", proposalId, err)
	}
	return uint32((*value).Uint64()), nil
}

// Get the veto quorum required to veto a proposal
func GetProposalVetoQuorum(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getProposalVetoQuorum", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting proposal veto quorum for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// The total number of Protocol DAO proposals
func GetTotalProposalCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getTotal"); err != nil {
		return 0, fmt.Errorf("error getting total proposal count: %w", err)
	}
	return (*value).Uint64(), nil
}

// Get the address of the proposer
func GetProposalProposer(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getProposer", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return common.Address{}, fmt.Errorf("error getting proposer for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the proposal's message
func GetProposalMessage(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return "", err
	}
	value := new(string)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getMessage", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return "", fmt.Errorf("error getting message for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the start time of this proposal, when voting begins
func GetProposalVotingStartTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getStart", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return time.Time{}, fmt.Errorf("error getting start block for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the phase 1 end time of this proposal
func GetProposalPhase1EndTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getPhase1End", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return time.Time{}, fmt.Errorf("error getting phase 1 end time for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the phase 2 end time of this proposal
func GetProposalPhase2EndTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getPhase2End", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return time.Time{}, fmt.Errorf("error getting phase 2 end time for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the time where the proposal expires and can no longer be executed if it is successful
func GetProposalExpiryTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getExpires", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return time.Time{}, fmt.Errorf("error getting expiry time for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the time the proposal was created
func GetProposalCreationTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getCreated", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return time.Time{}, fmt.Errorf("error getting creation time for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the cumulative amount of voting power voting in favor of this proposal
func GetProposalVotingPowerFor(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVotingPowerFor", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting total 'for' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power voting against this proposal
func GetProposalVotingPowerAgainst(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVotingPowerAgainst", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting total 'against' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that vetoed this proposal
func GetProposalVotingPowerVetoed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVotingPowerVeto", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting total 'veto' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that abstained from this proposal
func GetProposalVotingPowerAbstained(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVotingPowerAbstained", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting total 'abstained' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that must vote on this proposal for it to be eligible for execution if it succeeds
func GetProposalVotingPowerRequired(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVotingPowerRequired", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting required voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been destroyed
func GetProposalIsDestroyed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getDestroyed", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return false, fmt.Errorf("error getting destroyed status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been finalized
func GetProposalIsFinalized(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getFinalised", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return false, fmt.Errorf("error getting finalized status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been executed
func GetProposalIsExecuted(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getExecuted", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return false, fmt.Errorf("error getting executed status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal's veto quorum has been met and it has been vetoed
func GetProposalIsVetoed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getVetoed", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return false, fmt.Errorf("error getting veto status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the proposal's payload
func GetProposalPayload(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) ([]byte, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new([]byte)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getPayload", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return nil, fmt.Errorf("error getting payload of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get a proposal's payload as a human-readable string
func GetProposalPayloadString(rp *rocketpool.RocketPool, payload []byte, opts *bind.CallOpts) (string, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return "", err
	}

	// Get proposal DAO contract ABI
	daoContractAbi := rocketDAOProtocolProposals.ABI

	// Get proposal payload method
	method, err := daoContractAbi.MethodById(payload)
	if err != nil {
		return "", fmt.Errorf("error getting proposal payload method: %w", err)
	}

	// Get proposal payload argument values
	args, err := method.Inputs.UnpackValues(payload[4:])
	if err != nil {
		return "", fmt.Errorf("error getting proposal payload arguments: %w", err)
	}

	// Format argument values as strings
	argStrs := []string{}
	for ai, arg := range args {
		switch method.Inputs[ai].Type.T {
		case abi.AddressTy:
			argStrs = append(argStrs, arg.(common.Address).Hex())
		case abi.HashTy:
			argStrs = append(argStrs, arg.(common.Hash).Hex())
		case abi.FixedBytesTy:
			fallthrough
		case abi.BytesTy:
			argStrs = append(argStrs, hex.EncodeToString(arg.([]byte)))
		default:
			argStrs = append(argStrs, fmt.Sprintf("%v", arg))
		}
	}

	// Build & return payload string
	return strutils.Sanitize(fmt.Sprintf("%s(%s)", method.RawName, strings.Join(argStrs, ","))), nil
}

// Get the proposal's state
func GetProposalState(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (types.ProtocolDaoProposalState, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return types.ProtocolDaoProposalState_Pending, err
	}
	value := new(uint8)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getState", big.NewInt(0).SetUint64(proposalId)); err != nil {
		return types.ProtocolDaoProposalState_Pending, fmt.Errorf("error getting state of proposal %d: %w", proposalId, err)
	}
	return types.ProtocolDaoProposalState(*value), nil
}

// Get the option that the address voted on for the proposal, and whether or not it's voted yet
func GetAddressVoteDirection(rp *rocketpool.RocketPool, proposalId uint64, address common.Address, opts *bind.CallOpts) (types.VoteDirection, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return types.VoteDirection_NoVote, err
	}
	value := new(uint8)
	if err := rocketDAOProtocolProposal.Call(opts, value, "getReceiptDirection", big.NewInt(0).SetUint64(proposalId), address); err != nil {
		return types.VoteDirection_NoVote, fmt.Errorf("error getting voting status of proposal %d by address %s: %w", proposalId, address.Hex(), err)
	}
	return types.VoteDirection(*value), nil
}

// ====================
// === Transactions ===
// ====================

// Estimate the gas of a proposal submission
func estimateProposalGas(rp *rocketpool.RocketPool, message string, payload []byte, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	err = simulateProposalExecution(rp, payload)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error simulating proposal execution: %w", err)
	}
	return rocketDAOProtocolProposal.GetTransactionGasInfo(opts, "propose", message, payload, blockNumber, treeNodes)
}

// Submit a trusted node DAO proposal
// Returns the ID of the new proposal
func submitProposal(rp *rocketpool.RocketPool, message string, payload []byte, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	proposalCount, err := dao.GetProposalCount(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposal.Transact(opts, "propose", message, payload, blockNumber, treeNodes)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error submitting Protocol DAO proposal: %w", err)
	}
	return proposalCount + 1, tx.Hash(), nil
}

// Estimate the gas of VoteOnProposal
func EstimateVoteOnProposalGas(rp *rocketpool.RocketPool, proposalId uint64, voteDirection types.VoteDirection, votingPower *big.Int, nodeIndex uint64, witness []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposal.GetTransactionGasInfo(opts, "vote", big.NewInt(int64(proposalId)), voteDirection, votingPower, big.NewInt(int64(nodeIndex)), witness)
}

// Vote on a submitted proposal
func VoteOnProposal(rp *rocketpool.RocketPool, proposalId uint64, voteDirection types.VoteDirection, votingPower *big.Int, nodeIndex uint64, witness []types.VotingTreeNode, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposal.Transact(opts, "vote", big.NewInt(int64(proposalId)), voteDirection, votingPower, big.NewInt(int64(nodeIndex)), witness)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error voting on Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of OverrideVote
func EstimateOverrideVoteGas(rp *rocketpool.RocketPool, proposalId uint64, voteDirection types.VoteDirection, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposal.GetTransactionGasInfo(opts, "overrideVote", big.NewInt(int64(proposalId)), voteDirection)
}

// Override a delegate's vote during pDAO voting phase 2
func OverrideVote(rp *rocketpool.RocketPool, proposalId uint64, voteDirection types.VoteDirection, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposal.Transact(opts, "overrideVote", big.NewInt(int64(proposalId)), voteDirection)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error overriding vote on Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Finalize
func EstimateFinalizeGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposal.GetTransactionGasInfo(opts, "finalise", big.NewInt(int64(proposalId)))
}

// Finalizes a vetoed proposal by burning the proposer's bond
func Finalize(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposal.Transact(opts, "finalise", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error finalizing Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ExecuteProposal
func EstimateExecuteProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposal.GetTransactionGasInfo(opts, "execute", big.NewInt(int64(proposalId)))
}

// Execute a submitted proposal
func ExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposal.Transact(opts, "execute", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error executing Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Simulate a proposal's execution to verify it won't revert
func simulateProposalExecution(rp *rocketpool.RocketPool, payload []byte) error {
	rocketDAOProtocolProposal, err := getRocketDAOProtocolProposal(rp, nil)
	if err != nil {
		return err
	}
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return err
	}

	_, err = rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     *rocketDAOProtocolProposal.Address,
		To:       rocketDAOProtocolProposals.Address,
		GasPrice: big.NewInt(0),
		Value:    nil,
		Data:     payload,
	})
	return err
}

// Get contracts
var rocketDAOProtocolProposalLock sync.Mutex

func getRocketDAOProtocolProposal(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolProposalLock.Lock()
	defer rocketDAOProtocolProposalLock.Unlock()
	return rp.GetContract("rocketDAOProtocolProposal", opts)
}
