package voting

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Structure of the RootSubmitted event
type RootSubmitted struct {
	ProposalID  *big.Int               `json:"proposalId"`
	Proposer    common.Address         `json:"proposer"`
	BlockNumber uint32                 `json:"blockNumber"`
	Index       *big.Int               `json:"index"`
	RootHash    common.Hash            `json:"rootHash"`
	Sum         *big.Int               `json:"sum"`
	TreeNodes   []types.VotingTreeNode `json:"treeNodes"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Internal struct - returned by the RootSubmitted event
type rootSubmittedRaw struct {
	ProposalID  *big.Int               `json:"proposalId"`
	Proposer    common.Address         `json:"proposer"`
	BlockNumber uint32                 `json:"blockNumber"`
	Index       *big.Int               `json:"index"`
	RootHash    common.Hash            `json:"rootHash"`
	Sum         *big.Int               `json:"sum"`
	TreeNodes   []types.VotingTreeNode `json:"treeNodes"`
	Timestamp   *big.Int               `json:"timestamp"`
}

// Structure of the ChallengeSubmitted event
type ChallengeSubmitted struct {
	ProposalID *big.Int       `json:"proposalId"`
	Challenger common.Address `json:"challenger"`
	Index      *big.Int       `json:"index"`
	Timestamp  time.Time      `json:"timestamp"`
}

// Internal struct - returned by the ChallengeSubmitted event
type challengeSubmittedRaw struct {
	ProposalID *big.Int       `json:"proposalId"`
	Challenger common.Address `json:"challenger"`
	Index      *big.Int       `json:"index"`
	Timestamp  *big.Int       `json:"timestamp"`
}

// Estimate the gas of CreateChallenge
func EstimateCreateChallengeGas(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "createChallenge", big.NewInt(int64(proposalId)), big.NewInt((int64(index))))
}

// Submit the Merkle root for a proposal at the specific index in response to a challenge
func CreateChallenge(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "createChallenge", big.NewInt(int64(proposalId)), big.NewInt((int64(index))))
	if err != nil {
		return common.Hash{}, fmt.Errorf("error creating challenge: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SubmitRoot
func EstimateSubmitRootGas(rp *rocketpool.RocketPool, proposalId uint64, index uint64, witness []types.VotingTreeNode, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "submitRoot", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), witness, treeNodes)
}

// Submit the Merkle root for a proposal at the specific index in response to a challenge
func SubmitRoot(rp *rocketpool.RocketPool, proposalId uint64, index uint64, witness []types.VotingTreeNode, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "submitRoot", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), witness, treeNodes)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting proposal root: %w", err)
	}
	return tx.Hash(), nil
}

// Get RootSubmitted event info
func GetRootSubmittedEvent(rp *rocketpool.RocketPool, proposalID uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, opts *bind.CallOpts) (bool, RootSubmitted, error) {
	// Get the contract
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return false, RootSubmitted{}, err
	}

	// Construct a filter query for relevant logs
	proposalIdBig := big.NewInt(0).SetUint64(proposalID)
	rootSubmittedEvent := rocketDAOProtocolVerifier.ABI.Events["RootSubmitted"]
	proposalIdBytes := [32]byte{}
	proposalIdBig.FillBytes(proposalIdBytes[:])
	addressFilter := []common.Address{*rocketDAOProtocolVerifier.Address}
	topicFilter := [][]common.Hash{{rootSubmittedEvent.ID}, {proposalIdBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return false, RootSubmitted{}, err
	}
	if len(logs) == 0 {
		return false, RootSubmitted{}, nil
	}

	// Get the log info values
	values, err := rootSubmittedEvent.Inputs.Unpack(logs[0].Data)
	if err != nil {
		return false, RootSubmitted{}, fmt.Errorf("error unpacking RootSubmitted event data: %w", err)
	}

	// Convert to a native struct
	var raw rootSubmittedRaw
	err = rootSubmittedEvent.Inputs.Copy(&raw, values)
	if err != nil {
		return false, RootSubmitted{}, fmt.Errorf("error converting RootSubmitted event data to struct: %w", err)
	}

	// Get the decoded data
	eventData := RootSubmitted{
		ProposalID:  raw.ProposalID,
		Proposer:    raw.Proposer,
		BlockNumber: raw.BlockNumber,
		Index:       raw.Index,
		RootHash:    raw.RootHash,
		Sum:         raw.Sum,
		TreeNodes:   raw.TreeNodes,
		Timestamp:   time.Unix(raw.Timestamp.Int64(), 0),
	}

	return true, eventData, nil
}

// Get ChallengeSubmitted event info
func GetChallengeSubmittedEvent(rp *rocketpool.RocketPool, proposalID uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, opts *bind.CallOpts) (bool, ChallengeSubmitted, error) {
	// Get the contract
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return false, ChallengeSubmitted{}, err
	}

	// Construct a filter query for relevant logs
	proposalIdBig := big.NewInt(0).SetUint64(proposalID)
	challengeSubmittedEvent := rocketDAOProtocolVerifier.ABI.Events["ChallengeSubmitted"]
	proposalIdBytes := [32]byte{}
	proposalIdBig.FillBytes(proposalIdBytes[:])
	addressFilter := []common.Address{*rocketDAOProtocolVerifier.Address}
	topicFilter := [][]common.Hash{{challengeSubmittedEvent.ID}, {proposalIdBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return false, ChallengeSubmitted{}, err
	}
	if len(logs) == 0 {
		return false, ChallengeSubmitted{}, nil
	}

	// Get the log info values
	values, err := challengeSubmittedEvent.Inputs.Unpack(logs[0].Data)
	if err != nil {
		return false, ChallengeSubmitted{}, fmt.Errorf("error unpacking ChallengeSubmitted event data: %w", err)
	}

	// Convert to a native struct
	var raw challengeSubmittedRaw
	err = challengeSubmittedEvent.Inputs.Copy(&raw, values)
	if err != nil {
		return false, ChallengeSubmitted{}, fmt.Errorf("error converting ChallengeSubmitted event data to struct: %w", err)
	}

	// Get the decoded data
	eventData := ChallengeSubmitted{
		ProposalID: raw.ProposalID,
		Challenger: raw.Challenger,
		Index:      raw.Index,
		Timestamp:  time.Unix(raw.Timestamp.Int64(), 0),
	}

	return true, eventData, nil
}

// Get contracts
var rocketDAOProtocolVerifierLock sync.Mutex

func getRocketDAOProtocolVerifier(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolVerifierLock.Lock()
	defer rocketDAOProtocolVerifierLock.Unlock()
	return rp.GetContract("rocketDAOProtocolVerifier", opts)
}
