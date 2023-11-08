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
	"github.com/ethereum/go-ethereum/common/math"
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

// Proposal details
type ProposalDetails struct {
	ID                   uint64                         `json:"id"`
	ProposerAddress      common.Address                 `json:"proposerAddress"`
	Message              string                         `json:"message"`
	StartBlock           uint64                         `json:"startBlock"`
	Phase1EndBlock       uint64                         `json:"phase1EndBlock"`
	Phase2EndBlock       uint64                         `json:"phase2EndBlock"`
	ExpiryBlock          uint64                         `json:"expiryBlock"`
	CreatedTime          time.Time                      `json:"createdTime"`
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
}

// Get a proposal's details
func GetProposalDetails(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (ProposalDetails, error) {
	var wg errgroup.Group
	var prop ProposalDetails

	// Load data
	wg.Go(func() error {
		var err error
		prop.ProposerAddress, err = GetProposalProposer(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Message, err = GetProposalMessage(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.StartBlock, err = GetProposalStartBlock(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Phase1EndBlock, err = GetProposalPhase1EndBlock(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.Phase2EndBlock, err = GetProposalPhase2EndBlock(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		prop.ExpiryBlock, err = GetProposalExpiryBlock(rp, proposalId, opts)
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

	// Wait for data
	if err := wg.Wait(); err != nil {
		return ProposalDetails{}, err
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
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getProposalBlock", proposalId); err != nil {
		return 0, fmt.Errorf("error getting proposal block for proposal %d: %w", proposalId, err)
	}
	return uint32((*value).Uint64()), nil
}

// Get the veto quorum required to veto a proposal
func GetProposalVetoQuorum(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getProposalVetoQuorum", proposalId); err != nil {
		return nil, fmt.Errorf("error getting proposal veto quorum for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// The total number of Protocol DAO proposals
func GetTotalProposalCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getTotal"); err != nil {
		return nil, fmt.Errorf("error getting total proposal count: %w", err)
	}
	return *value, nil
}

// Get the address of the proposer
func GetProposalProposer(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getProposer", proposalId); err != nil {
		return common.Address{}, fmt.Errorf("error getting proposer for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the proposal's message
func GetProposalMessage(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return "", err
	}
	value := new(string)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getMessage", proposalId); err != nil {
		return "", fmt.Errorf("error getting message for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the start block of this proposal
func GetProposalStartBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(uint64)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getStart", proposalId); err != nil {
		return 0, fmt.Errorf("error getting start block for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the phase 1 end block of this proposal
func GetProposalPhase1EndBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(uint64)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getPhase1End", proposalId); err != nil {
		return 0, fmt.Errorf("error getting phase 1 end block for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the phase 2 end block of this proposal
func GetProposalPhase2EndBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(uint64)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getPhase2End", proposalId); err != nil {
		return 0, fmt.Errorf("error getting phase 2 end block for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the block where the proposal expires and can no longer be executed if it is successful
func GetProposalExpiryBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, err
	}
	value := new(uint64)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getExpires", proposalId); err != nil {
		return 0, fmt.Errorf("error getting expiry block for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the time the proposal was created
func GetProposalCreationTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Time, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getCreated", proposalId); err != nil {
		return time.Time{}, fmt.Errorf("error getting creation time for proposal %d: %w", proposalId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the cumulative amount of voting power voting in favor of this proposal
func GetProposalVotingPowerFor(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVotesFor", proposalId); err != nil {
		return nil, fmt.Errorf("error getting total 'for' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power voting against this proposal
func GetProposalVotingPowerAgainst(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVotesAgainst", proposalId); err != nil {
		return nil, fmt.Errorf("error getting total 'against' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that vetoed this proposal
func GetProposalVotingPowerVetoed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVotesVeto", proposalId); err != nil {
		return nil, fmt.Errorf("error getting total 'veto' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that abstained from this proposal
func GetProposalVotingPowerAbstained(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVotesAbstained", proposalId); err != nil {
		return nil, fmt.Errorf("error getting total 'abstained' voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the cumulative amount of voting power that must vote on this proposal for it to be eligible for execution if it succeeds
func GetProposalVotingPowerRequired(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVotesRequired", proposalId); err != nil {
		return nil, fmt.Errorf("error getting required voting power for proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been destroyed
func GetProposalIsDestroyed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getDestroyed", proposalId); err != nil {
		return false, fmt.Errorf("error getting destroyed status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been finalized
func GetProposalIsFinalized(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getFinalised", proposalId); err != nil {
		return false, fmt.Errorf("error getting finalized status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal has been executed
func GetProposalIsExecuted(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getExecuted", proposalId); err != nil {
		return false, fmt.Errorf("error getting executed status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get whether or not the proposal's veto quorum has been met and it has been vetoed
func GetProposalIsVetoed(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getVetoed", proposalId); err != nil {
		return false, fmt.Errorf("error getting veto status of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the proposal's payload
func GetProposalPayload(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) ([]byte, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new([]byte)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getPayload", proposalId); err != nil {
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
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return types.ProtocolDaoProposalState_Pending, err
	}
	value := new(types.ProtocolDaoProposalState)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getState", proposalId); err != nil {
		return types.ProtocolDaoProposalState_Pending, fmt.Errorf("error getting state of proposal %d: %w", proposalId, err)
	}
	return *value, nil
}

// Get the option that the address voted on for the proposal, and whether or not it's voted yet
func GetAddressVoteDirection(rp *rocketpool.RocketPool, proposalId uint64, address common.Address, opts *bind.CallOpts) (types.VoteDirection, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return types.VoteDirection_NoVote, err
	}
	value := new(types.VoteDirection)
	if err := rocketDAOProtocolProposals.Call(opts, value, "getReceiptDirection", proposalId, address); err != nil {
		return types.VoteDirection_NoVote, fmt.Errorf("error getting voting status of proposal %d by address %s: %w", proposalId, address.Hex(), err)
	}
	return *value, nil
}

// Estimate the gas of a proposal submission
func EstimateProposalGas(rp *rocketpool.RocketPool, message string, payload []byte, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	err = simulateProposalExecution(rp, payload)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error simulating proposal execution: %w", err)
	}
	return rocketDAOProtocolProposals.GetTransactionGasInfo(opts, "propose", message, payload, blockNumber, treeNodes)
}

// Submit a trusted node DAO proposal
// Returns the ID of the new proposal
func SubmitProposal(rp *rocketpool.RocketPool, message string, payload []byte, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	proposalCount, err := dao.GetProposalCount(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposals.Transact(opts, "propose", message, payload, blockNumber, treeNodes)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error submitting Protocol DAO proposal: %w", err)
	}
	return proposalCount + 1, tx.Hash(), nil
}

// Estimate the gas of ProposeSetMulti
func EstimateProposeSetMultiGas(rp *rocketpool.RocketPool, message string, contractNames []string, settingPaths []string, settingTypes []types.ProposalSettingType, values []any, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	encodedValues, err := abiEncodeMultiValues(settingTypes, values)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error ABI encoding values: %w", err)
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingMulti", contractNames, settingPaths, settingTypes, encodedValues)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error setting multi-set proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update multiple Protocol DAO settings at once
func ProposeSetMulti(rp *rocketpool.RocketPool, message string, contractNames []string, settingPaths []string, settingTypes []types.ProposalSettingType, values []any, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	encodedValues, err := abiEncodeMultiValues(settingTypes, values)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error ABI encoding values: %w", err)
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingMulti", contractNames, settingPaths, settingTypes, encodedValues)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error setting multi-set proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetBool
func EstimateProposeSetBoolGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error setting bool setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a bool Protocol DAO setting
func ProposeSetBool(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error setting bool setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetUint
func EstimateProposeSetUintGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a uint Protocol DAO setting
func ProposeSetUint(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set uint setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetAddress
func EstimateProposeSetAddressGas(rp *rocketpool.RocketPool, message, contractName, settingPath string, value common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingAddress", contractName, settingPath, value)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set address setting proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update an address Protocol DAO setting
func ProposeSetAddress(rp *rocketpool.RocketPool, message, contractName, settingPath string, value common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingAddress", contractName, settingPath, value)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set address setting proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeSetRewardsPercentage
func EstimateProposeSetRewardsPercentageGas(rp *rocketpool.RocketPool, message string, odaoPercentage *big.Int, pdaoPercentage *big.Int, nodePercentage *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingRewardsClaimers", odaoPercentage, pdaoPercentage, nodePercentage)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set rewards-claimers percent proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update the allocations of RPL rewards
func ProposeSetRewardsPercentage(rp *rocketpool.RocketPool, message string, odaoPercentage *big.Int, pdaoPercentage *big.Int, nodePercentage *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSettingRewardsClaimers", odaoPercentage, pdaoPercentage, nodePercentage)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set rewards-claimers percent proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeOneTimeTreasurySpend
func EstimateProposeOneTimeTreasurySpendGas(rp *rocketpool.RocketPool, message, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryOneTimeSpend", invoiceID, recipient, amount)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding set spend-treasury percent proposal payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to spend a portion of the Rocket Pool treasury one time
func ProposeOneTimeTreasurySpend(rp *rocketpool.RocketPool, message, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryOneTimeSpend", invoiceID, recipient, amount)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding set spend-treasury percent proposal payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of VoteOnProposal
func EstimateVoteOnProposalGas(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposals.GetTransactionGasInfo(opts, "vote", big.NewInt(int64(proposalId)), support)
}

// Vote on a submitted proposal
func VoteOnProposal(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposals.Transact(opts, "vote", big.NewInt(int64(proposalId)), support)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error voting on Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ExecuteProposal
func EstimateExecuteProposalGas(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolProposals.GetTransactionGasInfo(opts, "execute", big.NewInt(int64(proposalId)))
}

// Execute a submitted proposal
func ExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolProposals.Transact(opts, "execute", big.NewInt(int64(proposalId)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error executing Protocol DAO proposal %d: %w", proposalId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ProposeRecurringTreasurySpend
func EstimateProposeRecurringTreasurySpendGas(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryNewContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(startTime.Unix()), numberOfPeriods)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalTreasuryNewContract payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to spend a portion of the Rocket Pool treasury in a recurring manner
func ProposeRecurringTreasurySpend(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryNewContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), big.NewInt(startTime.Unix()), numberOfPeriods)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalTreasuryNewContract payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeRecurringTreasurySpendUpdate
func EstimateProposeRecurringTreasurySpendUpdateGas(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryUpdateContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), numberOfPeriods)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalTreasuryUpdateContract payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to update a recurrint Rocket Pool treasury spending plan
func ProposeRecurringTreasurySpendUpdate(rp *rocketpool.RocketPool, message string, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalTreasuryUpdateContract", contractName, recipient, amountPerPeriod, big.NewInt(int64(periodLength.Seconds())), numberOfPeriods)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalTreasuryUpdateContract payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeInviteToSecurityCouncil
func EstimateProposeInviteToSecurityCouncilGas(rp *rocketpool.RocketPool, message string, id string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityInvite", id, address)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityInvite payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to invite a member to the security council
func ProposeInviteToSecurityCouncil(rp *rocketpool.RocketPool, message string, id string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityInvite", id, address)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityInvite payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Estimate the gas of ProposeKickFromSecurityCouncil
func EstimateProposeKickFromSecurityCouncilGas(rp *rocketpool.RocketPool, message string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKick", address)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error encoding proposalSecurityKick payload: %w", err)
	}
	return EstimateProposalGas(rp, message, payload, blockNumber, treeNodes, opts)
}

// Submit a proposal to kick a member from the security council
func ProposeKickFromSecurityCouncil(rp *rocketpool.RocketPool, message string, address common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return 0, common.Hash{}, err
	}
	payload, err := rocketDAOProtocolProposals.ABI.Pack("proposalSecurityKick", address)
	if err != nil {
		return 0, common.Hash{}, fmt.Errorf("error encoding proposalSecurityKick payload: %w", err)
	}
	return SubmitProposal(rp, message, payload, blockNumber, treeNodes, opts)
}

// Simulate a proposal's execution to verify it won't revert
func simulateProposalExecution(rp *rocketpool.RocketPool, payload []byte) error {
	rocketDAOProposal, err := getRocketDAOProposal(rp, nil)
	if err != nil {
		return err
	}
	rocketDAOProtocolProposals, err := getRocketDAOProtocolProposals(rp, nil)
	if err != nil {
		return err
	}

	_, err = rp.Client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     *rocketDAOProposal.Address,
		To:       rocketDAOProtocolProposals.Address,
		GasPrice: big.NewInt(0),
		Value:    nil,
		Data:     payload,
	})
	return err
}

// Get the ABI encoding of multiple values for a ProposeSettingMulti call
func abiEncodeMultiValues(settingTypes []types.ProposalSettingType, values []any) ([][]byte, error) {
	// Sanity check the lengths
	settingCount := len(settingTypes)
	if settingCount != len(values) {
		return nil, fmt.Errorf("settingTypes and values must be the same length")
	}
	if settingCount == 0 {
		return [][]byte{}, nil
	}

	// ABI encode each value
	results := make([][]byte, settingCount)
	for i, settingType := range settingTypes {
		var encodedArg []byte
		switch settingType {
		case types.ProposalSettingType_Uint256:
			arg, success := values[i].(*big.Int)
			if !success {
				return nil, fmt.Errorf("value %d is not a *big.Int, but the setting type is Uint256", i)
			}
			encodedArg = math.U256Bytes(big.NewInt(0).Set(arg))

		case types.ProposalSettingType_Bool:
			arg, success := values[i].(bool)
			if !success {
				return nil, fmt.Errorf("value %d is not a bool, but the setting type is Bool", i)
			}
			if arg {
				encodedArg = math.PaddedBigBytes(common.Big1, 32)
			} else {
				encodedArg = math.PaddedBigBytes(common.Big0, 32)
			}

		case types.ProposalSettingType_Address:
			arg, success := values[i].(common.Address)
			if !success {
				return nil, fmt.Errorf("value %d is not an address, but the setting type is Address", i)
			}
			encodedArg = common.LeftPadBytes(arg.Bytes(), 32)

		default:
			return nil, fmt.Errorf("unknown proposal setting type [%v]", settingType)
		}
		results[i] = encodedArg
	}

	return results, nil
}

// Get contracts
var rocketDAOProtocolProposalsLock sync.Mutex
var rocketDAOProposalLock sync.Mutex

func getRocketDAOProtocolProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolProposalsLock.Lock()
	defer rocketDAOProtocolProposalsLock.Unlock()
	return rp.GetContract("rocketDAOProtocolProposals", opts)
}

func getRocketDAOProposal(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProposalLock.Lock()
	defer rocketDAOProposalLock.Unlock()
	return rp.GetContract("rocketDAOProposal", opts)
}
