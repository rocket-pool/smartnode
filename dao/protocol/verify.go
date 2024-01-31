package protocol

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	challengeStateBatchSize uint64 = 500
)

// Structure of the RootSubmitted event
type RootSubmitted struct {
	ProposalID  *big.Int               `json:"proposalId"`
	Proposer    common.Address         `json:"proposer"`
	BlockNumber uint32                 `json:"blockNumber"`
	Index       *big.Int               `json:"index"`
	Root        types.VotingTreeNode   `json:"root"`
	TreeNodes   []types.VotingTreeNode `json:"treeNodes"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Internal struct - returned by the RootSubmitted event
type rootSubmittedRaw struct {
	BlockNumber uint32                 `json:"blockNumber"`
	Index       *big.Int               `json:"index"`
	Root        types.VotingTreeNode   `json:"root"`
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
	Index     *big.Int `json:"index"`
	Timestamp *big.Int `json:"timestamp"`
}

// Get the depth-per-round for voting trees
func GetDepthPerRound(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolVerifier.Call(opts, value, "getDepthPerRound"); err != nil {
		return 0, fmt.Errorf("error getting depth per round: %w", err)
	}
	return (*value).Uint64(), nil
}

// Get the node of a proposal at the given index
func GetNode(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.CallOpts) (types.VotingTreeNode, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return types.VotingTreeNode{}, err
	}
	// define a struct to unmarshall the VotingTreeNode data from the smart contract call
	res := new(struct {
		Sum  *big.Int `json:"sum"`
		Hash [32]byte `json:"hash"`
	})
	err = rocketDAOProtocolVerifier.Call(opts, &res, "getNode", big.NewInt(int64(proposalId)), big.NewInt(int64(index)))
	if err != nil {
		return types.VotingTreeNode{}, fmt.Errorf("error getting proposal %d / index %d node: %w", proposalId, index, err)
	}
	// convert the [32]byte field into a common.Hash
	node := types.VotingTreeNode{
		Sum:  res.Sum,
		Hash: common.BytesToHash(res.Hash[:]),
	}

	return node, nil
}

// Estimate the gas of CreateChallenge
func EstimateCreateChallengeGas(rp *rocketpool.RocketPool, proposalId uint64, index uint64, node types.VotingTreeNode, witness []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "createChallenge", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), node, witness)
}

// Challenge a proposal at a specific tree node index, providing a Merkle proof of the node as well
func CreateChallenge(rp *rocketpool.RocketPool, proposalId uint64, index uint64, node types.VotingTreeNode, witness []types.VotingTreeNode, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "createChallenge", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), node, witness)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error creating challenge: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SubmitRoot
func EstimateSubmitRootGas(rp *rocketpool.RocketPool, proposalId uint64, index uint64, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "submitRoot", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), treeNodes)
}

// Submit the Merkle root for a proposal at the specific index in response to a challenge
func SubmitRoot(rp *rocketpool.RocketPool, proposalId uint64, index uint64, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "submitRoot", big.NewInt(int64(proposalId)), big.NewInt((int64(index))), treeNodes)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting proposal root: %w", err)
	}
	return tx.Hash(), nil
}

// Get the state of a challenge on a proposal and tree node index
func GetChallengeState(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.CallOpts) (types.ChallengeState, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return types.ChallengeState_Unchallenged, err
	}
	state := new(uint8)
	if err := rocketDAOProtocolVerifier.Call(opts, state, "getChallengeState", big.NewInt(int64(proposalId)), big.NewInt(int64(index))); err != nil {
		return types.ChallengeState_Unchallenged, fmt.Errorf("error getting proposal %d / index %d challenge state: %w", proposalId, index, err)
	}
	challengeState := types.ChallengeState(*state)
	return challengeState, nil
}

// Get the defeat index for a proposal
func GetDefeatIndex(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolVerifier.Call(opts, value, "getDefeatIndex", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("error getting proposal %d defeat index: %w", proposalId, err)
	}
	return (*value).Uint64(), nil
}

// Get the proposal bond for a proposal
func GetProposalBond(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolVerifier.Call(opts, value, "getProposalBond", big.NewInt(int64(proposalId))); err != nil {
		return nil, fmt.Errorf("error getting proposal %d proposal bond: %w", proposalId, err)
	}
	return *value, nil
}

// Get the challenge bond for a proposal
func GetChallengeBond(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolVerifier.Call(opts, value, "getChallengeBond", big.NewInt(int64(proposalId))); err != nil {
		return nil, fmt.Errorf("error getting proposal %d challenge bond: %w", proposalId, err)
	}
	return *value, nil
}

// Get the challenge period for a proposal
func GetChallengePeriod(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (time.Duration, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := rocketDAOProtocolVerifier.Call(opts, value, "getChallengePeriod", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("error getting proposal %d challenge period: %w", proposalId, err)
	}
	return time.Second * time.Duration((*value).Uint64()), nil
}

// Get the states of multiple challenges using multicall
// NOTE: wen v2...
func GetMultiChallengeStatesFast(rp *rocketpool.RocketPool, multicallAddress common.Address, proposalIds []uint64, challengedIndices []uint64, opts *bind.CallOpts) ([]types.ChallengeState, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return nil, err
	}

	if opts == nil {
		// Get the latest block
		blockNum, err := rp.Client.BlockNumber(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error getting latest block number: %w", err)
		}
		opts = &bind.CallOpts{
			BlockNumber: big.NewInt(int64(blockNum)),
		}
	}

	count := uint64(len(proposalIds))
	if count != uint64(len(challengedIndices)) {
		return nil, fmt.Errorf("have %d proposal IDs but %d challenge indices", count, len(challengedIndices))
	}

	// Sync
	var wg errgroup.Group

	// Run the getters in batches
	rawStates := make([]uint8, count)
	for i := uint64(0); i < count; i += challengeStateBatchSize {
		i := i
		max := i + challengeStateBatchSize
		if max > count {
			max = count
		}

		// Load details
		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				propID := big.NewInt(int64(proposalIds[j]))
				challengedIndex := big.NewInt(int64(challengedIndices[j]))
				mc.AddCall(rocketDAOProtocolVerifier, &rawStates[j], "getChallengeState", propID, challengedIndex)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Cast the results
	states := make([]types.ChallengeState, count)
	for i, state := range rawStates {
		states[i] = types.ChallengeState(state)
	}
	return states, nil
}

// Get RootSubmitted event info
func GetRootSubmittedEvents(rp *rocketpool.RocketPool, proposalIDs []uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, previousVerifierAddresses []common.Address, opts *bind.CallOpts) ([]RootSubmitted, error) {
	// Get the contract
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return nil, err
	}

	// Construct a filter query for relevant logs
	idBuffers := make([]common.Hash, len(proposalIDs))
	for i, id := range proposalIDs {
		proposalIdBig := big.NewInt(0).SetUint64(id)
		proposalIdBig.FillBytes(idBuffers[i][:])
	}
	rootSubmittedEvent := rocketDAOProtocolVerifier.ABI.Events["RootSubmitted"]
	addressFilter := append(previousVerifierAddresses, *rocketDAOProtocolVerifier.Address)
	topicFilter := [][]common.Hash{{rootSubmittedEvent.ID}, idBuffers}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return []RootSubmitted{}, nil
	}

	events := make([]RootSubmitted, 0, len(logs))
	for _, log := range logs {
		// Get the log info values
		values, err := rootSubmittedEvent.Inputs.Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("error unpacking RootSubmitted event data: %w", err)
		}

		// Get the topic values
		if len(log.Topics) < 3 {
			return nil, fmt.Errorf("event had %d topics but at least 3 are required", len(log.Topics))
		}
		idHash := log.Topics[1]
		proposerHash := log.Topics[2]
		propID := big.NewInt(0).SetBytes(idHash.Bytes())
		proposer := common.BytesToAddress(proposerHash.Bytes())

		// Convert to a native struct
		var raw rootSubmittedRaw
		err = rootSubmittedEvent.Inputs.Copy(&raw, values)
		if err != nil {
			return nil, fmt.Errorf("error converting RootSubmitted event data to struct: %w", err)
		}

		// Get the decoded data
		events = append(events, RootSubmitted{
			ProposalID:  propID,
			Proposer:    proposer,
			BlockNumber: raw.BlockNumber,
			Index:       raw.Index,
			Root:        raw.Root,
			TreeNodes:   raw.TreeNodes,
			Timestamp:   time.Unix(raw.Timestamp.Int64(), 0),
		})
	}

	return events, nil
}

// Get ChallengeSubmitted event info
func GetChallengeSubmittedEvents(rp *rocketpool.RocketPool, proposalIDs []uint64, intervalSize *big.Int, startBlock *big.Int, endBlock *big.Int, opts *bind.CallOpts) ([]ChallengeSubmitted, error) {
	// Get the contract
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, opts)
	if err != nil {
		return nil, err
	}

	// Construct a filter query for relevant logs
	idBuffers := make([]common.Hash, len(proposalIDs))
	for i, id := range proposalIDs {
		proposalIdBig := big.NewInt(0).SetUint64(id)
		proposalIdBig.FillBytes(idBuffers[i][:])
	}
	challengeSubmittedEvent := rocketDAOProtocolVerifier.ABI.Events["ChallengeSubmitted"]
	addressFilter := []common.Address{*rocketDAOProtocolVerifier.Address}
	topicFilter := [][]common.Hash{{challengeSubmittedEvent.ID}, idBuffers}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, endBlock, nil)
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return []ChallengeSubmitted{}, nil
	}

	events := make([]ChallengeSubmitted, 0, len(logs))
	for _, log := range logs {
		// Get the log info values
		values, err := challengeSubmittedEvent.Inputs.Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("error unpacking ChallengeSubmitted event data: %w", err)
		}

		// Get the topic values
		if len(log.Topics) < 3 {
			return nil, fmt.Errorf("event had %d topics but at least 3 are required", len(log.Topics))
		}
		idHash := log.Topics[1]
		challengerHash := log.Topics[2]
		propID := big.NewInt(0).SetBytes(idHash.Bytes())
		challenger := common.BytesToAddress(challengerHash.Bytes())

		// Convert to a native struct
		var raw challengeSubmittedRaw
		err = challengeSubmittedEvent.Inputs.Copy(&raw, values)
		if err != nil {
			return nil, fmt.Errorf("error converting ChallengeSubmitted event data to struct: %w", err)
		}

		// Get the decoded data
		events = append(events, ChallengeSubmitted{
			ProposalID: propID,
			Challenger: challenger,
			Index:      raw.Index,
			Timestamp:  time.Unix(raw.Timestamp.Int64(), 0),
		})
	}

	return events, nil
}

// Estimate the gas of ClaimBondChallenger
func EstimateClaimBondChallengerGas(rp *rocketpool.RocketPool, proposalID uint64, indices []uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	// Make the args
	proposalIDBig := big.NewInt(int64(proposalID))
	indicesBig := make([]*big.Int, len(indices))
	for i, index := range indices {
		indicesBig[i] = big.NewInt(int64(index))
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "claimBondChallenger", proposalIDBig, indicesBig)
}

// Claim any RPL bond refunds or rewards for a proposal, as a challenger
func ClaimBondChallenger(rp *rocketpool.RocketPool, proposalID uint64, indices []uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	// Make the args
	proposalIDBig := big.NewInt(int64(proposalID))
	indicesBig := make([]*big.Int, len(indices))
	for i, index := range indices {
		indicesBig[i] = big.NewInt(int64(index))
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "claimBondChallenger", proposalIDBig, indicesBig)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

// Estimate the gas of ClaimBondProposer
func EstimateClaimBondProposerGas(rp *rocketpool.RocketPool, proposalID uint64, indices []uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	// Make the args
	proposalIDBig := big.NewInt(int64(proposalID))
	indicesBig := make([]*big.Int, len(indices))
	for i, index := range indices {
		indicesBig[i] = big.NewInt(int64(index))
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "claimBondProposer", proposalIDBig, indicesBig)
}

// Claim any RPL bond refunds or rewards for a proposal, as the proposer
func ClaimBondProposer(rp *rocketpool.RocketPool, proposalID uint64, indices []uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	// Make the args
	proposalIDBig := big.NewInt(int64(proposalID))
	indicesBig := make([]*big.Int, len(indices))
	for i, index := range indices {
		indicesBig[i] = big.NewInt(int64(index))
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "claimBondProposer", proposalIDBig, indicesBig)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

// Estimate the gas of DefeatProposal
func EstimateDefeatProposalGas(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOProtocolVerifier.GetTransactionGasInfo(opts, "defeatProposal", big.NewInt(int64(proposalId)), big.NewInt(int64(index)))
}

// Defeat a proposal if it fails to respond to a challenge within the challenge window, providing the node index that wasn't responded to
func DefeatProposal(rp *rocketpool.RocketPool, proposalId uint64, index uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOProtocolVerifier, err := getRocketDAOProtocolVerifier(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOProtocolVerifier.Transact(opts, "defeatProposal", big.NewInt(int64(proposalId)), big.NewInt(int64(index)))
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDAOProtocolVerifierLock sync.Mutex

func getRocketDAOProtocolVerifier(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolVerifierLock.Lock()
	defer rocketDAOProtocolVerifierLock.Unlock()
	return rp.GetContract("rocketDAOProtocolVerifier", opts)
}
